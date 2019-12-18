# Generic reconciler

This doc aims to capture some thoughts around creating a generic reconciler that can be applied to multiple, generated CRD structs to help reduce the overhead in adding new services to the operator.

The high-level idea is to write a [generic reconciler](https://github.com/Azure/k8s-infra/issues/7) that can work with structs generated from swagger that represent resources in Azure. Along with the reconciler there would be a validating admission webhook that would validate incoming changes to give immediate feedback for a better user experience (a simple example being rejecting changes to read-only property).

The goal would be to make the generic handling as widely applicable as possible, but to accommodate variation across APIs it is anticipated that there would be a need to enable custom behaviours. One way to allow this would be to expose hooks at various points during the processing cycle where custom logic can be registered for a type.

## Generating the structs

### Understanding the relationship between resources

In `azbrowse` we don't currently have a requirement to process the swagger body content, however the swagger ingestion process does take the flat data structures and apply a hierarchical grouping based on the URLs. Due to the conventions in ARM, this hierarchy puts the top-level entities nearer the root of the tree. This hierarchy might prove useful in mapping out CRDs.

### Spec vs status

In Kubernetes, objects have `spec` and `status` properties which is generally a useful split for managing objects. Additionally, looking at Terraform, properties can fall into a number of categories: writeable (can be freely updated), readonly (computed values that can't be written but are useful to be able to retrieve), and create-only (can be specified when the resource is created but are immutable thereafter).

There a likely a number of heuristic that can be pulled out from the RP's swagger specs to inform the autogen code and allow is to make distinctions between `spec` and `status`.

#### Option 1: Diff between `METHOD` and accepted fields

Based on our current knowledge of the ARM API definitions, these distinctions aren't clearly marked. However, mutable resources in ARM tend to have endpoints that support `CREATE`, `POST`,`PUT`, `GET` (and `DELETE`). Here it is assumed that `POST` is used to create a new resource in a collection and `PUT` is used to update. By comparing the properties for the different methods, it might be possible to assemble this knowledge:

* Use the diff between the `GET` properties and (union of `POST` and `PUT` properies) we can find the readonly properties. I.e. those in `GET` but not in `POST` (create) or `PUT` (update) are readonly
* Use the diff between `POST` (create) properties and `PUT` (update) properties to find the properties that are settable when creating a resource but readonly after that point

### Option 2: Swagger annotations/descriptions

Path parameters in the `swagger` are, based on cursory review of a few RPs, a good indicator for `create-only` properties. For example, for `/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}`:

```json
{
  "name": "resourceGroupName",
  "in": "path",
  "required": true,
  "type": "string",
  "description": "The name of the resource group to create or update. Can include alphanumeric, underscore, parentheses, hyphen, period (except at end), and Unicode charactethat matthe allowed characters.",
  "pattern": "^[-\\w\\._\\(\\)]+$",
  "minLength": 1,
  "maxLength": 90
},
```

> Note: In this case a `regex` is provided which can also be used by the validation webhook. The Azure Golang SDK uses these extensively for pre-submission validation.

This can also be combined with `readOnly: true` annotations in the `body` definitions of the swagger to judge which properties should go into the `status` section of the `CRD`. For example `resourceGroup` has the following [(sections remove for easy reading)](https://github.com/Azure/azure-rest-api-specs/blob/8170b1d/specification/resources/resource-manager/Microsoft.Resources/stable/2019-05-01/resources.json#L3329):

```json
{
    ....
    {
    "id": {
      "readOnly": true,
      "type": "string",
      "description": "The ID of the resource group."
    },
    "name": {
      "readOnly": true,
      "type": "string",
      "description": "The name of the resource group."
    },
    ......
    "location": {
      "type": "string",
      "description": "The location of the resource group. It cannot be changed after the resource group has been created. It must be one of the supported Azure locations."
    },
    ......
    "tags": {
      "type": "object",
      "additionalProperties": {
        "type": "string",
        "description": "The additional properties. "
      },
      "description": "The tags attached to the resource group."
    }
  },
  "required": [
    "location"
  ],
},
```

Worth noting is the inconsistency in the spec here between `name` being `readOnly` and `location` not being `readOnly` but instead having a description stating `It cannot be changed after the resource group has been created`.

While not perfect the `description` field could also be reviewed to inform based on matching for certain strings, however, how viable that is across different RPs is currently unknown.

### Exposing this information on generated CRD structs

The `autogen` code would read the `swagger` spec and generate the CRD struct based off the information extracted adding `tags` and kubebuilder annotations to store the additional information on the struct which can then be used by the generic reconciler and validator.

Here is a rough example for `ResourceGroup` -> [Run here on goplayground](https://play.golang.org/p/v9wufOCfWJu)

```golang
package main

import (
 "fmt"
 "reflect"
)

func main() {
  type ResourceGroup struct {
     // kubebuilder validation annotations and description comment come from the swagger

     //+kubebuilder:validation:Required
     //+kubebuilder:validation:MinLength=1
     //+kubebuilder:validation:MaxLength=90
     //+kubebuilder:validation:Pattern=^[-\\w\\._\\(\\)]+$
     // The name of the resource group.
     Name  string `validation-type:"create-only"`
     Location string `validation-type:"create-only"`
     ID string `validation-type:"read-only"`
  }

  u := ResourceGroup{"RG1", "WestEurope", "/some/resource/id"}
  t := reflect.TypeOf(u)

  for _, fieldName := range []string{"Name", "Location", "ID"} {
    field, found := t.FieldByName(fieldName)
    if !found {
      continue
    }
    fmt.Printf("\nField: User.%s\n", fieldName)
    fmt.Printf("\tWhole tag value : %q\n", field.Tag)
    fmt.Printf("\tValue of 'validation-type': %q\n", field.Tag.Get("validation-type"))
  }
}
```

## Reconciler loop

The reconciler loop will fire when the object specs are changed, but also need to fire periodically to ensure that the object status is an accurate representation of the current resource state in ARM. (As an aside, it would be interesting to look at whether Event Grid could be used for push-based notification of status changes rather than polling for changes as a way to reduce overhead with a large number of managed objects).

### Updating Spec

Consider a `virtualMachineScaleSet` pseudo-object and suppose you created a VMSS using the following

```yaml
group: ...
kind: virtualMachineScaleSet
metadata:
  name: my-vmss
spec:
  sku:
    name: Standard_D2_v3
    capacity: 3
  tags:
    tag1: value
```

Here `capacity` indicates the number of VM instances in the scale set, so is conceptually similar to the `replicas` property for a Kubernetes `Deployment`. Whist this field is in the `spec`, it is possible that it is changed by something external to the operator (e.g. VMSS auto-scale rule). In this case, the reconcile loop should update the property to reflect the new value:

```yaml
group: ...
kind: virtualMachineScaleSet
metadata:
  name: my-vmss
spec:
  sku:
    name: Standard_D2_v3
    capacity: 5
  tags:
    tag1: value
```

### PATCH semantics

The line of thought from the previous section becomes more interesting when we consider changes applied to the CRD as well as updates that occur to the resource in ARM. When trying to work out what the desired behaviour would be for the following scenario, we considered an analogous scenario with a `Deployment` instead of `virtualMachineScaleSet` (and swapped `replicas` for `capacity`).

Again, start with creating the initial VMSS instance:

```yaml
group: ...
kind: virtualMachineScaleSet
metadata:
  name: my-vmss
spec:
  sku:
    name: Standard_D2_v3
    capacity: 3
  tags:
    tag1: value
```

After this, suppose that you want to issue a `PATCH` on the CRD to update a tag:

```yaml
group: ...
kind: virtualMachineScaleSet
metadata:
  name: my-vmss
spec:
  tags:
    tag1: new-value
```

If this update also coincides with the capacity being updated in ARM then we have a merge conflict: updating the CRD from ARM would lose the tag change and updating ARM from the CRD would undo the autoscaler's change.

To resolve this, we propose adding an annotation to the CRD to store the last written values. To illustrate this, the CRD state after the initial create would be:

```yaml
group: ...
kind: virtualMachineScaleSet
metadata:
  name: my-vmss
  k8s-infra-last-write: "{sku:{'name':'Standard_D2_v3',capacity:3},tags:{'tag1':'value'}}"
spec:
  sku:
    name: Standard_D2_v3
    capacity: 3
  tags:
    tag1: value
```

For ARM resources that support `PATCH` the reconcile loop can use the annotation to perform the following:

1. Diff the current CRD spec with the latest `last-write` value
2. Apply the changes in the diff
   1. issue a `PATCH` request to ARM to apply the diff changes to the resource
   2. apply the diff changes to the `last-write` annotation
3. Perform a `GET` against the ARM resource and use this to update the spec

For ARM resources that don't support `PATCH` the approach would be slightly more complicated:

1. Diff the current CRD spec with the latest `last-write` value
2. Apply the changes in the diff
   1. perform a `GET` against the ARM resource to get the latest state
   2. apply the changes from the diff to the resource state and issue a `PUT` request to ARM to apply the changes to the resource
   3. apply the diff changes to the `last-write` annotation
3. Perform a `GET` against the ARM resource and use this to update the spec

Ideally, the non-`PATCH` case would use etags for the update but as far as we know, that isn't supported.

A consideration here is that the behavior of this process is complex and across many RP's or based on user preference different merge behavior would be required:

* User A may want to ALWAYS enforce conformance to the `SPEC` of the CRD BUT only for fields they have set, allowing drift in other fields which are not set.
* User B may want to ALWAYS enforce conformance to the `SPEC` of the CRD for ALL fields, overwriting fields set outside the operator or removing additional fields set outside the operator which aren't specified on the CRD.
* User C may want the CRD to REFLECT changes made outside of the operator, such as the example above where an autoscaling rule in the RP is changing capacity.

Picking a few supported approaches and allowing user choice in this behavior feels like a sensible option. This approach is inspired by [terraforms usage of `ignore_changes`](https://www.terraform.io/docs/configuration/resources.html#ignore_changes).

## Managing objects that were initially created outside of k8s-infra

The thoughts so far in this proposal are centred around resources that are created via the operator. To enable adoption of this on non-greenfield projects it might be helpful to allow existing resources to become managed by the operator (similar to `terraform import`). One thought around this was to include a `resourceID` property on the CRDs.

If the resource is created via the operator then this would be populated with the resource ID from ARM by the operator in the reconcile loop:

```yaml
group: ...
kind: virtualMachineScaleSet
metadata:
  name: my-vmss
spec:
  resourceID: /subscriptions/00000000-0000-0000-0000-00000000/mygroup/providers/Microsoft.Compute/virtualMachineScaleSets/my-vmss
  sku:
    name: Standard_D2_v3
    capacity: 3
  tags:
    tag1: value
```

To 'import' a resource that already exists so that the operator can manage it, only the `resourceID` would be specified as shown below. Once the reconcile loop has run the spec would be updated to resemble the definition above.

```yaml
group: ...
kind: virtualMachineScaleSet
metadata:
  name: my-vmss
spec:
  resourceID: /subscriptions/00000000-0000-0000-0000-00000000/mygroup/providers/Microsoft.Compute/virtualMachineScaleSets/my-vmss
```

NOTE: this likely complicates the validation logic (e.g. clashing with fields that are required when creating a resource)

## TODO

This section has some high-level notes that haven't yet been fleshed out into the proposal text

* custom metadata in swagger (i.e. accessing `x-ms-*` properties for richer metadata) - is this possible via the standard swagger libraries?
* joining data acrross multiple calls (e.g. VMSS status info from /instanceView)

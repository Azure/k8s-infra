# Generic reconciler

This doc aims to capture some thoughts around creating a generic reconciler that can be applied to multiple, generated CRD structs to help reduce the overhead in adding new services to the operator.

The high-level idea is to write a generic reconciler that can work with structs generated from swagger that represent resources in Azure. Along with the reconciler there would be a validating admission webhook that would validate incoming changes to give immediate feedback for a better user experience (a simple example being rejecting changes to read-only property).

The goal would be to make the generic handling as widely applicable as possible, but to accommodate variation across APIs it is anticipated that there would be a need to enable custom behaviours. One way to allow this would be to expose hooks at various points during the processing cycle where hooks can be registered.

## Generating the structs

### Understanding the relationship between resources

In `azbrowse` we don't currently have a requirement to process the swagger body content, however the swagger ingestion process does take the flat data structures and apply a hierarchical grouping based on the URLS. Due to the conventions in ARM, this hierarchy puts the top-level entities nearer the root of the tree. This hierarchy might prove useful in mapping out CRDs.

### Spec vs status

In Kubernetes, objects have `spec` and `status` properties which is generally a useful split for managing objects. Additionally, looking at Terraform, properties can fall into a number of categories: writeable (can be freely updated), readonly (computed values that can't be read but are useful to be able to retrieve), and create-only (can be specified when the resource is created but are immutable thereafter).

Based on our current knowledge of the ARM API definitions, these distinctions aren't clearly marked. However, mutable resource in ARM tend to have endpoints that support `CREATE`, `POST`,`PUT`, `GET` (and `DELETE`). Here is is assumed that `POST` is used to create a new resource in a collection and `PUT` is used to update. By comparing the properties for the different methods, it might be possible to assemble this knowledge:

* Use the diff between the `GET` properties and (union of `POST` and `PUT` properies) we can find the readonly properties. I.e. those in `GET` but not in `POST` (create) or `PUT` (update) are readonly
* Use the diff between `POST` (create) properties and `PUT` (update) properties to find the properties that are settable when creating a resource but readonly after that point

## Reconciler loop

The reconciler loop will fire when the object specs are change, but also need to fire periodically to ensure that the object status is an accurate representation of the current resource state in ARM. (As an aside, it would be interesting to look at whether Event Grid could be used for push-based notification of status changes rather than polling for changes as a way to reduce overhead with a large number of managed objects).

### Updating Spec

Consider a `virtualMachineScaleSet` pseudo-object and suppose you created a VMSS using the following

```yaml
group: ...
kind: virtualMachineScaleSet
spec:
  sku:
    name: Standard_D2_v3
    capacity: 3
  tags:
    tag1: value
```

Here `capactiy` indicates the number of VM instances in the scale set, so is conceptually similar to the `replicas` property for a Kubernetes `Deployment`.

TODO - draw parallel with replicas

### PATCH semantics

TODO what happens if making a PATCH after a spec value has changed.

## TODO

This section has some high-level notes that should be fleshed out into the proposal text

* Generic validator driven off generated structs
  * enforces create-only and immutable fields not being changed
  * allow pluggable validation hooks for custom logic if needed
* Generic reconciler driven off generated structs
  * generic loop for CRD reconciliation
  * allow pluggable mappers/mutators as needed
  * Process - need to give background on motivation (lack of etags, lack of consistent ATCH etc)
    * Store `last-applied` state in CRD (annotation?)
    * when reconciling, diff `spec` and `last-applied` to generate a patch metadata
    * if ARM API supports `PATCH` then use then generate a patch with the metadata
    * if ARM API doesn't support patch then `GET` and apply patch metadata to the response, hen use this to issue a `PUT`
    * Apply changes in patch metadata to `last-applied`
    * Issue a `GET` and use this to update latest spec/status
    * Add an example here for VMSS (pseudo-CRD with flow of applying updates)
* Enable `terraform import` scenario. Add a ResourceID property that is populated with the ID of resource created via the operator, but also consider allowing that to be set when an object is created to attach to an existing resource and allow it to be managed via the operator
* custom metadata in swagger (i.e. accessing `x-ms-*` properties for richer metadata) - is this possible via the standard swagger libraries?
* joining data acrross multiple calls (e.g. VMSS status info from /instanceView)
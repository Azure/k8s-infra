# Type references and ownership

## Related reading
- [Kubernetes garbage collection](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/).

## Terminology
- ASO: [Azure Service Operator](https://github.com/Azure/azure-service-operator)
- k8s-infra: The handcrafted precurser to the code generation tool being designed.

## Goals
- Provide a way for customers to express relationships between Azure resources in Kubernetes in an idiomatic way.
- Provide automatic ownership and garbage collection (using [Kubernetes garbage collection](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/)) 
  where appropriate (e.g. ResourceGroup as an owner of all the resources inside of it)
  - Ideally ResourceGroup is handled the same as other owners and isn't special cased. 
- Define how Kubernetes interacts with Azure resources _not_ created/managed by Kubernetes, for example resources which were created prior to the 
  customer onboarding to the Azure Service Operator.
- References should be extensible to work across multiple Azure subscriptions, although
  initially we may not support that.
 
## Non-goals
- Managing ownership for resources/resource hierarchies that were not created
  by the service operator.
  While this proposal allows references to point to external resources not managed by the service operator,
  the operator is not watching/monitoring the resource in question
  and as such cannot propagate deletes. Put another way: for the operator to manage ownership/object lifecycles, the entire resource hierarchy must exist _within Kubernetes_. If only part of the resource hierarchy
  is managed by the service operator, only those parts can have their lifecycles managed.  

## Different kinds of resource relationships in Azure
#### Related/linked resources 
Two resources are related to one another ("has-a" relationship), but there is no ownership.
Example: [VMSS -> Subnet](https://github.com/Azure/k8s-infra/blob/master/apis/microsoft.compute/v1/virtualmachinescaleset_types.go#L169) ([json schema](https://schema.management.azure.com/schemas/2019-07-01/Microsoft.Compute.json)).

This relationship is stereotypically one-way (a VMSS refers to a Subnet, but a Subnet _does not_ refer to a VMSS).

#### Owner and dependent
Two resources have a relationship where one is owned by the other.

Examples:
- a [RouteTable owns many Routes](https://github.com/Azure/k8s-infra/blob/master/apis/microsoft.network/v1/routetable_types.go#L18) ([json schema](https://schema.management.azure.com/schemas/2020-03-01/Microsoft.Network.json))
- a BatchAccount owns many Pools ([json schema](https://schema.management.azure.com/schemas/2017-09-01/Microsoft.Batch.json))
- a ResourceGroup owns any resource

A relationship like those shown here tells us two things:
- Where to create/manage the dependent resource (this `Route` goes in that particular `RouteTable`, this `RouteTable` has that `Route`)
- That the dependent resource should be deleted when the parent resource is deleted. TODO: Race conditions here
    - There are theoretically two cases here:
        - The dependent resource must be deleted before the parent can be deleted.
        - Deletion of the parent automatically cascades to all dependent resources. To the best of my knowledge all
         owning resource relationships in Azure are this kind.

Note that sometimes an owning resource has its dependent resources embedded directly
(for example: `RouteTable` has the property [RouteTablePropertiesFormat](https://schema.management.azure.com/schemas/2020-03-01/Microsoft.Network.json)).
Most types do not embed the dependent resource directly in the owning resource. We need to cater to both scenarios.

## What do these relationships look like in existing solutions?

This section examines how other operator solutions have tackled these problems. We look at:
- ARM templates
- Azure Service Operator (ASO)
- k8s-infra

### Related/Linked resources

#### What ARM does
These are just properties (often but not always called `id`) which refer to the fully qualified ARM ID of another resource. For example see a
[sample deployment template for a VMSS refering to an existing vnet](https://github.com/Azure/azure-quickstart-templates/blob/master/201-vmss-existing-vnet/azuredeploy.json#L136).

```json
"properties": {
  "subnet": {
    "id": "[resourceId(parameters('existingVnetResourceGroupName'), 'Microsoft.Network/virtualNetworks/subnets', parameters('existingVnetName'), parameters('existingSubNetName'))]"
  },
}
```

#### What ASO does
Similar to how ARM templates behave, ASO uses the decomposition of fully qualified resource id to reference another resource, as seen [here for VMSS -> VNet](https://github.com/Azure/azure-service-operator/blob/92240406aff3863f3a267d8a1dc1e28aa3e841ae/api/v1alpha1/azurevmscaleset_types.go#L25)

```go
type AzureVMScaleSetSpec struct {
    ...
	Location               string `json:"location"`
	ResourceGroup          string `json:"resourceGroup"`
	VirtualNetworkName     string `json:"virtualNetworkName"`
	SubnetName             string `json:"subnetName"`
    ...
}
```

These properties are combined into a fully qualified ARM ID like so:

```go
subnetIDInput := helpers.MakeResourceID(
    client.SubscriptionID,
    resourceGroupName,
    "Microsoft.Network",
    "virtualNetworks",
    vnetName,
    "subnets",
    subnetName,
)
```

Currently ASO does not support cross-subscription references (and some of the resources such as VMSS don't allow cross-resource group references), but it in theory could by adding parameters.

#### What k8s-infra does
k8s-infra is a bit different in that resource references are in Kubernetes style (namespace + name) and not Azure style (resource-group + resource-name).
All resource references are done using the special type `KnownTypeReference` which contains the fully qualified Kubernetes name for the resource.

### Dependent Resources
#### What ARM does
ARM template deployments support [two different ways of deploying dependent resources](https://docs.microsoft.com/en-us/azure/azure-resource-manager/templates/child-resource-name-type):

- Embed the dependent resource directly in the same ARM template in which the owning resource is created, using the `resources` property.
    - This requires that the owning resource and the dependent resource are created at the same time.
- Deploy the dependent resource separately.
    - When doing this, you should specify `dependsOn` stating that your new dependent resource depends on its owner. This is not strictly required in cases where
      the owner already exists, but is required when it is still being created (possibly in parallel).
    - The `name` field of the dependent resource in this mode serves a dual purpose - to name the dependent resource and to identify the parent resources.
      Each segment of the name corresponds to an owning resource, so creating a Batch Pool `foo` in Batch account `account` would have `name` = `account/foo`.

#### What ASO does
Dependent resources in ASO have properties which map to the name/path to their owner. For example 
[MySQLFirewallRuleSpec](https://github.com/Azure/azure-service-operator/blob/92240406aff3863f3a267d8a1dc1e28aa3e841ae/api/v1alpha1/mysqlfirewallrule_types.go#L15) looks like this:

```go
type MySQLFirewallRuleSpec struct {
	ResourceGroup  string `json:"resourceGroup"`
	Server         string `json:"server"`
	StartIPAddress string `json:"startIpAddress"`
	EndIPAddress   string `json:"endIpAddress"`
}
```

The `ResourceGroup` and `Server` are references to the owners of this type.

#### What k8s-infra does
k8s-infra uses the same `KnownTypeReference` type mentioned above for ownership references too.
There are two patterns for ownership in k8s-infra today.

One pattern is used for ResourceGroup, where 
[top level resources have a link to the resource group they are in](https://github.com/Azure/k8s-infra/blob/14105b1cb3f6967cd086c9f8f75fb16bb85d6318/apis/microsoft.compute/v1/virtualmachinescaleset_types.go#L323).

```go
type VirtualMachineScaleSetSpec struct {
    // +kubebuilder:validation:Required
    ResourceGroupRef *azcorev1.KnownTypeReference `json:"resourceGroupRef" group:"microsoft.resources.infra.azure.com" kind:"ResourceGroup"`
    ...
}
```

The other pattern is where the owning resource has links to the dependent resources it is expecting to have:

```go
type RouteTableSpecProperties struct {
    DisableBGPRoutePropagation bool                          `json:"disableBgpRoutePropagation,omitempty"`
    RouteRefs                  []azcorev1.KnownTypeReference `json:"routeRefs,omitempty" group:"microsoft.network.infra.azure.com" kind:"Route" owned:"true"`
}
```

If the dependent resources aren't there, the status of the owning resource reflects an error explaining that.

## Proposal

### How to represent references
We construct a new type which is our reference holder. I am calling this approach `AugmentedKnownTypeReference` (real type name TBD), because
it's similar to the current approach that `k8s-infra` is taking but with a few changes.

```go
type AugmentedKnownTypeReference struct {
    Kubernetes KubernetesReference
    Azure      AzureReference
}

type KubernetesReference struct {
    // The "Group" of the owning resources GVK. There may be an annotation on the property 
    // that specifies Group, in which case this is optional (if supplied it must match the group
    // on the annotation). For resources which do not have a group annotation, this is required.
    // TODO: It's possibly going to be confusing to customers when they have to specify this and when they don't...
    Group string

    // The "Kind" of the owning resources GVK. There may be an annotation on the property 
    // that specifies Kind, in which case this is optional (if supplied it must match the kind
    // on the annotation). For resources which do not have a kind annotation, this is required.
    // TODO: It's possibly going to be confusing to customers when they have to specify this and when they don't... 
    Kind string
    
    // Note that no Version is required here because the objective is to link to a single Kubernetes
    // object, and in Kubernetes versions are all just views on the same entity. The assumption (which I believe Azure honors
    // is that every version of the Foo type has the same ID shape (it's fundamentally not the same kind of resource
    // if the ID has changed). Since all we need to do is look that entity up in etcd and get its name details so that
    // we can build a fully qualified name, we don't actually need version here.

    // This is the name of the kubernetes resource to reference. 
    // References across namespaces are not supported.
    Name string

    // Note that ownership across namespaces in Kubernetes is not allowed, but technically resource
    // references are. There are RBAC considerations here though so probably easier to just start by 
    // disallowing cross-namespace references
}

type AzureReference struct {
    // The subscription the Azure resource is in. If empty, this defaults to the subscription the resource 
    // holding the reference is in. 
    // TODO: If we're worried about this we could also remove it initially, ASO currently doesn't support cross sub references anyway
    Subscription string

    // The resource group the Azure resource is in. If empty, this defaults to the resource group the resource 
    // holding the reference is in.
    ResourceGroup string

    // The name of the resource. This is a subset of a fully qualified ARM ID - if you're 
    // referencing a top level resource (for example a virtual network): Microsoft.Network/virtualNetworks/{network}
    // If you're referencing a subnet: Microsoft.Network/virtualNetworks/{network}/subnets/{subnet}
    // TODO: There are a ton of options for how we do this... some of the requiring more manual work than others... we should discuss this
    Name string
}
```

This approach seems to have the most flexibility and power. It allows easy reference to static Azure resources
not managed by Kubernetes but also allows users to stay entirely within the Kubernetes ecosystem and refer to everything by
its Kubernetes name.

The primary downside of this approach is complexity - references aren't just a single string they're a complex
object and the customer will need to choose which flavor they want each time they use a reference.

### How to represent ownership and dependent resources
We will use the same `AugmentedTypeReference` type as an additional `Owner` field on dependent resource specifications.

When we determine that a resource is a dependent resource of another resource kind, we will code-generate
an `Owner` (actual name TBD) property in the dependent resource `Spec`. This will also include an annotation 
about the expected type of the resource.

```go
type SubnetSpec struct {
    Owner AugmentedKnownTypeReference `json:"ownerRef" group:"microsoft.network.infra.azure.com" kind:"VirtualNetwork"`
    ...
}
```

When users submit a dependent object we will validate that the provided owner reference is present.
This can be accomplished by making the property required in the CRD.

One major advantage of this approach is that the customer cannot really get the owning type wrong, because we've autogenerated the expected
group/kind information all names they supply must point to the right kind of resource. If they are using the `AzureReference` portion
of `AugmentedKnownTypeReference` we can look up the expected format of the ARM ID for the resource type corresponding to the referenced GVK
and validate that the customer has provided the correct one.

### Open Design Question - shape of Azure References
Instead of having an `AzureReference` as proposed above, we could instead just have the Kubernetes shaped
reference, and a way where customers of the operator could import external Azure resources into Kubernetes in
an unmanaged mode. Resources in this unmanaged mode couldn't be updated/modified through Kubernetes, and 
if deleted it would just delete the k8s representation of the object, not the actual backing object in ARM.

This is what [CAPZ does](https://github.com/kubernetes-sigs/cluster-api-provider-azure/blob/d93e75deff2be3c7647c2ddfd387ce0ef022e1cd/api/v1alpha2/tags.go#L78) today.

#### Advantages compared the proposal above
- References always look very Kubernetes native. This has an advantage especially when a customer is referring to a single
  shared resource (Storage Account or the like) across many YAMLs. The more usages they have of it the more awkward it
  gets to refer to it by fully qualified ARM ID all the time.
- The structure of the reference type would be simpler because it's not `Azure` or `Kubernetes`, it could just look like
  `KnownTypeReference` does today in k8s-infra.

#### Disadvantages compared to the proposal above
- Only able to reference other resources which are being tracked by Kubernetes. 
  We would likely need to introduce the notion of unmanaged 
  resources (which the operator _watches_ but does not _own_) for the purpose of 
  allowing references to those resources. This puts some additional burden on the customers mental model
  as in order to know what an action will do on a resource they need to know if it's managed or unmanaged.

### Implementation
There are three key pillars required for implementing the proposal as described

#### Identify resource relationships

##### Resource references (for related resources)
We must identify all resource references in the JSON schema and transform their type to `AugmentedKnownTypeReference`.
There is no specific marker which means: "This field is a reference" - most are called `id` but that's not a guarantee.
For example on the [VirtualMachineScaleSetIPConfigurationProperties](https://schema.management.azure.com/schemas/2019-07-01/Microsoft.Compute.json) 
the `subnet` field is of custom type `ApiEntityReference`, which has an `id` field where you put the ARM ID for a subnet.
This may require some manual work. One thing we can investigate doing long term
is see if there's a way to get teams to annotate "links" in their Swagger somehow...

##### Resource ownership (for dependent resources)
We must also identify all of the owner -> dependent relationships between resources.
As discussed in [what ARM does](#What-ARM-does), this is the `resources` property in the ARM deployment templates.
These are much easier to automatically detect, as the dependent types are called out in the `resources` property of the owning resource.

#### Build the fully qualified ARM ID
We have to be able to build the fully qualified ARM ID for any referenced resource when submitting to ARM.

#### Use the correct object shape for Kubernetes vs ARM
This new type of reference property replaces the corresponding `id` property, or adds a new `owner` property in the case of dependent resources.
This is a problem because now we have a single go type that has two representations - one for Kubernetes users/CRDs, where the 
`AugmentedTypeReference` is used, and one for submitting to ARM (where the `AugmentedTypeReference` must be transformed into a `name` or `id`
prior to submission). This duality means we cannot just override the JSON serialization for these types (which format would we choose?).

The proposed solution for this is that the code generator intelligently generates 2 types for cases where we know the CRD shape differs from ARM.
We will add an interface which types can optionally implement which allows them to transform themselves to
another type prior to serialization to/from ARM. This is also a useful hook for any manual customization for serialization we may need.

Here's an example of what that might look like:

```go
// TODO: This is unfortunately not well typed...?
type ARMTransformer interface {
    // TODO: Does this take an interface ArmResource which as a method ArmId() or something instead?
    // TODO: Could also have stuff like Type() and a few other things as well
    ToArm(parentArmId string) (interface{}, error)    
    FromArm(input interface{}) error
}

type VirtualNetworksSubnetsSpec struct {
    // This name is just the name of the subnet (not the fully qualified ARM name)
    // +kubebuilder:validation:Required
    Name string
	// +kubebuilder:validation:Required
	Owner genruntime.KnownTypeReference `json:"owner" kind:"VirtualNetworks" group:"microsoft.network"`
	// +kubebuilder:validation:Required
	/*Properties: Properties of the subnet.*/
	Properties SubnetPropertiesFormat `json:"properties"`
	// +kubebuilder:validation:Required
	Type VirtualNetworksSubnetsType `json:"type"`
}

// TODO: No KubeBuilder comments required here because not ever used to generate CRD
type VirtualNetworksSubnetsSpecArm struct {
	Name string
	/*Properties: Properties of the subnet.*/
	Properties SubnetPropertiesFormat `json:"properties"`
	Type VirtualNetworksSubnetsType `json:"type"`
}

// This would be autogenerated for ARM resources with references
func (o *VirtualNetworksSubnetsSpec) ToArm(parentArmId string) (interface{}, error) {
    result := VirtualNetworksSubnetsSpecArm{}
    
    // Copy properties
    result.Properties = o.Properties
    result.Type = o.Type
    
    // Name is special
    result.Name = CreateFullyQualifiedArmName(parentArmId, o.Name)
    return result, nil
}

// This would be autogenerated for ARM resources with references
func (o *VirtualNetworksSubnetsSpec) FromArm(input interface{}) (interface{}, error) {
    result := VirtualNetworksSubnetsSpecArm{}
    
    // Copy properties
    result.Properties = o.Properties
    result.Type = o.Type
    
    // Name is special
    result.Name = CreateFullyQualifiedArmName(parentArmId, o.Name)
    return result, nil
}

// Example usage
func ExamplePut() error {
    etcdObject := MagicallyGetStateFromEtcd()
    
    // Somehow we determine that this particular object has an owner - possibly we need to impl an interface?
    ownerArmId = etcdObject.FQArmId()

    var toSerialize interface{}
    toSerialize = etcdObject
    if armTransformer, ok := etcdObject.(ARMTransformer); ok {
        toSerialize = armTransformer.ToArm(ownerArmId)
    }
    json := json.Marshal(toSerialize)
    armClient.SendIt(json)

}

func ExampleGet() error {
    etcdObject := MagicallyGetStateFromEtcd()
    
    // TODO: Complete this

    // Somehow we determine that this particular object has an owner - possibly we need to impl an interface?
    ownerArmId = etcdObject.Owner.FQArmId()
    
    armId := etcdObject.FQArmId()

    // We need to provide the empty type to deserialize into
    // Somehow construct a new object of type etcdObject
    if armTransformer, ok := newEtcdObject.(ARMTransformer); ok {
        toGet = armTransformer.ToArm(ownerArmId) // This just converts from an empty etcdObject shape to an empty arm object shape
        armClient.GetIt(armId, toGet)
    
        armTransformer.FromArm(toGet)
    }
}
```

### Owner references vs Arbitrary references
Because we are code-generating all of the `ownerRef` fields, we can always supply the annotations for group and kind automatically, using the information
provided to us by the ARM json specification. **This is not the case for abitrary references (`id`'s) to other resources**. We do not actually know
programmatically what type that reference is, and it some cases it may actually be allowed to point to multiple different types 
(think custom image vs shared image gallery).

Currently, the plan is to trust the customer gets it right and ensure we have good
failure modes for resources if they actually got it wrong (which will cause their resource to move to an unhealthy state). 

## FAQ

#### What happens when a dependent resource specifies an `ownerRef` that doesn't exist?
The dependent resource will be stuck in an unprovisioned state with an error stating that the owner doesn't exist.
If the owner is created, the dependent resource will then be created by the reconciliation loop automatically.

#### What happens when a resource contains a link to another resource which doesn't exist?
The resource with the link will be stuck in an unprovisioned state with an error stating that the linked resource doesn't exist.
This behavior is the same as for a dependent resource with a non-existant owner.

#### How are the CRD entities going to be rendered as ARM deployments?
There are a few different ways to perform ARM deployments as [discussed in Dependent Resources](#dependent-resources).
Due to the nature of Kubernetes CRDs, each resource is managed separately and has its own reconcilation loop. It doesn't make sense to try to 
deploy a single ARM template with the entire resource graph. Each resource will be done in its own deployment (with
a `dependsOn` specified if required).

#### Aren't there going to be races in resource creation?
Yes. If you have a complex hierarchy of resources (where resources have involved relationships between one another) and submit 
all of their YAMLs to the operator at the same time it is likely that some requests when sent to ARM will fail because of missing dependencies.
Those resources that failed to deploy initially will be in an unprovisioned state in Kubernetes, and eventually
all the resources will be created through multiple iterations of the reconciliation loop. 

#### Aren't there going to be races in resource deletion?
Yes. `OwnerRef` as discussed in this specification is informing Kubernetes _how Azure behaves_. The fact that 
a `ResourceGroup` is the owner of a `VirtualMachineScaleSet` means that when the `ResourceGroup` is deleted in 
Azure, the `VirtualMachineScaleSet` will be too. 

This means that practically speaking, we don't need Kubernetes garbage collection to 
perform deletion of resources in Azure. Azure is already going to do that automatically. We need Kubernetes garbage collection
to easily maintain sync with Azure.

As far as implementation goes this just means that when we are performing deletes in the generic controller and the 
resource is already deleted in Azure we just swallow that error and allow the Kubernetes object to be deleted.

#### What exactly happens when a resource with an `OwnerRef` is created?
Once the resource has been accepted by the various admissions controllers and has been cofirmed to match 
the structural schema defined in the CRD, the generic controller will attempt to look up
the owning resource in etcd (or in ARM if it's an `AzureReference`).

If the generic controller finds the owning resource, it updates the `ownerReference` in the object metadata
to include the `uid` of the owning resource and then submits an ARM template to ARM using the 
name of the owner and the name of the resource to build the name specified in the ARM template. It will
include the name of the owner in the `dependsOn` field. 

#### What happens if an owning resource is deleted and immediately recreated?
Kubernetes garbage collection is based on object `uid`'s. As discussed above we bind to that `uid` on 
dependent resource creation. If a resource is deleted and then recreated Kubernetes will still understand
that the new resource is fundamentally different than the old resource and garbage collection will happen
as expected. The result will be that there is a new owning resource but all of its dependent resources were 
deleted (in Azure and in k8s).

## TODOs
- How can we allow customers to easily find all dependents for a particular owner (i.e. all subnets of a vnet) using `kubectl`?

## Questions
These are questions I am posing to the group - I don't expect to have an answer without input from the group.

- What to do with awkward resources where the owner requires at least 1 dependent to also be created with it?
  David Justice pointed out [this one](https://github.com/Azure/k8s-infra/blob/14105b1cb3f6967cd086c9f8f75fb16bb85d6318/apis/microsoft.network/v1/networkinterface_types.go#L43)
- Do we want to use the same type for ownership relationships and "related" relationships? Ownership has other angles such
  as how deletes propagate which in theory don't apply for other kinds of relationships.
- Do we need to worry about letting customers choose between [foreground cascading deletion](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion)
  and [background cascading deletion](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#background-cascading-deletion) or do we just pick 
  one behavior which is best for our case?

# The road not travelled
## How to represent references
### Use fully qualified ARM ID (a single string) for all references
#### Pros
- Super simple to implement, because it's what ARM expects at the end of the day anyway.

#### Cons
- You can't easily transplant your YAML between subscriptions/resource groups because those IDs are in the YAML - you need
  templates and variables so that you can easily move between different resource groups or Subscriptions.
- Customers can't stay in Kubernetes-land, they have to move their mental model to an "Azure" model.

### Use built-in OwnerReference for owner references (customer setting these directly)
#### Pros
- Basically none - customers are not supposed to set this directly.

#### Cons
- OwnerReference requires the object UID, which cannot be known at template authoring time.
- OwnerReference only works for ownership relationships, not for references.

## Where ownership references are specified
### Ownership is from owner to dependent
#### Pros
- It makes getting a list of all resources under a particular owner very easy.

#### Cons
- Adding/deleting a new dependent resource requires an update to the owner.
- The owner can be in a failed state because dependent resources are missing. It feels like we're repeating our intent here:
  On the one hand, we told the owner that it should have 3 dependents, while on the other hand we only created 2 of those 3.
  It feels like the state of the resources in kubernetes (i.e. how many dependents there actually _are_) is already expressing 
  the intent for how many we want, so having that also on the owner seems duplicate.

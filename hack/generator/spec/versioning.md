# Versioning

Specification for how storage versioning will operate for code generated CRD definitions.

We're generating a large volume of CRD definitions based on the JSON schema definitions available for Azure Resource Manager use.

## Goals

**Principle of Least Surprise:** The goal of the service operator is to allow users to consumer Azure resources without having to leave the tooling they are familiar with. We therefore want to do things in the idiomatic Kubernetes fashion, so that they don't experience any nasty surprises.

**Auto-generated conversions:** As far as practical, we want to autogenerate the schema for storage use along with the conversions to and from the actual ARM API versions. Hand coding all the required conversions doesn't scale across all the different Azure sevices, especially with the ongoing rate of change. 

**Allow for hand coded conversions:** While we expect to use code generation to handle the vast majority of needed conversions, we anticipate that some breaking API changes will require *part* of the conversion to be hand coded. We need to make it simple for these conversions to be introduced while still automatically generating the majority of the conversion.

**No modification of generated files.** Manual modification of generated files is a known antipattern that greatly increases the complexity and burden of updates. If some files have been manually changed, every difference showing after code generation needs to be manually reviewed before being committed. This is tedious and error prone because the vast majority of auto generated changes will be perfectly fine. 

**Compliance with Kubernetes versioning.** To quote [Kubebuilder's documentation](https://book.kubebuilder.io/multiversion-tutorial/api-changes.html):

> In Kubernetes, all versions must be safely round-tripable through each other. This means that if we convert from version 1 to version 2, and then back to version 1, we must not lose information. Thus, any change we make to our API must be compatible with whatever we supported in v1, and also need to make sure anything we add in v2 is supported in v1.

**Consistency of experience:** Early adopters should have a similar experience with the latest release of the service operator as new users who are adopting it for the first time. We don't want early adopters to be penalized for their enthusiasm. 

## Non-Goals

**Coverage of every case by code generation:** While it's likely that very high coverage will be achievable with code generation, we don't believe that it will be practical to handle every possible situation automatically. It's therefore necessary for the solution to have some form of extensibility allowing for the injection of hand written code.

## Requirements

***TBC***

## Case Studies

There are two case studies that accompany this specification, each one walking through one possible solution and showing how it will perform over time.

The [Rolling Versions](case-study-rolling-storage-versions.md) case study shows how the preferred solution adapts to changes as the Azure Resources evolve over time.

The [Fixed Version](case-study-rolling-storage-versions.md) case study shows how the primary alternative would fare, calling out some specific problems that will occur.

**TL;DR:** Using a *fixed storage version* appears simpler at first, and works well as long as the changes from version to version are simple. However, when the changes become complex (as they are bound to do over time), this approach starts to break down. While there is up front complexity to address with a *rolling storage version*, the approach doesn't break down over time.

Examples shown in this document are drawn from the case studies.

## Proposed Solution

In summary:

* For each supported Azure Resource Type, we will define a synthetic central hub type that will be used for storage/serialization across all versions of the API.

* Automatically generated conversions will allow for lossless conversions between the externally exposed API versions of resources and the central (hub) storage version.

* External metadata that we bundle with the code generator will document common changes that occur over time (including property and type name changes), extending the coverage of our automatically generated conversions.

* For cases where automatically generated conversion is not sufficient, standard extension points for each resource type will allow hand-coded conversion steps to be injected into the process at key points.

Each of these four points is expanded upon in detail below.

### Defining a central hub type

We'll base the schema of the central hub type on the latest GA release of the API for each resource, with the following modifications:

**All properties will be defined as optional** allowing for back compatibility with prior versions of the API that might not have included specific properties.

**Inclusion of a property bag** to provide for storage for properties present in older versions of the API that are no longer present.

Using a purpose designed hub type for storage avoids a number of version-to-version compatibility issues that can arise if the API version itself is used directly for storage.

To illustrate, if the API version defined the following `Person` type:

``` go
package v20110101

type Person struct {
    Id        Guid
    FirstName string
    LastName  string
}
```

Then the generated storage (hub) version will look like this:

``` go
package v20110101storage

type Person struct {
    PropertyBag
    FirstName   *string
    Id          *Guid
    LastName    *string
}
```

Using the latest version of the API as the basis for our storage version gives us maximum compatibility for the usual case, where a user defines their custom resource using the latest available version.

If a resource type has been dropped from the ARM API, we will still generate a storage schema for it based on the last ARM API version where it existed; this helps to ensure backward compatibility with existing service operator deployments.

Sequestering additional properties away within a property bag in the storage schema is more robust than using separate annotations as they are less subject to arbitary modification by users. This allows us to largely avoid situations where well meaning (but uninformed) consumers of the service operator innocently make changes that result in the operator becoming failing. We particularly want to avoid this failure mode because recovery will be difficult - restoration of the modified/deleted information may be impractical or impossible.

### Generated conversion methods

Each of the structs generated for ARM API will have the normal `ConvertTo()` and `ConvertFrom()` methods generated automatically, implementing the required [Convertible](https://book.kubebuilder.io/multiversion-tutorial/conversion.html) interface:

``` go
// ConvertTo converts this Person to the Hub storage version.
func (person *Person) ConvertTo(raw conversion.Hub) error {
    p := raw.(*storage.Person)
    return ConvertToStorage(p)
}

// ConvertFrom converts from the Hub storage version
func (person *Person) ConvertFrom(raw conversion.Hub) error {
    p := raw.(*storage.Person)
    return ConvertFromStorage(p)
}
```

As shown, these methods will delegate to two helper methods (`ConvertToStorage()` and `ConvertFromStorage()`) that are generated to handle the process of copying information across between instances.

Properties defined in the API type will be handled by the following rules:

1. For properties with a primitive type (**string**, **int**, **float**, or **bool**):

   * If the storage type has a matching property with the same name and type, the value will be directly copied across.
   * If the storage type does not have a matching property, the value will be stored/recalled using the property bag.

1. For properties with an enumeration type:

   * If the storage type has a matching property with the same name and a compatible type, the value will be directly copied across (with a cast if necessary).
   * If the storage type does not have a matching property, the value will be stored/recalled using the property bag.

1. For properties with a complex type (a type defined by the JSON schema)

   * If the storage type has a matching property with the same name and a compatible type, the `ConvertToStorage()` and `ConvertFromStorage()` methods defined on the API version of the type will be used to copy information across.
   * If the storage type does not have a matching property, the value will stored/recalled in JSON format using the property bag

Notes
* Property name comparisons are case-insensitive. See also the section below on property renaming.

* Enumeration types defined in two different versions are considered compatible if they have the same underlying base type and their names are the same (by case-insensitive comparison)

* Complex types defined in two different versions are considered compatible if they have the same name (by case-insensitive comparison).

> ***TODO: Show an example that includes all the cases***


### External Metadata for common changes

We'll capture common changes between versions in metadata that we bundle with the code generator, allowing it to handle a wider range of scenarios.

**If a property is renamed** in a particular API version, conversion of API versions *prior* to that point of change will instead match based on the new name of the property on the storage type.

> ***Outstanding issue***: Do we want to support **property** renaming in this way? How much will it help reduce manual boilerplate?

> ***TODO: Show an example***

**If a type has been renamed** in a particular API version, conversion of API versions *prior* to that point of change will instead match based on the new type of the property on the storage type.

> ***Outstanding issue***: Do we want to support **type** renaming in this way? How much will it help reduce manual boilerplate?

> ***TODO: Show an example***

> ***Outstanding Issue:*** Are there other kinds of common change we want to support?  
Are there other cases of changes between versions that we may be able to handle automatically. 
Can we find examples? Do we want to support these cases?

### Standard extension points

Code generation will include interfaces to allow easy injection of manual conversion steps. 

For each storage type, two interfaces will be generated, one to be called by `ConvertToStorage()` and one to be called by `ConvertFromStorage()`:

``` go
type AssignableToPerson interface {
    AssignToPerson(person Person) error
}

type AssignableFromPerson interface {
    AssignFromPerson(person Person) error
}
```
If an API type implements one (or both) of these interfaces, they will be automatically invoked *after* the standard conversion code has completed.

## Testing

It's vital that we are able to correctly convert between versions. We will therefore generate a set of unit tests to help ensure that the conversions work correctly. Coverage won't be perfect (as there are conversion steps we can't automatically verify) but these tests will help ensure correctness.

### Round Trip Testing

We will generate a unit test to ensure that every spoke version can round trip to the hub version and back again with no loss of information.

This will help to ensure a base level of compliance, that information is not lost through serialization.

> TODO: Flesh this out
* Use of fuzz testing to generate instances
* ConvertToStorage() followed by ConvertFromStorage() and then compare that all properties match

* **string**, **int**, **bool** much match exactly
* **Float64** match within tollerance
* What else?

### Forward Testing


### Testing extensibility


## Alternative Solutions

AKA the road not travelled

### Alternative: Fixed storage version

The "v1" storage version of each supported resource type will be created by merging all of the fields of all the distinct versions of the resource type, creating a *superset* type that includes every property declared across every version of the API.

To maintain backward compatibility as Azure APIs evolve over time, we will include properties across all versions of the API, even for versions we are not currently generating as output. This ensures that properties in use by older APIs are still present and available for forward conversion to newer APIs, even as those older APIs age out of use.

> ***TODO***: Reference the case study

> ***TODO***: Copy in to here the limitations described in the case study

----
----

#### Limitation: Old properties aging out

When a changed property ages out and the schema is no longer available, the code generator will be unaware of its existence. Therefore, it will be omitted from our generated storage schema resulting in a breaking change for any resource already persisted.

* This problem becomes worse if we include only the latest two (or latest ***n***) versions of the schema when generating the storage version as old properties will drop out much more quickly.

* Mitigating the issue by merging the schema version to ensure old properties are not lost over time is possible, but will require educated manual intervention each time the code generator is run to update the generated types from ARM. This will be fragile and prone to error.

#### Limitation: Property type changes

If the type of a property changes, we don't have any good way to define a concrete Go type that easily permits both values. 

* Changing the name of the deprecated property would be a breaking change as we wouldn't be able to rehydrate resources already persisted by an older version of the operator.

* Using a different name for the *new* version of the property would be possible, but runs the risk a naming collision with other properties. It also creates a wart that would persist for the life of the tool. This wart would be encountered even for users who never used the older version of the operator.

To illustrate the sorts of changes we need to support, consider these changes made by **Microsoft.ServiceFabric**:

| ClusterProperties  | v20160301                | v20160901                         |
| ------------------ | ------------------------ | --------------------------------- |
| NodeTypes          | []NodeTypes              | []NodeTypeDescription             |
| ReliabilityLevel   | Level                    | ClusterPropertiesReliabilityLevel |
| UpgradeDescription | PaasClusterUpgradePolicy | ClusterUpgradePolicy              |

----
----

### Alternative: Use the latest API version

The supported storage version of each resource type will simply be the latest version of the API supported by ARM. Any additional information not supported on that version will be persisted via annotations on the resource.

**Limitation**: This is a known antipattern that we should avoid.

TODO: Include references from DavidJ

**Limitation**: Annotations are publicly visible on the cluster and can easily modified. This makes it spectacularly easy for a user to make an innocent change that would break the functionality of the operator. 

TODO: Verify this assumption

## See Also

* [Hubs, spokes, and other wheel metaphors](https://book.kubebuilder.io/multiversion-tutorial/conversion-concepts.html)

* [Falsehoods programmers believe about addresses](https://www.mjt.me.uk/posts/falsehoods-programmers-believe-about-addresses/)

https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definition-versioning/


## Glossary

**ARM:** Azure Resource Manager

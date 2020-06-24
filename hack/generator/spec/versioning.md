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

The [**Rolling Versions**](case-study-rolling-storage-versions.md) case study shows how the preferred solution adapts to changes as the Azure Resources evolve over time.

The [**Fixed Version**](case-study-rolling-storage-versions.md) case study shows how the primary alternative would fare, calling out some specific problems that will occur.

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
    Id          *Guid
    FirstName   *string
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

As shown, these methods will delegate to two strongly typed helper methods (`ConvertToStorage()` and `ConvertFromStorage()`) that are generated to handle the process of copying information across between instances.

The `ConvertToStorage()` method is responsible for copying all of the properties from the API type onto the storage type. The `ConvertFromStorage()` method is its mirror, responsible for populating all of the properties on the API type from the storage type.

Each property defined in the API type is considered in turn, and will require different handling based on its type and whether a suitable match is found on the storage type:

![](images/versioning-property-mapping-flowchart.png)

**For properties with a primitive type** a matching property must have the same name and the identical type. If found, a simple assignment will copy the value over. If not found, the value will be stashed-in/recalled-from the property bag present on the storage type.

* Primitive types are **string**, **int**, **float64**, and **bool**
* Name comparisons are case-insensitive

**For properties with an enumeration type** a matching property must have the same property name and an enumeration type with the same type name. If found, a simple assignment will copy the value over. If not found, the value will be stashed-in/recalled-from the property bag present on the storage type using the underlying type of the enumeration.

* Name comparisons are case-insensitive for both property names and enumeration type names
* Enumeration types are generated independently for each version, so they will never be identical types

**For properties with a custom type** a matching property must have the same name and a custom type with same type name. If found, a new instance will be created and the appropriate `ConvertToStorage()` or `ConvertFromStorage()` method for the custom type will be used. If not found, JSON serialization will be used with the property bag for storage.

* Name comparisons are case-insensitive for both property names and custom type names
* Custom types are generated independently for each version, so they will never be identical types

> ***TODO: Show an example that includes all the cases***

### External Metadata for common changes

We'll capture common changes between versions in metadata (likely a YAML file) that we bundle with the code generator, allowing it to handle a wider range of scenarios.

**If a property is renamed** in a particular API version, conversion of API versions *prior* to that point of change will instead match based on the new name of the property on the storage type. 

There are more than 40 cases of properties being renamed across versions of the ARM API.

> ***TODO: Show an example***

**If a type has been renamed** in a particular API version, conversion of API versions *prior* to that point of change will instead match based on the new type of the property on the storage type.

Therea are 160 cases of properties changing type acdross versions of the ARM API. Many of these can be handled automatically by capturing type renames in metadata.

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

We will generate a unit test to ensure that every spoke version can round trip to the hub version and back again to the same type with no loss of information.

This will help to ensure a base level of compliance, that information is not lost through serialization.

> TODO: Flesh this out
* Use of fuzz testing to generate instances
* ConvertToStorage() followed by ConvertFromStorage() and then compare that all properties match

* **string**, **int**, **bool** much match exactly
* **Float64** match within tolerance
* What else?

### Golden Tests

For API (spoke) types where the optional interfaces `AssignableTo...()` and `AssignableFrom...()` have been implemented, we'll generate golden tests to verify that they are generating the expected results.

These tests will be particularly useful when a new version of the ARM api is released for a given service as they will help to catch any changes that now need to be handled.

We'll generate two golden tests for each type in each API type, one to test verify conversion _**to**_ the latest version, and one to test conversion _**from**_ the latest version.

**Testing conversion to the latest version** will check that an instance of a older version of the API can be correctly upconverted to the latest version:

![](images/versioning-golden-tests-to-latest.png)

The test will involve these steps:

* Create an exemplar instance of the older API type 
* Convert it to the storage type using `ConvertToStorage()`
* Convert it to the latest API type using `ConvertFromStorage()`
* Check that it matches the golden file from a previous run

Testing will only occur if one (or both) types implements one of the optional interfaces. That is, one or both of the following must be true:
* The older API type implements `AssignableTo...()`
* The latest API type implements `AssignableFrom...()`

If neither rule is satisfied, the test will silently null out.

**Testing conversion from the latest version** will check that an instance of the latest version of the API can be correctly downconverted to an older version.

![](images/versioning-golden-tests-from-latest.png)

* Create an exemplar instance of the latest API type 
* Convert it to the storage type using `ConvertToStorage()`
* Convert it to the older API type using `ConvertFromStorage()`
* Check that it matches the golden file from a previous run

Testing will only occur if one (or both) types implements one of the optional interfaces. That is, one or both of the following must be true:
* The older API type implements `AssignableFrom...()`
* The latest API type implements `AssignableTo...()`

If neither rule is satisfied, the test will silently null out.

## Alternative Solutions

AKA the road not travelled

### Alternative: Fixed storage version

The "v1" storage version of each supported resource type will be created by merging all of the fields of all the distinct versions of the resource type, creating a *superset* type that includes every property declared across every version of the API.

To maintain backward compatibility as Azure APIs evolve over time, we will include properties across all versions of the API, even for versions we are not currently generating as output. This ensures that properties in use by older APIs are still present and available for forward conversion to newer APIs, even as those older APIs age out of use.

This approach has a number of issues that are called out in detail in the [fixed storage version case study](case-study-fixed-storage-version.md).

**Property Bloat**: As our API evolves over time, our storage version is accumulating all the properties that have ever existed, bloating the storage version with obsolete properties that are seldom (if ever) used. Even properties that only ever existed on a single preview release of an ARM API need to be correctly managed for the lifetime of the service operator.

**Property Amnesia**: Our code generator only knows about properties defined in current versions of the API. Once an API version has been excluded (or if the JSON schema definition is no longer available), the generator completely forgets about older properties. This would cause compatibility issues for established users who would find upgrading the service operator breaks their cluster.

**Type Collision**: Identically named properties with different types can't be stored in the same property; mitigation is possible for a limited time, though eventually *property amnesia* will cause a breaking change.

### Alternative: Use the latest API version

The supported storage version of each resource type will simply be the latest version of the API supported by ARM. Any additional information not supported on that version will be persisted via annotations on the resource.

**This is a known antipattern that we should avoid.**

Annotations are publicly visible on the cluster and can easily modified. This makes it spectacularly easy for a user to make an innocent change that would break the functionality of the operator. 

## See Also

* [Hubs, spokes, and other wheel metaphors](https://book.kubebuilder.io/multiversion-tutorial/conversion-concepts.html)

* [Falsehoods programmers believe about addresses](https://www.mjt.me.uk/posts/falsehoods-programmers-believe-about-addresses/)

https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definition-versioning/

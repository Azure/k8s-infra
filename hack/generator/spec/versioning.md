# Versioning

Specification for how storage versioning will operate for code generated CRD definitions.

## Goals

**Auto-generated conversions:** As far as practical, we want to autogenerate the schema for each hub version along with the conversions to and from the actual ARM API versions. Hand coding all the required conversions doesn't scale across all the different Azure sevices, especially with the ongoing rate of change. 

**Allow for hand coded conversions:** While we expect to use code generation to handle the vast majority of needed conversions, we anticipate that some breaking API changes will require some of the conversion to be hand coded. We need to make it simple for these conversions to be introduced while still automatically generating the majority of the conversion.

**No modification of generated files.** Manual modification of generated files is a known antipattern that greatly increases the complexity and burden of updates. If some files have been manually changed, every difference showing after code generation needs to be manually reviewed before being committed. This is tedious and error prone because the vast majority of auto generated changes will be perfectly fine. 

## Non-Goals

**Coverage of every case by code generation:** While it's likely that very high coverage will be achievable with code generation, we don't believe that it will be practical to handle every possible situation automatically. It's therefore necessary for the solution to have some form of extensibility allowing for the injection of hand written code.

## Proposed Solution

In summary (details below):

* Base the storage version on the latest available API version for each resource type
* Augment the storage version with a name/value property bag for storing additional information
* Generate the required `ConvertTo()` and `ConvertFrom()` methods automatically
* Allow configuration to support common changes from version to version
* Define extension points for hand written conversions

### Use the latest available API Version

As a starting point for defining the storage schema for each resource type, we will use the latest available version of the matching ARM API. This ensures maximum compatibility with that API when it is used, giving a good starting point for new adoption of the service operator.

If a resource type has been dropped from the ARM API, we will still generate a storage schema for it based on the last ARM API version where it existed; this ensures backward compatibility with existing service operator deployments.

### Augment storage with a name/value property bag

To facilitate easy and safe storage of additional configuration in storage, we'll include a name/value property bag for storage, along with helper methods that allow for strongly typed access to values stored within.

Sequestering the information away within the storage schema is more robust than using separate annotations as they are less subject to arbitary modification by users. This allows us to largely avoid situations where well meaning (but uninformed) consumers of the service operator innocently make changes that result in the operator becoming failing. We particularly want to avoid this failure mode because recovery will be difficult - restoration of the modified/deleted information may be impractical or impossible.

---

****Outstanding issue:** Where do the helper functions for the name/value property bag reside?**

We could generate those anew for each type, making them conveniently available and avoiding the need for our generated types to depend on anything more than the standard Go library Alternatively, we could define the property bag as a new struct, wrapping up storage & helper methods into a reusable form.

---

### Generate conversion methods

Each of the structs generated for ARM API will have the normal `ConvertTo()` and `ConvertFrom()` methods generated automatically, mapping like-for-like properties between the hub version and the current type.

* When the name and type of properties match exactly, conversion in both directions will be simple assignments.
* Where the name matches but the type does not, the property will be skipped by default.
* If missing on the hub version, the name/value property bag will also be used.

### Configuration for common changes

We'll facilitate common types of changes between versions by providing explicit configuration options.

**Property rename** - if a property is simply renamed between versions (say for a spelling correction), we can generate the required assignment for compatiblity between spoke and hub versions.

**Type rename** - if a subtype is renamed between versions, maintaining largely the same structure, we can automatically map between those two types.


---

***Outstanding Issue:** What do we do for other kinds of version-to-version change?*

Are there other cases of changes between versions that we may be able to handle automatically. 
Can we find examples? Do we want to support these cases?

---


### Define extension points

Optional interfaces will be generated to allow injection of additional steps to the generated conversions. If implemented on the spoke versions (not the hub version), they will be automatically invoked as required.

For each actual API version, an interface will be generated containing strongly typed `AssignTo()` and `AssignFrom()` methods:

``` go
type ClusterPropertiesv20160301Converter {
    AssignTo(clusterProperties ClusterPropertiesvStorage)
    AssignFrom(clusterProperties ClusterPropertiesvStorage)
}
```

If implemented on the v20160301 version of ClusterProperties, it will be automatically invoked at the end of the standard `ConvertTo()` and `ConvertFrom()` methods, allowing for their behaviour to be modified/augmented.

When the type of a property is changed between versions, we'll generate an interface allowing for conversion between those two types.

``` go
type ClusterPropertiesv20160301ReliabilityLevelConverter {
    ConvertLevelToClusterPropertiesReliabilityLevel(level Level) *ClusterPropertiesReliabilityLevel
    ConvertClusterPropertiesReliabilityLevelToLevel(clusterPropertiesReliabilityLevel ClusterPropertiesReliabilityLevel) Level
}
```

Again, if implemented on the v20160301 version of ClusterProperties, these will be automaticaly invoked by `ConvertTo()` and `ConvertFrom()`.

***Issue:** Should these be invoked before or after the main interface? We need to define the sequence for predictability.*

## Illustration of Conversion

To illustrate the way conversions will work using `ClusterProperties` from Microsoft.ServiceFabric. This explores conversion between version v20160301 and the hub storage version, based on a later version:

| Fields                 | v20160301                        | vStorage (Hub)                     | Change           | Handler      |
| --------------------------------- | -------------------------------- | ---------------------------------- | ---------------- | ------------ |
| AzureActiveDirectory              | *AzureActiveDirectory            | *AzureActiveDirectory              | None             | Copy value   |
| Certificate                       | *CertificateDescription          | *CertificateDescription            | None             | Copy value   |
| ClientCertificateCommonNames      | *[]ClientCertificateCommonName   | *[]ClientCertificateCommonName     | None             | Copy value   |
| ClientCertificateThumbprints      | *[]ClientCertificateThumbprint   | *[]ClientCertificateThumbprint     | None             | Copy value   |
| ClusterCodeVersion                |                                  | *string                            | New Property     | Skip         |
| DiagnosticsStorageAccountConfig   | *DiagnosticsStorageAccountConfig | *DiagnosticsStorageAccountConfig   | None             | Copy value   |
| FabricSettings                    | *[]SettingsSectionDescription    | *[]SettingsSectionDescription      | None             | Copy value   |
| HttpApplicationGatewayCertificate | *CertificateDescription          |                                    | Property Removed | Property Bag |
| ManagementEndpoint                | string                           | string                             | None             | Copy value   |
| NodeTypes                         | []NodeTypes                      | []NodeTypeDescription              | Change of Type   | Property Bag |
| ReliabilityLevel                  | *Level                           | *ClusterPropertiesReliabilityLevel | Change of Type   | Property Bag |
| ReverseProxyCertificate           |                                  | *CertificateDescription            | New Property     | Skip         |
| UpgradeDescription                | *PaasClusterUpgradePolicy        | *ClusterUpgradePolicy              | Change of Type   | Property Bag |
| UpgradeMode                       |                                  | *ClusterPropertiesUpgradeMode      | New Property     | Skip         |
| VmImage                           | *string                          | *string                            | None             | Copy value   |

Of the 15 properties listed:

* 8 are unchanged (same name and type); these values are simply copied across.  
  (AzureActiveDirectory, Certificate, ClientCertificateCommonNames, ClientCertificateThumbprints, DiagnosticsStorageAccountConfig, FabricSettings, ManagementEndpoint, and VmImage)
* 3 new properties are introduced; these are ignored by the conversion.  
  (ClusterCodeVersion, ReverseProxyCertificate, and UpgradeMode)
* 1 property was dropped; it will be stored in the name/value property bag.  
  (HttpApplicationGatewayCertificate)
* 3 properties had their types changed; these will also be serialized into the name/value property bag.  
  (NodeTypes, ReliabilityLevel, and UpgradeDescription)




## Testing

It's vital that we are able to correctly round trip every spoke version to the hub version and back again, so we will generate unit tests to ensure that this works properly. This will help to enusre a base level of compliance, that information is not lost through serialization.



## Alternative Solutions

AKA the road not travelled

### Alternative: Superset storage schema

The "v1" storage version of each supported resource type will be created by merging all of the fields of all the distinct versions of the resource type, creating a *superset* type that includes every property declared across every version of the API.

To maintain backward compatibility as Azure APIs evolve over time, we will include properties across all versions of the API, even for versions we are not currently generating as output. This ensures that properties in use by older APIs are still present and available for forward conversion to newer APIs, even as those older APIs age out of use.

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


### Alternative: Use the latest API version

The supported storage version of each resource type will simply be the latest version of the API supported by ARM. Any additional information not supported on that version will be persisted via annotations on the resource.

**Limitation**: This is a known antipattern that we should avoid.

TODO: Include references from DavidJ

**Limitation**: Annotations are publicly visible on the cluster and can easily modified. This makes it spectacularly easy for a user to make an innocent change that would break the functionality of the operator. 

TODO: Verify this assumption

## See Also

[Hubs, spokes, and other wheel metaphors](https://book.kubebuilder.io/multiversion-tutorial/conversion-concepts.html)

## Glossary

**ARM:** Azure Resource Manager

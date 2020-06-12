# Versioning

Specification for how storage versioning will operate for code generated CRD definitions.

## Goals

**Auto-generated conversions:** As far as practical, we want to autogenerate the schema for each hub version along with the conversions to and from the actual ARM API versions. Hand coding all the required conversions doesn't scale across all the different Azure sevices, especially with the ongoing rate of change. 

**Allow for hand coded conversions:** While we expect to use code generation to handle the vast majority of needed conversions, we anticipate that some breaking API changes will require some of the conversion to be hand coded. We need to make it simple for these conversions to be introduced while still automatically generating the majority of the conversion.

**No modification of generated files.** Manual modification of generated files is a known antipattern that greatly increases the complexity and burden of updates. If some files have been manually changed, every difference showing after code generation needs to be manually reviewed before being committed. This is tedious and error prone because the vast majority of auto generated changes will be perfectly fine. 

**Compliance with Kubernetes versioning.** To quote [Kubebuilder's documentation](https://book.kubebuilder.io/multiversion-tutorial/api-changes.html):

> In Kubernetes, all versions must be safely round-tripable through each other. This means that if we convert from version 1 to version 2, and then back to version 1, we must not lose information. Thus, any change we make to our API must be compatible with whatever we supported in v1, and also need to make sure anything we add in v2 is supported in v1.

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
    AssignTo(clusterProperties *ClusterPropertiesvStorage)
    AssignFrom(clusterProperties *ClusterPropertiesvStorage)
}
```

If implemented on the v20160301 version of ClusterProperties, it will be automatically invoked at the end of the standard `ConvertTo()` and `ConvertFrom()` methods, allowing for their behaviour to be modified/augmented.

When the type of a property is changed between versions, we'll generate an interface allowing for conversion between those two types.

``` go
type ClusterPropertiesv20160301ReliabilityLevelConverter {
    ConvertLevelToClusterPropertiesReliabilityLevel(level Level) *ClusterPropertiesReliabilityLevel
    ConvertClusterPropertiesReliabilityLevelToLevel(clusterPropertiesReliabilityLevel *ClusterPropertiesReliabilityLevel) Level
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


## Case Study - Rolling Storage Versions

In this section, we'll explore a case study showing the evolution of a hypothetical ARM resource over time and show how this will be handled with rolling storage versions. Synthetic examples are used to allow them to focus on specific cases one by one, providing motivation for specific features as defined elsewhere.

Examples shown in this section are deliberately simplified in order to focus on specific details, and therefore details here should not be considered binding. Reference the formal specification sections of this document for precise details.

For the case study, we'll be following the version by version evolution of a theoretical ARM service that provides customer resource management (CRM) services. 

### Version 2011-01-01 - Initial Release

The initial release of the CRM includes a simple definition to capture information about a particular person

``` go
package v20110101

type Person struct {
    Id        Guid
    FirstName string
    LastName  string
}
```

Our storage version is always independent from the API version (independent types), but with a structure based on the latest (in this case the _only_) version of the resource. It therefore has a very similar definition:

``` go
package v20110101storage

type Person struct {
    PropertyBag
    Id        *Guid
    FirstName *string
    LastName  *string
}

// Hub marks this type as a conversion hub.
func (*Person) Hub() {}
```

Every property is marked as optional. Optionality doesn't matter at this point, as we are only concerned with a single version of the API. However, as we'll see with later versions, forward and backward compatibility issues would arise if they were not optional.

The `PropertyBag` type provides storage for other properties, plus helper methods. It is always included in storage versions, but in this case will be unused. The method `Hub()` marks this version as the storage schema.

Our original version now needs to implement the [Convertible](https://book.kubebuilder.io/multiversion-tutorial/conversion.html) interface to allow conversion each way:

``` go
package v20110101

import storage "v20110101storage"

// ConvertTo converts this Person to the Hub storage version.
func (person *Person) ConvertTo(raw conversion.Hub) error {
    p := raw.(*storage.Person)
    return ConvertToStorage(p)
}

// ConvertToStorage converts this Person to a storage version
func (person *Person) ConvertToStorage(dest storage.Person) error {
    // Copy simple properties across
    dest.Id = person.Id
    dest.FirstName = person.FirstName
    dest.LastName = person.LastName

    return nil
}

// ConvertFrom converts from the Hub storage version
func (person *Person) ConvertFrom(raw conversion.Hub) error {
    p := raw.(*storage.Person)
    return ConvertFromStorage(p)
}

// ConvertFrom converts from a storage version to this version.
func (person *Person) ConvertFromStorage(source storage.Person) error {
    // Copy simple properties across
    person.Id = source.Id
    person.FirstName = source.FirstName
    person.LastName = source.LastName

    return nil
}

```

These methods will be automatically generated in order to handle the majority of the required conversions. Since they never change, the `ConvertTo()` and `ConvertFrom()` methods are omitted from the following discussion.

### Version 2012-02-02 - No Change

In this release of the CRM service, there are no changes made to the structure of `Person`:

``` go
package v20120202

type Person struct {
    Id        Guid
    FirstName string
    LastName  string
}
```

Conversions to and from the storage version will be identical (except for the import statements for referenced types) to those generated for the prior version. 

### Version 2013-03-03 - New Property

In response to customer feedback, this release of the CRM adds a new property to `Person` to allow a persons middle name to be stored:

``` go
package v20130303

type Person struct {
    Id         Guid
    FirstName  string
    MiddleName string // *** New in this version ***
    LastName   string
}
```

The new storage version, based on this version, updates accordingly:

``` go
package v20130303storage

type Person struct {
    PropertyBag
    Id         *Guid
    FirstName  *string
    MiddleName *string // *** New storage for new property ***
    LastName   *string
}

// Hub marks this type as a conversion hub.
func (*Person) Hub() {}
```

Conversions to and from earlier versions of Person are unchanged, as those versions do not support `MiddleName`. For the new version of `Person`, the new property will be included in the generated methods:

``` go
package v20130303

import storage "v20130303storage"

// ConvertTo converts this Person to the Hub version.
func (person *Person) ConvertToStorage(dest storage.Person) error {
    dest.Id = person.Id
    dest.FirstName = person.FirstName
    dest.LastName = person.LastName
    dest.MiddleName = person.MiddleName // *** New property copied too ***

    return nil
}

// ConvertFrom converts from the Hub version to this version.
func (person *Person) ConvertFromStorage(source storage.Person) error {
    person.Id = source.Id
    person.FirstName = source.FirstName
    person.LastName = source.LastName
    person.MiddleName = source.MiddleName // *** New property copied too ***

    return nil
}
```

The new property is shown at the end of the list not because it is new, but because values are copied across in alphabetical order to guarantee that code generation is deterministic and generates the same result each time.

Conversion methods for earlier API versions of `Person` are essentially unchanged. The import statement at the top of the file will be updated to the new storage version; no other changes are necessary.

*Statistics:* At the time of writing, there were 381 version-to-version changes where the only change between versions was the addition of new properties. Of those, 249 were adding just a single property, and 71 added two properties. 

### Version 2014-04-04 Preview - Schema Change

To allow the CRM to better support cultures that have differing ideas about how names are written, a preview release of the service modifies the schema considerably:

``` go
package v20140404preview

type Person struct {
    Id         Guid   // ** Only Id is unchanged from the prior version ***
    FullName   string
    FamilyName string
    KnownAs    string
}
```

This is a preview version, so the storage version is _left unchanged_, based on the latest non-preview release (version 2011-03-03). We don't want to make changes to our storage versions based on speculative changes.

The new properties don't exist on the storage version of `Person`, so the generated `ConvertToStorage()` and `ConvertFromStorage()` methods use the `PropertyBag` to carry the properties:

``` go
package v20140404preview

import storage "v20130303storage"

// ConvertTo converts this Person to the Hub version.
func (person *Person) ConvertToStorage(dest storage.Person) error {
    dest.Id = person.Id

    // *** Store in the property bag ***
    dest.WriteString("FamilyName", person.FamilyName)
    dest.WriteString("FullName", person.FullName)
    dest.WriteString("KnownAs", person.KnownAs)

    return nil
}

// ConvertFrom converts from the Hub version to this version.
func (person *Person) ConvertFromStorage(source storage.Person) error {
    person.Id = source.Id

    // *** Read from the property bag ***
    person.FamilyName = source.ReadString("FamilyName")
    person.FullName = source.ReadString("FullName")
    person.KnownAs = source.ReadString("KnownAs")

    return nil
}
```

This provides round-trip support for the preview release, but does not provide backward compatibility with prior official releases. 

The storage version of `Person` written by the preview release will have no values for `FirstName`, `LastName`, and `MiddleName`.

These kinds of cross-version conversions cannot be automatically generated as they require more understanding the semantic changes between versions. 

To allow injection of manual conversion steps, an interface will be generated as follows:

``` go
package v20130303storage

// AssignableWithPersonStorage provides methods to
// augment conversion to/from the storage version
type AssignableWithPersonStorage interface {
    AssignTo(person Person) error
    AssignFrom(person Person) error
}
```

This interface can be optionally implemented by API versions (spoke types) to augment the generated conversion.

The generated `ConvertToStorage()` and `ConvertFromStorage()` methods will test for the presence of this interface and will call it if available:

``` go
package v20140404preview

import storage "v20130303storage"

// ConvertTo converts this Person to the Hub version.
func (person *Person) ConvertToStorage(dest storage.Person) error {
    // … elided …

    // *** Check for the interface and use it if found ***
    if assignable, ok := person.(AssignableWithPersonStorage); ok {
        assignable.AssignTo(dest)
    }

    return nil
}

// ConvertFrom converts from the Hub version to this version.
func (person *Person) ConvertFromStorage(source storage.Person) error {
    // … elided …

    // *** Check for the interface and use it if found ***
    if assignable, ok := person.(AssignableWithPersonStorage); ok {
        assignable.AssignFrom(source)
    }

    return nil
}
```

### Version 2014-04-04 - Schema Change

Based on feedback generated by the preview release, the CRM schema changes have gone ahead with a few minor changes:

``` go
package v20140404

type Person struct {
    Id         Guid
    LegalName  string // Was FullName in preview
    FamilyName string
    KnownAs    string
    AlphaKey   string // Added after preview
}
```

No longer being a preview release, the storage version is also regenerated:

``` go
package v20140404storage

type Person struct {
    PropertyBag
    Id         *Guid
    LegalName  *string
    FamilyName *string
    KnownAs    *string
    AlphaKey   *string
}

// Hub marks this type as a conversion hub.
func (*Person) Hub() {}
```

The `ConvertToStorage()` and `ConvertFromStorage()` methods for the new version of `Person` are generated as expected, copying across values and invoking the `AssignableWithPersonStorage` interface if present:

``` go
package v20140404

import storage "v20140404storage"

// ConvertTo converts this Person to the Hub version.
func (person *Person) ConvertToStorage(dest storage.Person) error {
    dest.AlphaKey = person.AlphaKey
    dest.FamilyName = person.FamilyName
    dest.Id = person.Id
    dest.KnownAs = person.KnownAs
    dest.LegalName = person.LegalName

    // *** Check for the interface and use it if found ***
    if assignable, ok := person.(AssignableWithPersonStorage); ok {
        assignable.AssignTo(dest)
    }

    return nil
}

// ConvertFrom converts from the Hub version to this version.
func (person *Person) ConvertFromStorage(source storage.Person) error {
    person.AlphaKey = source.AlphaKey
    person.FamilyName = source.FamilyName
    person.Id = source.Id
    person.KnownAs = source.KnownAs"
    person.LegalName = source.LegalName

    // *** Check for the interface and use it if found ***
    if assignable, ok := person.(AssignableWithPersonStorage); ok {
        assignable.AssignFrom(source)
    }

    return nil
}
```

For older versions of `Person`, the conversion methods change considerably when regenerated. The properties they used to use are no longer present on the storage version, so instead of direct assignment, they now use the `PropertyBag` to stash the required values away:

``` go
package v20110101

import storage "v20140404storage"

// ConvertTo converts this Person to the Hub version.
func (person *Person) ConvertToStorage(dest storage.Person) error {
    dest.Id = person.Id
    dest.WriteString("FirstName", person.FirstName)
    dest.WriteString("LastName",  person.LastName)

    if assignable, ok := person.(AssignableWithPersonStorage); ok {
        assignable.AssignTo(dest)
    }

    return nil
}

// ConvertFrom converts from the Hub version to this version.
func (person *Person) ConvertFromStorage(source storage.Person) error {
    person.Id = source.Id
    person.FirstName = source.ReadString("FirstName")
    person.LastName = source.ReadString("LastName")

    if assignable, ok := person.(AssignableWithPersonStorage); ok {
        assignable.AssignFrom(source)
    }

    return nil
}
```

To interoperate between different versions of `Person`, we need to add some manual conversions.

When a newer version of `Person` is written to storage, we need to also populate `FirstName`, `LastName` and `MiddleName` within the `PropertyBag` to allow older versions to be requested. Similarly, when an older version of `Person` is written, we need to populate `AlphaKey`, `FamilyName`, `KnownAs` and `LegalName` so that newer versions can be requested.

To avoid repetition of code across multiple implementations of `AssignTo()` and `AssignFrom()`, we write some helper methods on the storage version:

``` go
package v20140404storage

func (person *Person) PopulateFromFirstMiddleLastName(firstName string, middleName string, lastName string) {
    person.KnownAs = firstName
    person.FamilyName = lastName
    person.LegalName = firstName +" "+ middleName + " " + lastName
    person.AlphaKey = lastName
}

func (person *Person) PopulateLegacyFields() {
    person.WriteString("FirstName", person.KnownAs)
    person.WriteString("LastName",  person.FamilyName)
    person.WriteString("MiddleName", ... elided ...)
}
```

With these methods available, implementing the interface `AssignableWithPersonStorage` becomes straightforward. For the first release of `Person`:

``` go
package v20110101

import storage "v20140404storage"

func (person *Person) AssignTo(dest storage.Person) error {
    dest.PopulateFromFirstMiddleLastName(person.FirstName, "", person.LastName)
}

func (person *Person) AssignFrom(source storage.Person) error {
}
```

For the later release that introduced `MiddleName` the code is very similar:

``` go
package v20130303


import storage "v20140404storage"

func (person *Person) AssignTo(dest storage.Person) error {
    dest.PopulateFromFirstMiddleLastName(person.FirstName, person.MiddleName, person.LastName)
}

func (person *Person) AssignFrom(source storage.Person) error {
}
```

### Version 2015-05-05 - Property Rename

The term `AlphaKey` was found to be confusing to users, so in this release of the API it is renamed to `SortKey` to better reflect its purpose of sorting names together (e.g. *McDonald* gets sorted as though spelt *MacDonald*).

``` go
package v20150505

type Person struct {
    Id         Guid
    LegalName  string
    FamilyName string
    KnownAs    string
    SortKey    string // *** Used to be AlphaKey ***
}
```

As expected the storage version is also regenerated:

``` go
package v20150505storage

type Person struct {
    PropertyBag
    Id         *Guid
    LegalName  *string
    FamilyName *string
    KnownAs    *string
    SortKey    *string // *** Used to be AlphaKey ***
}

// Hub marks this type as a conversion hub.
func (*Person) Hub() {}
```

By documenting the renames in the configuration of our code generator, this rename will be automatically handled within the `ConvertTo()` and `ConvertFrom()` methods, as shown here for the prior version of `Person`:

``` go
package v20140404

import storage "v20150505storage"

// ConvertTo converts this Person to the Hub version.
func (person *Person) ConvertToStorage(dest storage.Person) error {
    dest.SortKey = person.AlphaKey // *** Rename is automatically handled ***
    dest.FamilyName = person.FamilyName
    dest.Id = person.Id
    dest.KnownAs = person.KnownAs
    dest.LegalName = person.LegalName

    if assignable, ok := person.(AssignableWithPersonStorage); ok {
        assignable.AssignTo(dest)
    }

    return nil
}

// ConvertFrom converts from the Hub version to this version.
func (person *Person) ConvertFromStorage(source storage.Person) error {
    person.AlphaKey = source.SortKey // *** Rename is automatically handled ***
    person.FamilyName = source.FamilyName
    person.Id = source.Id
    person.KnownAs = source.KnownAs"
    person.LegalName = source.LegalName

    if assignable, ok := person.(AssignableWithPersonStorage); ok {
        assignable.AssignFrom(source)
    }

    return nil
}
```

*Statistics:* At the time of writing, there were nearly 60 cases of fields being renamed between versions; 17 of these involved changes to letter case alone. (Count is inexact because renaming was inferred from the similarity of names.)

### Version 2016-06-06 - Complex Properties

With some customers expressing a desire to send physical mail to their customers, this release extends the API to allow a mailing address to be optionally specified for each person.

``` go
package v20160606

type Address struct {
    Street string
    City   string
}

type Person struct {
    Id             Guid
    LegalName      string
    FamilyName     string
    KnownAs        string
    SortKey        string
    MailingAddress Address
}
```

We now have two structs that make up our storage version:

``` go
package v20160606storage

type Person struct {
    PropertyBag
    Id             *Guid
    LegalName      *string
    FamilyName     *string
    KnownAs        *string
    SortKey        *string
    MailingAddress *Address
}

type Address struct {
    PropertyBag
    Street *string
    City   *string
}

// Hub marks this type of Person as a conversion hub.
func (*Person) Hub() {}
```

The required `ConvertToStorage()` and `ConvertFromStorage()` methods get generated in the expected way:

``` go
package v20160606

import storage "v20160606storage"

// ConvertTo converts this Person to the Hub version.
func (person *Person) ConvertToStorage(dest storage.Person) error {
    dest.SortKey = person.AlphaKey
    dest.FamilyName = person.FamilyName
    dest.Id = person.Id
    dest.KnownAs = person.KnownAs
    dest.LegalName = person.LegalName

    // *** Copy the mailing address over too ***
    address := &storage.Address{}
    err := person.MailingAddress.ConvertToStorage(address)
    if err != nil {
        return err
    }

    dest.MailingAddress = address

    if assignable, ok := person.(AssignableWithPersonStorage); ok {
        err := assignable.AssignTo(dest)
        if err != nill {
            return err
        }
    }

    return nil
}

// ConvertToStorage converts this Address to the hub storage version
func (address *Address) ConvertToStorage(dest storage.Address) error {
    dest.Street = address.Street
    dest.City = address.City

    if assignable, ok := person.(AssignableWithAddressStorage); ok {
        err := assignable.AssignTo(dest)
        if err != nill {
            return err
        }
    }

    return nil
}

// ConvertFrom converts from the Hub version to this version.
func (person *Person) ConvertFromStorage(source storage.Person) error {
    person.AlphaKey = source.SortKey // *** Rename is automatically handled ***
    person.FamilyName = source.FamilyName
    person.Id = source.Id
    person.KnownAs = source.KnownAs"
    person.LegalName = source.LegalName

    // *** Copy the mailing address over too ***
    if storage.MailingAddress != nil {
        address := &Address{}
        err := address.ConvertFromStorage(storage.Address)
        person.MailingAddress = address
    }

    if assignable, ok := person.(AssignableWithPersonStorage); ok {
        err := assignable.AssignFrom(source)
        if err != nill {
            return err
        }
    }

    return nil
}

// ConvertFromStorage converts from the hub storage version to this version
func (address *Address) ConvertFromStorage(source storage.Address) error {
    address.Street = source.Street
    address.City = source.City

    if assignable, ok := person.(AssignableWithAddressStorage); ok {
        err := assignable.AssignFrom(source)
        if err != nill {
            return err
        }
    }

    return nil
}
```


We're recursively applying the same conversion pattern to `Address` as we have already been using for `Person`. This scales to any level of nesting without the code becoming unweildy.

### Version 2017-07-07 - Optionality changes

TODO: Illustrate the changes if address is no longer mandatory and discuss what happens the other way around



Properties are always optional on the storage version, so the code in `ConvertFromStorage()` needs to handle **nil**. In this example, `MailingAddress` is required in the API version of `Person`, so we don't need such a check. However, if it became optional, we would then need to generate the check.


Field became optional: 100 times
Field becamse required: 99 times


### Version 2018-08-08 - Extending nested properties

TODO: Illustrate how additional properties get added to Address

### Version 2019-09-09 - Changing types

TODO: Illustrate how we handle the case when the property type changes

Field type changed: 160 times



### TODO

Scenarios still to add to the case study


https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definition-versioning/







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

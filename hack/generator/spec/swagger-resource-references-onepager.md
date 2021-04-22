# The feature ask
My team would like the ability to look at a particular field in the Swagger and know whether that field is a resource reference (a reference to an ARM ID). Our understanding is that today that is not possible. Some examples of references in existing specs:
1. [Compute](https://github.com/Azure/azure-rest-api-specs/blob/master/specification/compute/resource-manager/Microsoft.Compute/stable/2021-03-01/cloudService.json#L1745)
2. [Networking](https://github.com/Azure/azure-rest-api-specs/blob/master/specification/network/resource-manager/Microsoft.Network/stable/2020-07-01/network.json#L171)
3. [Batch](https://github.com/Azure/azure-rest-api-specs/blob/master/specification/batch/resource-manager/Microsoft.Batch/stable/2021-01-01/BatchManagement.json#L2324)

Not every team handles ARM references the same way (which is fine), but that means that as a client of the Swagger it is currently impossible to tell if a given field is a reference or not.

## Digging in

We want to add support for a new format: `arm-id` for documenting that a field is an ARM ID.

Optionally, if you want to document that a specific kind of resource is being referred to, you can use the extension:
```
"x-ms-allowed-arm-resource-types": {
  "allowedTypes": [
    "/subscriptions/{}/resourceGroups/{}/providers/{resourceProviderNamespace}/{resourceType}",
    "/subscriptions/{}/providers/{resourceProviderNamespace}/{resourceType}",
  ]
}
```
The `{resourceProviderNamespace}` and `{resourceType}` must be provided.

## Examples

### Basic
```
"MyExampleType": {
  "properties": {
    "id": {
      "type": "string",
      "format": "arm-id"
    }
  }
}
```

### Specific kind of reference
```
"MyExampleType": {
  "properties": {
    "vnetId": {
      "type": "string",
      "format": "arm-id",
      "x-ms-allowed-arm-resource-types": {
        "allowedTypes": [
          "/subscriptions/{}/resourceGroups/{}/providers/Microsoft.Network/virtualNetworks"
        ]
      }
    }
  }
}
```

### Referencing the type in common-types (preferred)
```
"MyExampleType": {
  "properties": {
    "id": {
      "$ref": "../../../../../common-types/resource-management/v2/types.json#/definitions/ArmId",
    }
  }
}
```

### Referencing a shared type in your spec
```
"MyExampleType": {
  "properties": {
    "vnetId": {
      "$ref": "#/definitions/VNetId",
    }
  }
},
"VNetId": {
  "type": "string",
  "format": "arm-id",
  "x-ms-allowed-arm-resource-types": {
    "allowedTypes": [
      "/subscriptions/{}/resourceGroups/{}/providers/Microsoft.Network/virtualNetworks"
    ]
  }
}
```

## What is Azure Service Operator (ASO)?
It's a Kubernetes operator which supports creating Azure resources. You can read more [here](https://github.com/Azure/azure-service-operator). It allows customers to create Azure resources directly in their Kubernetes cluster by defining YAML files representing the Azure resource and then doing `kubectl apply -f <my-yaml>`.

## How do resource references work in ASO?
ASO allows you to represent various Azure resources in Kubernetes. As a consequence of the fact that the resource is reflected in Kubernetes, you also may _refer_ to resources in two different ways:
1. A reference to the resource via its ARM ID, the same as you would do in an ARM template or the SDK.
2. A reference to the resource via its Kubernetes name. This is more natural to Kubernetes developers and plays to Kubernetes strengths by allowing you to more easily apply the same specification in multiple places (generating distinct resources with distinct ARM IDs) without having to update the resource specification for each kind of deployment. An example of this would be creating the same resource in Kubernetes in two different namesapces which map to two different subscriptions in Azure - the "Kubernetes reference" looks the same but the ARM ID of the actual referenced resource is different.

You can read more about the references design [here](https://github.com/Azure/k8s-infra/blob/master/hack/generator/spec/type_references_and_ownership.md)

## Why do we want this ability?
In a lot of ways, the maintainers of ASO (that's my team and I) are in the same position as the SDK team: there are tons of Azure services and there's no way we can keep up with the pace of new API versions/services without some sort of automation helping us. We are currently most of the way through developing a code-generator that generates Kubernetes CRDs from the azure-rest-api-specs Swager documents - you can see this generator here: [k8sinfra](https://github.com/Azure/k8s-infra).

In order for the generator to generate the right shape for resources that have references, we need a way to determine that a particular field is a reference, so that we can replace the simple `string` field with something like:
```
type ResourceReference struct {
  // Group is the Kubernetes group of the resource.
  Group string `json:"group"`
  // Kind is the Kubernetes kind of the resource.
  Kind string `json:"kind"`
  // Namespace is the Kubernetes namespace of the resource.
  Namespace string `json:"namespace"`
  // Name is the Kubernetes name of the resource.
  Name string `json:"name"`

  /// +kubebuilder:validation:Pattern="^/subscriptions/(.+?)/resourcegroups/(.+?)/providers/(.+?)/((.+?)/(.+?))+$"
  // ARMID is a string of the form /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/{resourceProviderNamespace}/{resourceType}/{resourceName}.
  // ARMID is mutually exclusive with Group, Kind, Namespace and Name.
  ARMID string `json:"armId"`
}
```

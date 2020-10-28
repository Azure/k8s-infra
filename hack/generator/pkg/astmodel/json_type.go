/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

// APIExtensions is the package containing the type we use to
// represent arbitrary JSON fields.
const APIExtensions = "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

var JSONType = MakeTypeName(MakeExternalPackageReference(APIExtensions), "JSON")

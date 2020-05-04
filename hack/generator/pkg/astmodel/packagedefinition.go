/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

type PackageDefinition struct {
	PackageReference

	definitions []Definition
}

func NewPackageDefinition(reference PackageReference) *PackageDefinition {
	return &PackageDefinition{reference, nil}
}

func (pkgDef *PackageDefinition) AddDefinition(def Definition) {
	pkgDef.definitions = append(pkgDef.definitions, def)
}

func allocateTypeToFile(def Definition, filesToGenerate map[string][]Definition) string {

	var allocatedToFile string
	for fileName, fileDefs := range filesToGenerate {
		for _, fileDef := range fileDefs {
			if fileDef.Type().References(def.Reference()) {
				if allocatedToFile == "" {
					allocatedToFile = fileName
					break // only need to find at least one owner per file
				} else if allocatedToFile != fileName {
					// more than one owner... put it in its own file
					return def.FileNameHint()
				}
			}
		}
	}

	return allocatedToFile
}

func (pkgDef *PackageDefinition) EmitDefinitions(outputDir string) {

	// pull out all resources
	var resources []*StructDefinition
	var otherDefs []Definition

	for _, def := range pkgDef.definitions {
		if structDef, ok := def.(*StructDefinition); ok && structDef.IsResource() {
			resources = append(resources, structDef)
		} else {
			otherDefs = append(otherDefs, def)
		}
	}

	// initialize with 1 resource per file
	filesToGenerate := make(map[string][]Definition)
	for _, resource := range resources {
		filesToGenerate[resource.FileNameHint()] = []Definition{resource}
	}

	// allocate other types to these files
	for len(otherDefs) > 0 {
		// dequeue!
		otherDef := otherDefs[0]
		otherDefs = otherDefs[1:]

		allocateToFile := allocateTypeToFile(otherDef, filesToGenerate)

		if allocateToFile == "" {
			// couldn't find a file to put it in
			// see if any other types will reference it on a future round
			foundReference := false
			for _, def := range otherDefs {
				if def.Type().References(otherDef.Reference()) {
					foundReference = true
					break
				}
			}

			if !foundReference {
				// couldn't find any references, put it in its own file
				allocateToFile = otherDef.FileNameHint()
			}
		}

		if allocateToFile != "" {
			filesToGenerate[allocateToFile] = append(filesToGenerate[allocateToFile], otherDef)
		} else {
			// re-queue it for later, it will eventually be allocated
			otherDefs = append(otherDefs, otherDef)
		}
	}

	// emit each definition
	for fileName, defs := range filesToGenerate {
		genFile := NewFileDefinition(defs[0].Reference().PackageReference, defs...)
		outputFile := filepath.Join(outputDir, fileName+"_types.go")
		log.Printf("Writing '%s'\n", outputFile)
		genFile.SaveTo(outputFile)
	}

	pkgDef.emitGroupVersionFile(outputDir)
}

func (pkgDef *PackageDefinition) emitGroupVersionFile(outputDir string) {
	gvFile := filepath.Join(outputDir, "groupversion_info.go")

	// TODO: better way to do this?
	content := fmt.Sprintf(`
/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

// Package %[2]s contains API Schema definitions for the %[1]s %[2]s API group
// +kubebuilder:object:generate=true
// +groupName=%[1]s.infra.azure.com
package %[2]s

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "%[1]s.infra.azure.com", Version: "%[2]s"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme

	localSchemeBuilder = SchemeBuilder.SchemeBuilder
)`, pkgDef.GroupName(), pkgDef.PackageName())

	ioutil.WriteFile(gvFile, []byte(content), 0700)
}

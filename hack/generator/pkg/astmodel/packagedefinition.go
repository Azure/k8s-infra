/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"bytes"
	"io/ioutil"
	"log"
	"path/filepath"
	"text/template"
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

var groupVersionFileTemplate = template.Must(template.New("groupVersionFile").Parse(`
/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

// Package {{.PackageName}} contains API Schema definitions for the {{.GroupName}} {{.PackageName}} API group
// +kubebuilder:object:generate=true
// +groupName={{.GroupName}}.infra.azure.com
package {{.PackageName}}

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "{{.GroupName}}.infra.azure.com", Version: "{{.PackageName}}"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme

	localSchemeBuilder = SchemeBuilder.SchemeBuilder
)`))

func (pkgDef *PackageDefinition) emitGroupVersionFile(outputDir string) {
	buf := &bytes.Buffer{}
	groupVersionFileTemplate.Execute(buf, pkgDef)

	gvFile := filepath.Join(outputDir, "groupversion_info.go")
	ioutil.WriteFile(gvFile, buf.Bytes(), 0700)
}

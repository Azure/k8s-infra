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

func (pkgDef *PackageDefinition) EmitDefinitions(outputDir string) {

	resources, otherDefinitions := partitionDefinitions(pkgDef.definitions)

	// initialize with 1 resource per file
	filesToGenerate := make(map[string][]Definition)
	for _, resource := range resources {
		filesToGenerate[resource.FileNameHint()] = []Definition{resource}
	}

	allocateTypesToFiles(otherDefinitions, filesToGenerate)
	emitFiles(filesToGenerate, outputDir)
	emitGroupVersionFile(pkgDef, outputDir)
}

func emitFiles(filesToGenerate map[string][]Definition, outputDir string) {
	for fileName, defs := range filesToGenerate {
		genFile := NewFileDefinition(defs[0].Reference().PackageReference, defs...)
		outputFile := filepath.Join(outputDir, fileName+"_types.go")
		log.Printf("Writing '%s'\n", outputFile)
		genFile.Tidy()
		genFile.SaveTo(outputFile)
	}
}

func anyReferences(defs []Definition, t Type) bool {
	for _, def := range defs {
		if def.Type().References(t) {
			return true
		}
	}

	return false
}

func partitionDefinitions(definitions []Definition) (resourceStructs []*StructDefinition, otherDefinitions []Definition) {

	var resources []*StructDefinition
	var notResources []Definition

	for _, def := range definitions {
		if structDef, ok := def.(*StructDefinition); ok && structDef.IsResource() {
			resources = append(resources, structDef)
		} else {
			notResources = append(notResources, def)
		}
	}

	return resources, notResources
}

/*
TODO: Fix allocation to files so it is deterministic

Get all files referencing the type
If more than one, put it in its own file

Get all outstanding references to the type
If no outstanding references
	If referenced by just one file, put in that file
	otherwise put into its own file

But how to handle cycles?
*/

func allocateTypesToFiles(typesToAllocate []Definition, filesToGenerate map[string][]Definition) {
	// Keep track of how many types we've checked since we last allocated one
	// This lets us detect if/when we stall during allocation
	skippedTypes := 0

	// Limit to how many types we skip before taking remedial action
	// len(typesToAllocate) changes during execution, so we cache it
	skipLimit := len(typesToAllocate)

	for len(typesToAllocate) > 0 {
		// dequeue!
		typeToAllocate := typesToAllocate[0]
		typesToAllocate = typesToAllocate[1:]

		allocatedFileName := ""
		filesReferencingType := findFilesReferencingType(typeToAllocate, filesToGenerate)
		pendingReferences := anyReferences(typesToAllocate, typeToAllocate.Reference())

		if len(filesReferencingType) > 1 {
			// Referenced by more than one file, put it in its own file
			allocatedFileName = typeToAllocate.FileNameHint()
		}

		if len(filesReferencingType) == 1 && !pendingReferences {
			// Only referenced in one place, and not refeferenced anywhere else, put it in that file
			allocatedFileName = filesReferencingType[0]
		}

		if len(filesReferencingType) == 0 && !pendingReferences {
			// Not referenced by any file and not referenced elsewhere, put it in its own file
			allocatedFileName = typeToAllocate.FileNameHint()
		}

		if skippedTypes > skipLimit {
			// We've processed the entire queue without allocating any files, so we have a cycle of related types to allocate
			// None of these types are referenced by multiple files (they would already be allocated, per rule above)
			// So either they're referenced by one file, or none at all
			// Breaking the cycle requires allocating one of these types; we need to do this deterministicly
			// So we prefer allocating to an existing file if we can, and we prefer names earlier in the alphabet

			breakCycle := true
			for _, t := range typesToAllocate {
				refs := findFilesReferencingType(t, filesToGenerate)
				if len(refs) > len(filesReferencingType) {
					// Type 't' would go into an existing file; `typeToAllocate` wouldn't
					breakCycle = false
					break
				}

				if t.FileNameHint() < typeToAllocate.FileNameHint() {
					// Type `t` is closer to the start of the alphabet
					breakCycle = false
					break
				}
			}

			if breakCycle {
				allocatedFileName = typeToAllocate.FileNameHint()
			}
		}

		if allocatedFileName != "" {
			filesToGenerate[allocatedFileName] = append(filesToGenerate[allocatedFileName], typeToAllocate)
			skippedTypes = 0
			continue
		}

		// re-queue it for later, it will eventually be allocated
		typesToAllocate = append(typesToAllocate, typeToAllocate)
		skippedTypes++

	}
}

func findFilesReferencingType(def Definition, filesToGenerate map[string][]Definition) []string {

	var result []string
	for fileName, fileDefs := range filesToGenerate {
		if anyReferences(fileDefs, def.Reference()) {
			result = append(result, fileName)
		}
	}

	return result
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

func emitGroupVersionFile(pkgDef *PackageDefinition, outputDir string) {
	buf := &bytes.Buffer{}
	groupVersionFileTemplate.Execute(buf, pkgDef)

	gvFile := filepath.Join(outputDir, "groupversion_info.go")
	ioutil.WriteFile(gvFile, buf.Bytes(), 0700)
}

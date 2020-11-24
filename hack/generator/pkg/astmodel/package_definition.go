/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"bytes"
	"fmt"
	"io/ioutil"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// PackageDefinition is the definition of a package
type PackageDefinition struct {
	GroupName        string
	PackageName      string
	GeneratorVersion string
	definitions      Types
}

// NewPackageDefinition constructs a new package definition
func NewPackageDefinition(groupName string, packageName string, genVersion string) *PackageDefinition {
	return &PackageDefinition{groupName, packageName, genVersion, make(Types)}
}

func (pkgDef *PackageDefinition) Definitions() Types {
	return pkgDef.definitions
}

func (pkgDef *PackageDefinition) GetDefinition(typeName TypeName) (TypeDefinition, error) {
	for _, def := range pkgDef.definitions {
		if def.Name().Equals(typeName) {
			return def, nil
		}
	}

	return TypeDefinition{}, errors.Errorf("No error with name %v found", typeName)
}

// AddDefinition adds a Definition to the PackageDefinition
func (pkgDef *PackageDefinition) AddDefinition(def TypeDefinition) {
	pkgDef.definitions.Add(def)
}

// EmitDefinitions emits the PackageDefinition to an output directory
func (pkgDef *PackageDefinition) EmitDefinitions(outputDir string, generatedPackages map[PackageReference]*PackageDefinition) (int, error) {

	filesToGenerate := allocateTypesToFiles(pkgDef.definitions)

	err := pkgDef.emitFiles(filesToGenerate, outputDir, generatedPackages)

	if err != nil {
		return 0, err
	}

	err = emitGroupVersionFile(pkgDef, outputDir)
	if err != nil {
		return 0, err
	}

	return len(filesToGenerate), nil
}

// DefinitionCount returns the count of definitions that have been sorted into this package
func (pkgDef *PackageDefinition) DefinitionCount() int {
	return len(pkgDef.definitions)
}

func (pkgDef *PackageDefinition) emitFiles(filesToGenerate map[string][]TypeDefinition, outputDir string, generatedPackages map[PackageReference]*PackageDefinition) error {
	var errs []error
	for fileName, defs := range filesToGenerate {
		codeFilePath := filepath.Join(
			outputDir,
			fmt.Sprintf("%v_types%v.go", fileName, CodeGeneratedFileSuffix))

		err := pkgDef.writeCodeFile(codeFilePath, defs, generatedPackages)
		if err != nil {
			errs = append(errs, err)
		}

		testFilePath := filepath.Join(
			outputDir,
			fmt.Sprintf("%v_types%v_test.go", fileName, CodeGeneratedFileSuffix))

		err = pkgDef.writeTestFile(testFilePath, defs, generatedPackages)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return kerrors.NewAggregate(errs)
	}

	return nil
}

func (pkgDef *PackageDefinition) writeCodeFile(
	outputFile string,
	defs []TypeDefinition,
	packages map[PackageReference]*PackageDefinition) error {

	ref := defs[0].Name().PackageReference
	genFile := NewFileDefinition(ref, defs, packages)

	klog.V(5).Infof("Writing code file %q\n", outputFile)
	err := genFile.SaveToFile(outputFile)
	if err != nil {
		return errors.Wrapf(err, "error writing definitions to file %q", outputFile)
	}

	return nil
}

func (pkgDef *PackageDefinition) writeTestFile(
	outputFile string,
	defs []TypeDefinition,
	packages map[PackageReference]*PackageDefinition) error {

	// First check to see if we have test cases to write
	haveTestCases := false
	for _, def := range defs {
		if def.HasTestCases() {
			haveTestCases = true
			break
		}
	}

	if !haveTestCases {
		// No test
		return nil
	}

	ref := defs[0].Name().PackageReference
	genFile := NewTestFileDefinition(ref, defs, packages)

	klog.V(5).Infof("Writing test case file %q\n", outputFile)
	err := genFile.SaveToFile(outputFile)
	if err != nil {
		return errors.Wrapf(err, "error writing test cases to file %q", outputFile)
	}

	return nil
}

func allocateTypesToFiles(types Types) map[string][]TypeDefinition {
	graph := NewReferenceGraphWithResourcesAsRoots(types)

	type Root struct {
		depth int
		name  TypeName
	}

	// rootFor maps a type name to the best root for it
	rootFor := make(map[TypeName]Root)

	for root := range graph.roots {
		collected := make(ReachableTypes)
		graph.collectTypes(0, root, collected)
		for node, depth := range collected {
			if currentRoot, ok := rootFor[node]; ok {
				// use depth and then root name as a tiebreaker
				if depth < currentRoot.depth ||
					(depth == currentRoot.depth && root.Name() < currentRoot.name.Name()) {
					rootFor[node] = Root{depth, root}
				}
			} else {
				rootFor[node] = Root{depth, root}
			}
		}
	}

	filesToGenerate := make(map[string][]TypeDefinition)

	for _, def := range types {
		var fileName string
		if root, ok := rootFor[def.name]; ok {
			fileName = FileNameHint(root.name)
		} else {
			fileName = FileNameHint(def.name)
		}

		filesToGenerate[fileName] = append(filesToGenerate[fileName], def)
	}

	return filesToGenerate
}

var groupVersionFileTemplate = template.Must(template.New("groupVersionFile").Parse(`
/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

// Code generated by k8s-infra-gen. DO NOT EDIT.
// Generator version: {{.GeneratorVersion}}

// Package {{.PackageName}} contains API Schema definitions for the {{.GroupName}} {{.PackageName}} API group
// +kubebuilder:object:generate=true
// All object properties are optional by default, this will be overridden when needed:
// +kubebuilder:validation:Optional
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
)
`))

func emitGroupVersionFile(pkgDef *PackageDefinition, outputDir string) error {
	buf := &bytes.Buffer{}
	err := groupVersionFileTemplate.Execute(buf, pkgDef)
	if err != nil {
		return err
	}

	gvFile := filepath.Join(outputDir, "groupversion_info"+CodeGeneratedFileSuffix+".go")

	err = ioutil.WriteFile(gvFile, buf.Bytes(), 0600)
	if err != nil {
		return errors.Wrapf(err, "error writing group version file %q", gvFile)
	}

	return nil
}

package main

import (
	"fmt"
	"github.com/Azure/k8s-infra/swagger"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

// The input folder structure is as below
// The bash script that generates this ensures that there is only a single version
// spec folder for each resource type. It is most likely to be `stable`, but it could be
// `preview` if no `stable` version exists for that type
//
//  swagger-specs
//   |- top-level
//          |-service1 (e.g. `cdn` or `compute`)
//          |   |-common   (want these)
//          |   |-quickstart-templates
//          |   |-data-plane
//          |   |-resource-manager (we're only interested in the contents of this folder)
//          |       |- resource-type1 (e.g. `Microsoft.Compute`)
//          |       |    |- common
//          |       |    |   |- *.json (want these)
//          |       |    |- stable (NB - may preview if no stable)
//          |       |    |    |- 2018-10-01
//          |       |    |        |- *.json   (want these)
//          |       |- misc files (e.g. readme)
//           ...

func main() {
	fmt.Println("*******************************************")
	fmt.Println("  Processing ARM Specs ")
	fmt.Println("*******************************************")
	config := getARMConfig()
	paths := loadARMSwagger(config)
	writeOutput(paths, config, "./internal/pkg/expanders/swagger-armspecs.generated.go", "SwaggerAPISetARMResources")
	fmt.Println()
}

func loadARMSwagger(config *swagger.Config) []*swagger.Path {
	var paths []*swagger.Path
	serviceFileInfos, err := ioutil.ReadDir("swagger-specs")
	if err != nil {
		panic(err)
	}
	for _, serviceFileInfo := range serviceFileInfos {
		if serviceFileInfo.IsDir() && serviceFileInfo.Name() != "common-types" {
			fmt.Printf("Processing service folder: %s\n", serviceFileInfo.Name())
			resourceTypeFileInfos, err := ioutil.ReadDir(fmt.Sprintf("swagger-specs/%s/resource-manager", serviceFileInfo.Name()))
			if err != nil {
				continue // may just be data-plane folder
			}
			for _, resourceTypeFileInfo := range resourceTypeFileInfos {
				if resourceTypeFileInfo.IsDir() && resourceTypeFileInfo.Name() != "common" {
					swaggerPath := getFirstNonCommonPath(getFirstNonCommonPath(fmt.Sprintf("swagger-specs/%s/resource-manager/%s", serviceFileInfo.Name(), resourceTypeFileInfo.Name())))
					swaggerFileInfos, err := ioutil.ReadDir(swaggerPath)
					if err != nil {
						panic(err)
					}
					// Build up paths for all files in the folder to allow proper sorting
					folderPaths := []swagger.Path{}
					for _, swaggerFileInfo := range swaggerFileInfos {
						if !swaggerFileInfo.IsDir() && strings.HasSuffix(swaggerFileInfo.Name(), ".json") {
							fmt.Printf("\tprocessing %s/%s\n", swaggerPath, swaggerFileInfo.Name())
							doc := loadDoc(swaggerPath + "/" + swaggerFileInfo.Name())
							filePaths, err := swagger.GetPathsFromSwagger(doc, config, "")
							if err != nil {
								panic(err)
							}
							folderPaths = append(folderPaths, filePaths...)
						}
					}
					if len(folderPaths) > 0 {
						paths, err = swagger.MergeSwaggerPaths(paths, config, folderPaths, true, "")
						if err != nil {
							panic(err)
						}
					}
				}
			}
		}
	}
	return paths
}

// getARMConfig returns the config for ARM Swagger processing
func getARMConfig() *swagger.Config {
	config := &swagger.Config{
		Overrides: map[string]swagger.PathOverride{},
	}
	return config
}

func loadDoc(path string) *loads.Document {

	document, err := loads.Spec(path)
	if err != nil {
		log.Panicf("Error opening Swagger: %s", err)
	}

	document, err = document.Expanded(&spec.ExpandOptions{RelativeBase: path})
	if err != nil {
		log.Panicf("Error expanding Swagger: %s", err)
	}

	return document
}
func writeOutput(paths []*swagger.Path, config *swagger.Config, filename string, structName string) {
	_ = os.MkdirAll(filepath.Dir(filename), os.FileMode(0777))
	writer, err := os.Create(filename)
	if err != nil {
		panic(fmt.Errorf("Error opening file: %s", err))
	}
	defer func() {
		err := writer.Close()
		if err != nil {
			panic(fmt.Errorf("Failed to close output file: %s", err))
		}
	}()

	writeTemplate(writer, paths, config, structName)
}
func writeTemplate(w io.Writer, paths []*swagger.Path, config *swagger.Config, structName string) {

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
	}
	t := template.Must(template.New("code-gen").Funcs(funcMap).Parse(swagger.Tmpl))

	type Context struct {
		Paths      []*swagger.Path
		StructName string
	}

	context := Context{
		Paths:      paths,
		StructName: structName,
	}

	err := t.Execute(w, context)
	if err != nil {
		panic(err)
	}
}

// getFirstNonCommonPath returns the first subfolder under path that is not named 'common'
func getFirstNonCommonPath(path string) string {
	// get the first non `common` path
	subfolders, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, subpath := range subfolders {
		if subpath.IsDir() && subpath.Name() != "common" {
			return path + "/" + subpath.Name()
		}
	}
	panic(fmt.Errorf("No suitable path found"))
}

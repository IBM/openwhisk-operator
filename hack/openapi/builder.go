/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	// Modify this
	owv1 "github.com/ibm/cloud-functions-operator/pkg/apis/ibmcloud/v1alpha1"
	generated "github.com/ibm/cloud-functions-operator/pkg/openapi"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
	"k8s.io/kube-openapi/pkg/builder"
	"k8s.io/kube-openapi/pkg/common"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	defaultSwaggerFile      = "_build/swagger.json"
	defaultExampleDirectory = "docs/examples/"

	extensionGVK = "x-kubernetes-group-version-kind"
)

var (
	pkgs = []string{
		"com.github.ibm.openwhisk-operator.pkg.apis.",
		"com.github.ibm.cloud-operators.pkg.types.apis.",
		"com.github.ibm.cloud-operators.pkg.lib.",
	}
)

func main() {
	// Get the name of the generated swagger file from the args
	// if it exists; otherwise use the default file name.
	swaggerFilename := defaultSwaggerFile
	if len(os.Args) > 1 {
		swaggerFilename = os.Args[1]
	}

	definitionsFunc := generated.GetOpenAPIDefinitions
	definitions := definitionsFunc(func(name string) spec.Ref {
		parts := strings.Split(name, "/")
		return spec.MustCreateRef(fmt.Sprintf("#/definitions/%s.%s",
			common.EscapeJsonPointer(parts[len(parts)-2]),
			common.EscapeJsonPointer(parts[len(parts)-1])))
	})

	// Generate the definition names from the map keys returned
	// from GetOpenAPIDefinitions. Anonymous function returning empty
	// Ref is not used.
	var defNames []string
	for name := range definitions {
		defNames = append(defNames, name)
	}

	scheme := runtime.NewScheme()
	owv1.SchemeBuilder.AddToScheme(scheme)
	definitionsNamer := NewDefinitionNamer(scheme)

	// Create a minimal builder config, then call the builder with the definition names.
	config := createOpenAPIBuilderConfig(definitionsNamer)
	config.GetDefinitions = definitionsFunc
	swagger, serr := builder.BuildOpenAPIDefinitionsForResources(config, defNames...)
	if serr != nil {
		log.Fatalf("ERROR: %s", serr.Error())
	}

	// Marshal the swagger spec into JSON, then write it out.
	specBytes, err := json.MarshalIndent(swagger, " ", " ")
	if err != nil {
		panic(fmt.Sprintf("json marshal error: %s", err.Error()))
	}

	os.MkdirAll(filepath.Dir(swaggerFilename), os.ModePerm)
	err = ioutil.WriteFile(swaggerFilename, specBytes, 0644)
	if err != nil {
		log.Fatalf("stdout write error: %s", err.Error())
	}

	// Generate swagger for documentation by trimming common prefix

	// trim definition names
	for name, def := range swagger.Definitions {
		trimmed := name
		for _, prefix := range pkgs {
			trimmed = strings.Replace(trimmed, prefix, "", 1)
		}
		if trimmed != name {
			delete(swagger.Definitions, name)
			swagger.Definitions[trimmed] = def
		}

	}

	// add examples
	if err := addExamples(swagger); err != nil {
		log.Fatalf("ERROR: %s", err.Error())
	}

	specBytes, err = json.MarshalIndent(swagger, " ", " ")
	if err != nil {
		panic(fmt.Sprintf("json marshal error: %s", err.Error()))
	}

	swaggerDocFilename := swaggerFilename[:len(swaggerFilename)-5] + "-doc.json"
	err = ioutil.WriteFile(swaggerDocFilename, specBytes, 0644)
	if err != nil {
		log.Fatalf("stdout write error: %s", err.Error())
	}

}

// CreateOpenAPIBuilderConfig hard-codes some values in the API builder
// config for testing.
func createOpenAPIBuilderConfig(definitionNamer *DefinitionNamer) *common.Config {
	return &common.Config{
		ProtocolList:   []string{"https"},
		IgnorePrefixes: []string{"/swaggerapi"},
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title:   "Kubernetes Apache Openwhisk Operator",
				Version: "1.0",
				Description: `Collection of Kubernetes operators for
 managing Apache Openwhisk resources, such as actions, packages, triggers and rules.`,
			},
		},
		GetDefinitionName: definitionNamer.GetDefinitionName,
	}
}

func addExamples(swagger *spec.Swagger) error {
	return filepath.Walk(defaultExampleDirectory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		var example interface{}
		err = yaml.Unmarshal(bytes, &example)
		if err != nil {
			return err
		}

		key := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		def, ok := swagger.Definitions[key]
		if !ok {
			log.Printf("Warning: definition %s does not exist", key)
		}
		def.Example = example
		swagger.Definitions[key] = def

		return nil
	})
}

// Below: Verbatim code from kubernetes/staging/src/k8s.io/apiserver/pkg/endpoints/openapi/openapi.go

type groupVersionKinds []v1.GroupVersionKind

func (s groupVersionKinds) Len() int {
	return len(s)
}

func (s groupVersionKinds) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s groupVersionKinds) Less(i, j int) bool {
	if s[i].Group == s[j].Group {
		if s[i].Version == s[j].Version {
			return s[i].Kind < s[j].Kind
		}
		return s[i].Version < s[j].Version
	}
	return s[i].Group < s[j].Group
}

// DefinitionNamer is the type to customize OpenAPI definition name.
type DefinitionNamer struct {
	typeGroupVersionKinds map[string]groupVersionKinds
}

func gvkConvert(gvk schema.GroupVersionKind) v1.GroupVersionKind {
	return v1.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}
}

func friendlyName(name string) string {
	nameParts := strings.Split(name, "/")
	// Reverse first part. e.g., io.k8s... instead of k8s.io...
	if len(nameParts) > 0 && strings.Contains(nameParts[0], ".") {
		parts := strings.Split(nameParts[0], ".")
		for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
			parts[i], parts[j] = parts[j], parts[i]
		}
		nameParts[0] = strings.Join(parts, ".")
	}
	friendly := strings.Join(nameParts, ".")
	if index := strings.Index(friendly, ".pkg."); index != -1 {
		friendly = friendly[index+5:]
	}
	if index := strings.Index(friendly, "apis."); index != -1 {
		friendly = friendly[index+5:]
	}
	return friendly
}

func typeName(t reflect.Type) string {
	path := t.PkgPath()
	if strings.Contains(path, "/vendor/") {
		path = path[strings.Index(path, "/vendor/")+len("/vendor/"):]
	}
	return fmt.Sprintf("%s.%s", path, t.Name())
}

// NewDefinitionNamer constructs a new DefinitionNamer to be used to customize OpenAPI spec.
func NewDefinitionNamer(schemes ...*runtime.Scheme) *DefinitionNamer {
	ret := &DefinitionNamer{
		typeGroupVersionKinds: map[string]groupVersionKinds{},
	}
	for _, s := range schemes {
		for gvk, rtype := range s.AllKnownTypes() {
			newGVK := gvkConvert(gvk)
			exists := false
			for _, existingGVK := range ret.typeGroupVersionKinds[typeName(rtype)] {
				if newGVK == existingGVK {
					exists = true
					break
				}
			}
			if !exists {
				ret.typeGroupVersionKinds[typeName(rtype)] = append(ret.typeGroupVersionKinds[typeName(rtype)], newGVK)
			}
		}
	}
	for _, gvk := range ret.typeGroupVersionKinds {
		sort.Sort(gvk)
	}
	return ret
}

// GetDefinitionName returns the name and tags for a given definition
func (d *DefinitionNamer) GetDefinitionName(name string) (string, spec.Extensions) {
	if groupVersionKinds, ok := d.typeGroupVersionKinds[name]; ok {
		return friendlyName(name), spec.Extensions{
			extensionGVK: []v1.GroupVersionKind(groupVersionKinds),
		}
	}
	return friendlyName(name), nil
}

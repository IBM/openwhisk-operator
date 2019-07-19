/*
 * Copyright 2017-2018 IBM Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package test

import (
	"io/ioutil"
	"time"

	yaml2 "github.com/ghodss/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	owv1 "github.com/ibm/cloud-functions-operator/pkg/apis/ibmcloud/v1alpha1"

	rcontext "github.com/ibm/cloud-operators/pkg/context"
)

// PostFunction creates a Function object
func PostFunction(context rcontext.Context, name string, spec owv1.FunctionSpec, async bool) runtime.Object {
	obj := makeFunction(context.Namespace(), name, spec)
	return post(context, &obj, async, 0)
}

// PostPackage creates a Package object
func PostPackage(context rcontext.Context, name string, spec owv1.PackageSpec, async bool) runtime.Object {
	obj := makePackage(context.Namespace(), name, spec)
	return post(context, &obj, async, 0)
}

// PostInvocation creates a Function object
func PostInvocation(context rcontext.Context, name string, spec owv1.InvocationSpec, async bool) runtime.Object {
	obj := makeInvocation(context.Namespace(), name, spec)
	return post(context, &obj, async, 0)
}

func makeFunction(namespace string, name string, spec owv1.FunctionSpec) owv1.Function {
	return owv1.Function{
		TypeMeta: metav1.TypeMeta{
			APIVersion: owv1.SchemeGroupVersion.Group + "/" + owv1.SchemeGroupVersion.Version,
			Kind:       "Function",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
}

func makePackage(namespace string, name string, spec owv1.PackageSpec) owv1.Package {
	return owv1.Package{
		TypeMeta: metav1.TypeMeta{
			APIVersion: owv1.SchemeGroupVersion.Group + "/" + owv1.SchemeGroupVersion.Version,
			Kind:       "Package",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
}

func makeInvocation(namespace string, name string, spec owv1.InvocationSpec) owv1.Invocation {
	return owv1.Invocation{
		TypeMeta: metav1.TypeMeta{
			APIVersion: owv1.SchemeGroupVersion.Group + "/" + owv1.SchemeGroupVersion.Version,
			Kind:       "Invocation",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
}

// PostInNs the object
func PostInNs(context rcontext.Context, obj runtime.Object, async bool, delay time.Duration) runtime.Object {
	obj.(metav1.ObjectMetaAccessor).GetObjectMeta().SetNamespace(context.Namespace())
	return post(context, obj, async, delay)
}

// Post the object
func post(context rcontext.Context, obj runtime.Object, async bool, delay time.Duration) runtime.Object {
	done := make(chan bool)

	go func() {
		if delay > 0 {
			time.Sleep(delay)
		}
		err := context.Client().Create(context, obj)
		if err != nil {
			panic(err)
		}
		done <- true
	}()

	if !async {
		<-done
	}
	return obj
}

func deleteObject(context rcontext.Context, obj runtime.Object, async bool) {
	done := make(chan bool)

	go func() {
		err := context.Client().Delete(context, obj)
		if err != nil {
			panic(err)
		}
		done <- true
	}()

	if !async {
		<-done
	}
}

// LoadFunction loads the YAML spec into obj
func LoadFunction(filename string) owv1.Function {
	return *LoadObject(filename, &owv1.Function{}).(*owv1.Function)
}

// LoadTrigger loads the YAML spec into obj
func LoadTrigger(filename string) owv1.Trigger {
	return *LoadObject(filename, &owv1.Trigger{}).(*owv1.Trigger)
}

// LoadPackage loads the YAML spec into obj
func LoadPackage(filename string) owv1.Package {
	return *LoadObject(filename, &owv1.Package{}).(*owv1.Package)
}

// LoadRule loads the YAML spec into obj
func LoadRule(filename string) owv1.Rule {
	return *LoadObject(filename, &owv1.Rule{}).(*owv1.Rule)
}

// LoadInvocation loads the YAML spec into obj
func LoadInvocation(filename string) owv1.Invocation {
	return *LoadObject(filename, &owv1.Invocation{}).(*owv1.Invocation)
}

// LoadObject loads the YAML spec into obj
func LoadObject(filename string, obj runtime.Object) runtime.Object {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	yaml2.Unmarshal(bytes, obj)
	return obj
}

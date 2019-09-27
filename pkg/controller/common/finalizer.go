/*
 * Copyright 2019 IBM Corporation
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

package common

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ibm/cloud-functions-operator/pkg/injection"
)

var flog = log.Log

// EnsureFinalizerAndPut makes sure obj contains finalizer. If not update obj and server
func EnsureFinalizerAndPut(ctx context.Context, client client.Client, obj runtime.Object, finalizer string) error {
	meta := ObjectMeta(obj)
	log := flog.WithValues("Namespace", meta.GetNamespace(), "Name", meta.GetName())

	if !HasFinalizer(obj, finalizer) {
		addFinalizer(obj, finalizer)
		if err := injection.GetKubeClient(ctx).Update(ctx, obj); err != nil {
			log.Info("error setting finalizer", "error", err)
			return err
		}
	}
	return nil
}

// RemoveFinalizerAndPut removes finalizer from obj (if present). Update obj and server when needed
func RemoveFinalizerAndPut(ctx context.Context, obj runtime.Object, finalizer string) error {
	meta := ObjectMeta(obj)
	log := flog.WithValues("Namespace", meta.GetNamespace(), "Name", meta.GetName())

	if HasFinalizer(obj, finalizer) {
		RemoveFinalizer(obj, finalizer)
		if err := injection.GetKubeClient(ctx).Update(ctx, obj); err != nil {
			log.Info("error setting finalizer", "error", err)
			return err
		}
	}
	return nil
}

// ObjectMeta gets the resource ObjectMeta field
func ObjectMeta(obj runtime.Object) metav1.Object {
	return obj.(metav1.ObjectMetaAccessor).GetObjectMeta()
}

// EnsureFinalizer makes sure the given object has the given finalizer.
// Return true if finalizer has been added
func EnsureFinalizer(obj runtime.Object, name string) bool {
	if HasFinalizer(obj, name) {
		return false
	}
	addFinalizer(obj, name)
	return true
}

func addFinalizer(obj runtime.Object, name string) runtime.Object {
	ObjectMeta(obj).SetFinalizers(append(ObjectMeta(obj).GetFinalizers(), name))
	return obj
}

// RemoveFinalizer clears the given finalizer from the list of the obj finalizers.
// Return true if finalizer has been removed
func RemoveFinalizer(obj runtime.Object, name string) bool {
	if !HasFinalizer(obj, name) {
		return false
	}

	finalizers := make([]string, 0)
	for _, finalizer := range ObjectMeta(obj).GetFinalizers() {
		if finalizer != name {
			finalizers = append(finalizers, finalizer)
		}
	}

	ObjectMeta(obj).SetFinalizers(finalizers)
	return true
}

// HasFinalizer returns true if the resource has the given finalizer name
func HasFinalizer(obj runtime.Object, name string) bool {
	finalizers := ObjectMeta(obj).GetFinalizers()
	for _, finalizer := range finalizers {
		if finalizer == name {
			return true
		}
	}
	return false
}

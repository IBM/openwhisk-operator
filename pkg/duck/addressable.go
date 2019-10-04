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

package duck

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/ibm/cloud-functions-operator/pkg/apis/duck"
	"github.com/ibm/cloud-functions-operator/pkg/injection"
)

// GetAddressable retreives the addressable object from the k8s cluster
func GetAddressable(ctx context.Context, ref *v1.ObjectReference) (*duck.AddressableType, error) {
	namespace := injection.GetRequest(ctx).Namespace
	client := injection.GetKubeClient(ctx)

	// Resolve ref as Addressable
	u := unstructured.Unstructured{}

	gvk := schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind)
	u.SetGroupVersionKind(gvk)
	err := client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: ref.Name}, &u)
	if err != nil {
		return nil, fmt.Errorf("Object not found: %s/%s", namespace, ref.Name)
	}

	b, err := u.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("Invalid object: %s/%s", namespace, ref.Name)
	}

	addressable := duck.AddressableType{}
	err = json.Unmarshal(b, &addressable)
	if err != nil {
		return nil, fmt.Errorf("Invalid object (not addressable): %s/%s", namespace, ref.Name)
	}
	return &addressable, nil
}

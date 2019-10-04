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
	"fmt"

	"github.com/ibm/cloud-functions-operator/pkg/injection"
	corev1 "k8s.io/api/core/v1"
)

// ResolveURL gets the URL associated to the given object reference.
// Support:
// - Core k8s services
// - Knative addressable objects
func ResolveURL(ctx context.Context, ref *corev1.ObjectReference) (string, error) {
	if isCoreService(ref) {
		namespace := injection.GetRequest(ctx).Namespace
		return fmt.Sprintf("https://%s.%s.svc.cluster.local", ref.Name, namespace), nil
	}

	// Try Addressable
	addressable, err := GetAddressable(ctx, ref)
	if err != nil {
		return "", err
	}

	if addressable.Status.Address == nil || addressable.Status.Address.URL == "" {
		return "", nil
	}

	return addressable.Status.Address.URL, nil
}

func isCoreService(ref *corev1.ObjectReference) bool {
	return ref.APIVersion == "v1" && ref.Kind == "Service"
}

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

package resources

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"

	"github.com/ibm/cloud-functions-operator/pkg/injection"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var slog = log.Log.WithName("secret")

// GetSecret gets a kubernetes secret
func GetSecret(ctx context.Context, secretname string, fallback bool) (*v1.Secret, error) {
	request := injection.GetRequest(ctx)
	if request == nil {
		return nil, errors.New("missing request in context")
	}

	namespace := request.Namespace
	log := slog.WithName(fmt.Sprintf("%s/%s", namespace, secretname))
	log.V(5).Info("getting secret")

	cl := injection.GetKubeClient(ctx)
	if cl == nil {
		return nil, errors.New("missing client in context")
	}

	var secret v1.Secret
	if err := cl.Get(ctx, client.ObjectKey{Namespace: namespace, Name: secretname}, &secret); err != nil {
		if namespace != "default" && fallback {
			if err := cl.Get(ctx, client.ObjectKey{Namespace: "default", Name: secretname}, &secret); err != nil {
				log.V(5).Info("secret not found")
				return nil, err
			}
		} else {
			log.V(5).Info("secret not found")
			return nil, err
		}
	}
	log.V(5).Info("secret found")
	return &secret, nil
}

// HasSecret checks if a secret exists
func HasSecret(context context.Context, name string, fallback bool) bool {
	slog.Info("Checking secret %s exist", name)
	_, err := GetSecret(context, name, fallback)
	return err != nil
}

// GetSecretValue gets the value of a secret in the given namespace. If not found and fallback is true, check default namespace
func GetSecretValue(context context.Context, name string, key string, fallback bool) ([]byte, error) {
	secret, err := GetSecret(context, name, fallback)
	if err != nil {
		return nil, err
	}

	return secret.Data[key], nil
}

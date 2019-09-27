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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/ibm/cloud-functions-operator/pkg/injection"
)

var clog = log.Log.WithName("configmap")

// GetConfigMap gets the ubernetes configmap of the given name.
func GetConfigMap(ctx context.Context, cmname string, fallback bool) (*v1.ConfigMap, error) {
	request := injection.GetRequest(ctx)
	if request == nil {
		return nil, errors.New("missing request in context")
	}

	namespace := request.Namespace

	log := clog.WithName(fmt.Sprintf("%s/%s", namespace, cmname))
	log.V(5).Info("getting configmap")

	cl := injection.GetKubeClient(ctx)
	if cl == nil {
		return nil, errors.New("missing client in context")
	}

	var cm v1.ConfigMap
	if err := cl.Get(ctx, client.ObjectKey{Namespace: namespace, Name: cmname}, &cm); err != nil {
		if namespace != "default" && fallback {
			if err := cl.Get(ctx, client.ObjectKey{Namespace: "default", Name: cmname}, &cm); err != nil {
				log.V(5).Info("configmap not found")
				return nil, err
			}
		} else {
			log.V(5).Info("configmap not found")
			return nil, err
		}
	}
	log.V(5).Info("configmap found")
	return &cm, nil
}

// HasConfigMap checks if a configmap exists
func HasConfigMap(context context.Context, name string, fallback bool) bool {
	clog.Info("Checking configmap %s exists", name)
	_, err := GetConfigMap(context, name, fallback)
	return err != nil
}

// GetConfigMapValue gets the value of the configmap of the given name in the given namespace. If not found and fallback is true, check default namespace
func GetConfigMapValue(context context.Context, name string, key string, fallback bool) (string, error) {
	cm, err := GetConfigMap(context, name, fallback)
	if err != nil {
		return "", err
	}

	return cm.Data[key], nil
}

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

package injection

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// kubeClientKey associated with client.Client
type kubeClientKey struct{}

// WithKubeClient extends context with client
func WithKubeClient(parent context.Context, client client.Client) context.Context {
	return context.WithValue(parent, kubeClientKey{}, client)
}

// KubeClient return client value
func GetKubeClient(ctx context.Context) client.Client {
	return ctx.Value(kubeClientKey{}).(client.Client)
}

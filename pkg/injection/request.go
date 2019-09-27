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

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// requesttKey associated with reconcile.Request
type requesttKey struct{}

// WithRequest extends context with reconcile.Request
func WithRequest(parent context.Context, request *reconcile.Request) context.Context {
	return context.WithValue(parent, requesttKey{}, request)
}

// GetRequest return client value
func GetRequest(ctx context.Context) *reconcile.Request {
	return ctx.Value(requesttKey{}).(*reconcile.Request)
}

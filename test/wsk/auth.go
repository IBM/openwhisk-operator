/*

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

package wsk

import "fmt"

// NewAuthResponse creates an auth response
func NewAuthResponse(org, space string) *Response {
	ns := fmt.Sprintf(`{"subject":"me","namespaces": [{"name": "%s_%s", "key":"akey", "uuid":"auuid"}]}`, org, space)
	return &Response{
		Path:   "/bluemix/v2/authenticate",
		Method: "GET",
		Body:   []byte(ns),
	}
}

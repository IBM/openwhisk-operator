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

package v1beta1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resv1 "github.com/ibm/cloud-operators/pkg/types/apis/resource/v1"
)

// Composition is the Schema for the compositions API
// +k8s:openapi-gen=true
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Composition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CompositionSpec   `json:"spec,omitempty"`
	Status CompositionStatus `json:"status,omitempty"`
}

// CompositionList contains a list of Composition
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CompositionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Composition `json:"items"`
}

// CompositionSpec defines the desired state of Composition
type CompositionSpec struct {
	// Composition name. Override metadata.name. Does not include the package name (see below)
	// +optional
	Name *string `json:"name,omitempty"`
	// Composition package name. Add it to the default package when not specified
	// +optional
	Package *string `json:"package,omitempty"`
	// The location of the composition to deploy. Support `http(s)` and `file` protocols.
	// +optional
	CompositionURI *string `json:"compositionURI,omitempty"`
	// The inline composition to deploy.
	// +optional
	Composition *string `json:"composition,omitempty"`

	// Reference to a secret representing where to deploy this entity
	// Default is `seed-default-owprops`
	// The secret must defines these fields:
	// apihost (string) : The OpenWhisk host
	// auth (string): the authorization key
	// cert (string):  the client certificate (optional)
	// insecure (bool):  Whether or not to bypass certificate checking (optional, default is false)
	// +optional
	ContextFrom *v1.SecretEnvSource `json:"contextFrom,omitempty"`
}

// CompositionStatus defines the observed state of Composition
type CompositionStatus struct {
	resv1.ResourceStatus `json:",inline"`

	// Last synced generation. Set by the system
	// +optional
	Generation int64 `json:"generation"`
}

func init() {
	SchemeBuilder.Register(&Composition{}, &CompositionList{})
}

// GetStatus returns the function status
func (r *Composition) GetStatus() resv1.Status {
	return &r.Status
}

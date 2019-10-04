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

package rule

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ibmcloudv1alpha1 "github.com/ibm/cloud-functions-operator/pkg/apis/ibmcloud/v1alpha1"
)

func RedirectFunctionName(rule *ibmcloudv1alpha1.Rule) string {
	return fmt.Sprintf("rule-%s-owc-redirect", rule.Name)
}

func NewRedirectFunction(rule *ibmcloudv1alpha1.Rule, redirectURL string) *ibmcloudv1alpha1.Function {
	code := fmt.Sprintf("const main = () => ({ headers: { location: '%s' }, statusCode: 302 })", redirectURL)
	return &ibmcloudv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RedirectFunctionName(rule),
			Namespace: rule.Namespace,
			// TODO OwnerRef
		},
		Spec: ibmcloudv1alpha1.FunctionSpec{
			Code:    &code,
			Runtime: "nodejs:8",
		},
	}
}

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

package function

import (
	"strings"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	Ω "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	owv1alpha1 "github.com/ibm/cloud-functions-operator/pkg/apis/ibmcloud/v1alpha1"
	"github.com/ibm/cloud-functions-operator/test"
	"github.com/ibm/cloud-functions-operator/test/wsk"
)

const (
	// Location of the testing data
	testdata = "../../../test/data/"
)

func TestFunction(t *testing.T) {
	owv1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)

	Ω.RegisterFailHandler(ginkgo.Fail)

	ginkgo.RunSpecs(t, "Function Suite")
}

var _ = ginkgo.Describe("Function", func() {

	table.DescribeTable("Reconcile",

		func(testCase TestCase) {
			server := wsk.NewServer()
			defer server.Close()

			initObjs := append(testCase.InitObjects, NewSeedSecret(server.URL))
			client := fake.NewFakeClient(initObjs...)
			reconciler := ReconcileFunction{Client: client, scheme: scheme.Scheme}

			testCase.Run(reconciler.Reconcile, server)
		},

		table.Entry("no object", TestCase{
			Request: reconcile.Request{NamespacedName: types.NamespacedName{Name: "doesnotexist"}},
		}),
		table.Entry("echo-null", TestCase{
			Name: "echo-null",
			InitObjects: []runtime.Object{
				test.LoadFunction(testdata + "owf-echo-null.yaml"),
			},
			Request:  reconcile.Request{NamespacedName: types.NamespacedName{Name: "echo-null"}},
			Expected: `{"name":"echo-null","exec":{"kind":"nodejs:6","code":"const main = params => params || {}"}}`,
		}),
	)
})

type ReconcileFn func(request reconcile.Request) (reconcile.Result, error)

type TestCase struct {
	Name        string
	InitObjects []runtime.Object
	Request     reconcile.Request
	Expected    string
}

func (t *TestCase) Run(reconcile ReconcileFn, server *wsk.Server) {
	_, err := reconcile(t.Request)

	Ω.Expect(err).ShouldNot(Ω.HaveOccurred())

	if t.Expected != "" {
		requests := server.Requests
		Ω.Expect(len(requests)).To(Ω.Equal(1))
		Ω.Expect(strings.TrimSpace(requests[0].Body)).To(Ω.Equal(t.Expected))
	}
}

func NewSeedSecret(url string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "seed-defaults-owprops",
		},
		Data: map[string][]byte{
			"apihost": []byte(url),
			"auth":    []byte("akey"),
		},
	}
}

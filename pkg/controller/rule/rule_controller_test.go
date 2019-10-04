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
	"context"
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
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func TestRule(t *testing.T) {
	owv1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)

	Ω.RegisterFailHandler(ginkgo.Fail)

	ginkgo.RunSpecs(t, "Rule Unit Test Suite")
}

var _ = ginkgo.Describe("Rule Unit Test", func() {

	table.DescribeTable("Reconcile",

		func(testCase TestCase) {
			server := wsk.NewServer()
			defer server.Close()

			initObjs := append(testCase.InitObjects, NewSeedSecret(server.URL))
			client := fake.NewFakeClient(initObjs...)
			reconciler := ReconcileRule{Client: client, scheme: scheme.Scheme}

			testCase.Run(reconciler.Reconcile, client, server)
		},

		table.Entry("no object", TestCase{
			Request: reconcile.Request{NamespacedName: types.NamespacedName{Name: "doesnotexist"}},
		}),
		table.Entry("with trigger and action", TestCase{
			InitObjects: []runtime.Object{
				test.LoadFunction(testdata + "owf-echo-null.yaml"),
				test.LoadRule(testdata + "owr-hello-location.yaml"),
			},
			Request:  reconcile.Request{NamespacedName: types.NamespacedName{Name: "hello-location"}},
			Expected: []string{`{"name":"hello-location","status":"","trigger":"/_/location-update-trigger","action":"/_/echo-null","publish":false}`},
		}),
		table.Entry("with trigger and object reference to addressable", TestCase{
			InitObjects: []runtime.Object{
				test.LoadUnstructured(testdata + "addressable-hello.yaml"),
				test.LoadRule(testdata + "owr-hello-location-with-addressable.yaml"),
			},
			Request: reconcile.Request{NamespacedName: types.NamespacedName{Name: "hello-location-with-addressable"}},
			Expected: []string{
				`{"name":"hello-location-with-addressable","status":"","trigger":"/_/location-update-trigger","action":"rule-hello-location-with-addressable-owc-redirect","publish":false}`,
			},
		}),
		table.Entry("with trigger and object reference to core service", TestCase{
			InitObjects: []runtime.Object{
				test.LoadUnstructured(testdata + "svc-hello.yaml"),
				test.LoadRule(testdata + "owr-hello-location-with-svc.yaml"),
			},
			Request: reconcile.Request{NamespacedName: types.NamespacedName{Name: "hello-location-with-svc"}},
			Expected: []string{
				`{"name":"hello-location-with-svc","status":"","trigger":"/_/location-update-trigger","action":"rule-hello-location-with-svc-owc-redirect","publish":false}`,
			},
		}),
	)
})

type ReconcileFn func(request reconcile.Request) (reconcile.Result, error)

type TestCase struct {
	InitObjects []runtime.Object
	Request     reconcile.Request
	Expected    []string
}

func (t *TestCase) Run(reconcile ReconcileFn, cl client.Client, server *wsk.Server) {
	_, err := reconcile(t.Request)

	Ω.Expect(err).ShouldNot(Ω.HaveOccurred())

	got := owv1alpha1.Rule{}
	cl.Get(context.TODO(), t.Request.NamespacedName, &got)

	if t.Expected != nil {
		requests := server.Requests
		Ω.Expect(len(requests)).To(Ω.Equal(len(t.Expected)))
		for i, e := range t.Expected {
			Ω.Expect(strings.TrimSpace(requests[i].Body)).To(Ω.Equal(e))
		}
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

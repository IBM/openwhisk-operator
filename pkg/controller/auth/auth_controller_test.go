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

package auth

import (
	"context"
	"log"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ibm/cloud-functions-operator/pkg/injection"
	owtest "github.com/ibm/cloud-functions-operator/test"
	"github.com/ibm/cloud-functions-operator/test/wsk"
)

var (
	c         client.Client
	cfg       *rest.Config
	namespace string
	ctx       context.Context
	t         *envtest.Environment
	stop      chan struct{}
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyPollingInterval(1 * time.Second)
	SetDefaultEventuallyTimeout(30 * time.Second)

	RunSpecs(t, "Auth Suite")
}

var _ = BeforeSuite(func() {
	// Start kube apiserver
	t = &envtest.Environment{
		ControlPlaneStartTimeout: 2 * time.Minute,
	}
	var err error
	if cfg, err = t.Start(); err != nil {
		log.Fatal(err)
	}

	// Setup the Manager and Controller.
	mgr, err := manager.New(cfg, manager.Options{})
	Expect(err).NotTo(HaveOccurred())
	c = mgr.GetClient()

	recFn := newReconciler(mgr)
	Expect(add(mgr, recFn)).NotTo(HaveOccurred())

	stop = owtest.StartTestManager(mgr)

	// Initialize objects
	org := "testorg"
	space := "testspace"

	wskserver := wsk.NewServer(wsk.NewAuthResponse(org, space))
	namespace = owtest.SetupKubeOrDie(cfg, "openwhisk-auth-", map[string]string{"org": org, "space": space, "apihost": wskserver.URL})

	ctx = injection.WithRequest(context.Background(), &reconcile.Request{NamespacedName: types.NamespacedName{Name: "", Namespace: namespace}})
	ctx = injection.WithKubeClient(ctx, c)
})

var _ = AfterSuite(func() {
	close(stop)
	t.Stop()
})

var _ = Describe("auth", func() {

	It("should be created", func() {
		secret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "seed-defaults" + SecretSuffix,
			},
		}
		Eventually(owtest.GetObject(ctx, &secret)).Should(And(Not(BeNil()), WithTransform(getData, And(HaveKey("auth"), HaveKey("apihost")))))
	})

})

func getData(secret *v1.Secret) map[string][]byte {
	if secret != nil {
		return secret.Data
	}
	return nil
}

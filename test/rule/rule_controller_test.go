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
	"log"
	"path/filepath"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	Ω "github.com/onsi/gomega"

	"github.com/apache/openwhisk-client-go/whisk"

	"github.com/ibm/cloud-functions-operator/pkg/apis"
	ow "github.com/ibm/cloud-functions-operator/pkg/controller/common"
	owf "github.com/ibm/cloud-functions-operator/pkg/controller/function"
	"github.com/ibm/cloud-functions-operator/pkg/controller/rule"
	owt "github.com/ibm/cloud-functions-operator/pkg/controller/trigger"
	"github.com/ibm/cloud-functions-operator/pkg/injection"
	"github.com/ibm/cloud-functions-operator/test"
)

var (
	c         client.Client
	cfg       *rest.Config
	namespace string
	ctx       context.Context
	wskclient *whisk.Client
	t         *envtest.Environment
	stop      chan struct{}
)

func TestRule(t *testing.T) {
	Ω.RegisterFailHandler(Fail)
	Ω.SetDefaultEventuallyPollingInterval(1 * time.Second)
	Ω.SetDefaultEventuallyTimeout(30 * time.Second)

	RunSpecs(t, "Rule Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(logf.ZapLoggerTo(GinkgoWriter, true))

	// Start kube apiserver
	t = &envtest.Environment{
		CRDDirectoryPaths:        []string{filepath.Join("..", "..", "config", "crds")},
		ControlPlaneStartTimeout: 2 * time.Minute,
	}
	apis.AddToScheme(scheme.Scheme)

	var err error
	if cfg, err = t.Start(); err != nil {
		log.Fatal(err)
	}

	// Setup the Manager and Controller.
	mgr, err := manager.New(cfg, manager.Options{})
	Ω.Expect(err).NotTo(Ω.HaveOccurred())

	c = mgr.GetClient()

	// Add reconcilers
	Ω.Expect(rule.Add(mgr)).NotTo(Ω.HaveOccurred())
	Ω.Expect(owt.Add(mgr)).NotTo(Ω.HaveOccurred())
	Ω.Expect(owf.Add(mgr)).NotTo(Ω.HaveOccurred())

	stop = test.StartTestManager(mgr)

	// Initialize objects
	namespace = test.SetupKubeOrDie(cfg, "openwhisk-rule-", nil)
	ctx = injection.WithRequest(context.Background(), &reconcile.Request{NamespacedName: types.NamespacedName{Name: "", Namespace: namespace}})
	ctx = injection.WithKubeClient(ctx, c)

	clientset := test.GetClientsetOrDie(cfg)
	test.ConfigureOwprops("seed-defaults-owprops", clientset.CoreV1().Secrets(namespace))

	wskclient, err = ow.NewWskClient(ctx, nil)
	Ω.Expect(err).NotTo(Ω.HaveOccurred())
})

var _ = AfterSuite(func() {
	close(stop)
	t.Stop()
})

var _ = Describe("rule", func() {

	DescribeTable("Firing Events",

		func(tc test.IntegrationCase) {
			tc.Init(ctx)
			tc.WaitOnline(ctx)

			_, err := test.Fire(wskclient, "location-update-trigger", map[string]string{"name": "John", "place": "ykt"})
			Ω.Expect(err).NotTo(Ω.HaveOccurred())
			Ω.Eventually(test.GetActivation(wskclient, "rule-hello", tc.StartTime)).
				Should(test.MatchResult(map[string]interface{}{"payload": "Hello, John from ykt"}))
		},

		Entry("rule linking a simple trigger and the hello function", test.IntegrationCase{
			Case: test.Case{
				InitObjects: []runtime.Object{
					test.LoadRule("testdata/owr-location-update-rule.yaml"),
					test.LoadFunction("testdata/owf-rule-hello.yaml"),
					test.LoadTrigger("testdata/owt-location-update-trigger.yaml"),
				}},
		}),
	)
})

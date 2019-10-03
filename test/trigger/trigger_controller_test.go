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

package trigger

import (
	"context"
	"log"
	"path/filepath"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
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
	. "github.com/onsi/gomega"

	"github.com/apache/openwhisk-client-go/whisk"

	resv1 "github.com/ibm/cloud-operators/pkg/lib/resource/v1"

	"github.com/ibm/cloud-functions-operator/pkg/apis"
	ow "github.com/ibm/cloud-functions-operator/pkg/controller/common"
	owfn "github.com/ibm/cloud-functions-operator/pkg/controller/function"
	owpkg "github.com/ibm/cloud-functions-operator/pkg/controller/pkg"
	owrule "github.com/ibm/cloud-functions-operator/pkg/controller/rule"
	"github.com/ibm/cloud-functions-operator/pkg/controller/trigger"
	"github.com/ibm/cloud-functions-operator/pkg/injection"
	owtest "github.com/ibm/cloud-functions-operator/test"
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

func TestTrigger(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyPollingInterval(1 * time.Second)
	SetDefaultEventuallyTimeout(30 * time.Second)

	RunSpecs(t, "Trigger Suite")
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
	Expect(err).NotTo(HaveOccurred())

	c = mgr.GetClient()

	// Add reconcilers
	Expect(trigger.Add(mgr)).NotTo(HaveOccurred())
	Expect(owrule.Add(mgr)).NotTo(HaveOccurred())
	Expect(owpkg.Add(mgr)).NotTo(HaveOccurred())
	Expect(owfn.Add(mgr)).NotTo(HaveOccurred())

	stop = owtest.StartTestManager(mgr)

	// Initialize objects
	namespace = owtest.SetupKubeOrDie(cfg, "openwhisk-trigger-", nil)
	ctx = injection.WithRequest(context.Background(), &reconcile.Request{NamespacedName: types.NamespacedName{Name: "", Namespace: namespace}})
	ctx = injection.WithKubeClient(ctx, c)

	clientset := owtest.GetClientsetOrDie(cfg)
	owtest.ConfigureOwprops("seed-defaults-owprops", clientset.CoreV1().Secrets(namespace))

	secret := owtest.LoadObject("testdata/secrets-kafka.yaml", &v1.Secret{})
	clientset.CoreV1().Secrets(namespace).Create(secret.(*v1.Secret))

	wskclient, err = ow.NewWskClient(ctx, nil)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	close(stop)
	t.Stop()
})

var _ = Describe("trigger", func() {

	DescribeTable("should be ready",
		func(specfile, pkgfile, rlfile string) {
			trigger := owtest.LoadTrigger("testdata/" + specfile)
			pkg := owtest.LoadPackage("testdata/" + pkgfile)
			rule := owtest.LoadRule("testdata/" + rlfile)

			owtest.PostInNs(ctx, &pkg, false, 0)
			ruleo := owtest.PostInNs(ctx, &rule, true, 0)
			obj := owtest.PostInNs(ctx, &trigger, false, 0)

			Eventually(owtest.GetState(ctx, obj)).Should(Equal(resv1.ResourceStateOnline))
			Eventually(owtest.GetState(ctx, ruleo)).Should(Equal(resv1.ResourceStateOnline))

			params := make(map[string]string)
			params["topic"] = "openwhisk-test-topic1"
			params["value"] = "a message from seed"

			Expect(owtest.ActionInvocation(wskclient, "trigger-kafka-binding/messageHubProduce", params)).Should(HaveKeyWithValue("success", true))
		},
		Entry("kafka", "owt-kafka.yaml", "owp-kafka.yaml", "owr-kafka.yaml"),
	)
})

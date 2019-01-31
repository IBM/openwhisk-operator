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

package composition

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	context "github.com/ibm/cloud-operators/pkg/context"
	ic "github.com/ibm/cloud-operators/pkg/lib/ibmcloud"
	resv1 "github.com/ibm/cloud-operators/pkg/types/apis/resource/v1"

	openwhiskv1beta1 "github.com/ibm/openwhisk-operator/pkg/apis/openwhisk/v1beta1"
	ow "github.com/ibm/openwhisk-operator/pkg/controller/common"
)

var clog = logf.Log

// Add creates a new Composition Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileComposition{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("composition-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Composition
	err = c.Watch(&source.Kind{Type: &openwhiskv1beta1.Composition{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileComposition{}

// ReconcileComposition reconciles a Composition object
type ReconcileComposition struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Composition object and makes changes based on the state read
// and what is in the Composition.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=openwhisk.seed.ibm.com,resources=compositions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openwhisk.seed.ibm.com,resources=compositions/status,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileComposition) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	context := context.New(r.Client, request)

	// Fetch the Function instance
	composition := &openwhiskv1beta1.Composition{}
	err := r.Get(context, request.NamespacedName, composition)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Reconcile or finalize?
	if composition.GetDeletionTimestamp() != nil {
		return r.finalize(context, composition)
	}

	log := clog.WithValues("namespace", composition.Namespace, "name", composition.Name)

	// Check generation
	currentGeneration := composition.Generation
	syncedGeneration := composition.Status.Generation
	if currentGeneration != 0 && syncedGeneration >= currentGeneration {
		// condition generation matches object generation. Nothing to do
		log.Info("composition up-to-date")
		return reconcile.Result{}, nil
	}

	// Check Finalizer is set
	if !resv1.HasFinalizer(composition, ow.Finalizer) {
		composition.SetFinalizers(append(composition.GetFinalizers(), ow.Finalizer))

		if err := r.Update(context, composition); err != nil {
			log.Info("setting finalizer failed. (retrying)", "error", err)
			return reconcile.Result{}, err
		}
	}

	// Make sure status is Pending
	if err := ow.SetStatusToPending(context, r.Client, composition, "deploying"); err != nil {
		return reconcile.Result{}, err
	}

	retry, err := r.updateComposition(context, composition)
	if err != nil {
		if !retry {
			log.Error(err, "deployment failed")

			// Non recoverable error.
			composition.Status.Generation = currentGeneration
			composition.Status.State = resv1.ResourceStateFailed
			composition.Status.Message = fmt.Sprintf("%v", err)
			if err := resv1.PutStatusAndEmit(context, composition); err != nil {
				log.Info("failed to set status. (retrying)", "error", err)
			}
			return reconcile.Result{}, nil
		}
		log.Error(err, "deployment failed (retrying)", "error", err)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r ReconcileComposition) updateComposition(context context.Context, obj *openwhiskv1beta1.Composition) (bool, error) {
	log := clog.WithValues("namespace", obj.Namespace, "name", obj.Name)

	spec := obj.Spec

	pkgName := "_"
	if spec.Package != nil {
		pkgName = *spec.Package
	}

	name := obj.Name
	if spec.Name != nil {
		name = *spec.Name
	}

	log.Info("deploying composition")

	if spec.CompositionURI != nil && spec.Composition != nil {
		return false, fmt.Errorf("both compositionURI and composition have been set")
	}

	if spec.CompositionURI == nil && spec.Composition == nil {
		return false, fmt.Errorf("missing compositionURI and composition")
	}

	var composition string
	var contentType = "text/plain"
	if spec.CompositionURI != nil {
		log.Info("downloading composition", "URI", *spec.CompositionURI)
		dat, erRead := ic.Read(context, *spec.CompositionURI)
		if erRead != nil {
			return false, fmt.Errorf("Error reading %s : %v", *spec.CompositionURI, erRead)
		}
		composition = string(dat)

		if strings.HasSuffix(*spec.CompositionURI, ".json") {
			contentType = "application/json"
		} else {
		}

	} else {
		contentType = "application/json"
		composition = *spec.Composition
	}

	client, err := ow.NewCompositionClient(context, spec.ContextFrom)
	if err != nil {
		return true, fmt.Errorf("Error creating composition service client: %v. (Retrying)", err)
	}

	log.Info("update composition using contentType", "content-type", contentType)

	_, _, err = client.Update(fmt.Sprintf("%s/%s", pkgName, name), composition, contentType)

	if err != nil {
		// TODO: recoverable vs non-recoverable errors

		return true, fmt.Errorf("error while deploying composition: %v. (Retyring)", err)
	}

	obj.Status.Generation = obj.Generation
	obj.Status.State = resv1.ResourceStateOnline
	obj.Status.Message = time.Now().Format(time.RFC850)

	return false, resv1.PutStatusAndEmit(context, obj)
}

func (r *ReconcileComposition) finalize(context context.Context, obj *openwhiskv1beta1.Composition) (reconcile.Result, error) {
	log := clog.WithValues("namespace", obj.Namespace, "name", obj.Name)

	spec := obj.Spec

	pkgName := "_"
	if spec.Package != nil {
		pkgName = *spec.Package
	}

	name := obj.Name
	if spec.Name != nil {
		name = *spec.Name
	}

	client, err := ow.NewCompositionClient(context, spec.ContextFrom)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("Error creating composition service client: %v. (Retrying)", err)
	}

	_, _, err = client.Delete(fmt.Sprintf("%s/%s", pkgName, name))

	if err != nil {
		log.Error(err, "(ignored)")
		// return fmt.Errorf("Error while deleting composition: %v. (Retyring)", err)
	}

	return reconcile.Result{}, resv1.RemoveFinalizerAndPut(context, obj, ow.Finalizer)
}

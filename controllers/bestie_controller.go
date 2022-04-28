/*
Copyright The L5 Operator Authors

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

package controllers

import (
	"context"
	"time"

	"github.com/opdev/l5-operator-demo/internal/util"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
	srv1 "github.com/opdev/l5-operator-demo/internal/sub_reconcilers"
)

// BestieReconciler reconciles a Bestie object.
type BestieReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pets.bestie.com,resources=besties,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pets.bestie.com,resources=besties/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pets.bestie.com,resources=besties/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=*
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=*
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=*
//+kubebuilder:rbac:groups="",resources=configmaps;endpoints;events;persistentvolumeclaims;pods;namespaces;secrets;serviceaccounts;services;services/finalizers,verbs=*
//+kubebuilder:rbac:groups=postgres-operator.crunchydata.com,resources=postgresclusters,verbs=*
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=*
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheuses;servicemonitors,verbs=*
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Bestie object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *BestieReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx, "Request.Namespace", req.Namespace, "Request.Name", req.Name)
	log.Info("Reconciling Bestie")

	// Fetch the Bestie instance
	log.Info("get latest bestie instance")
	bestie := &petsv1.Bestie{}
	err := r.Get(ctx, req.NamespacedName, bestie)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			log.Info("Bestie resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Bestie")
		return ctrl.Result{}, err
	}

	subReconcilerList := []srv1.Reconciler{
		srv1.NewPostgresClusterCRReconciler(r.Client, log, r.Scheme),
		srv1.NewDatabaseSeedJobReconciler(r.Client, log, r.Scheme),
		srv1.NewDeploymentReconciler(r.Client, log, r.Scheme),
		srv1.NewDeploymentSizeReconciler(r.Client, log, r.Scheme),
		srv1.NewDeploymentImageReconciler(r.Client, log, r.Scheme),
		srv1.NewServiceReconciler(r.Client, log, r.Scheme),
		srv1.NewHPAReconciler(r.Client, log, r.Scheme),
		srv1.NewRouteReconciler(r.Client, log, r.Scheme),
	}

	requeueResult := false
	requeueDelay := time.Duration(0)
	for _, subReconciler := range subReconcilerList {
		subResult, err := subReconciler.Reconcile(ctx, bestie.DeepCopy())
		if err != nil {
			log.Error(err, "re-queuing with error")
			return subResult, err
		}
		requeueResult = requeueResult || subResult.Requeue
		if requeueDelay < subResult.RequeueAfter {
			requeueDelay = subResult.RequeueAfter
		}
	}
	return ctrl.Result{Requeue: requeueResult, RequeueAfter: requeueDelay}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BestieReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&petsv1.Bestie{})
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.Service{})
	builder.Owns(&networkv1.Ingress{})
	builder.Owns(&autoscalingv1.HorizontalPodAutoscaler{})
	if util.IsRouteAPIAvailable() {
		builder.Owns(&routev1.Route{})
	}
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: 2})

	return builder.Complete(r)
}

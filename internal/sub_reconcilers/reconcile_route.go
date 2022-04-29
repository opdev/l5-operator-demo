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

package sub_reconcilers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
	"github.com/opdev/l5-operator-demo/internal/util"
)

type RouteReconciler struct {
	client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func NewRouteReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *RouteReconciler {
	return &RouteReconciler{
		client: client,
		Log:    log,
		Scheme: scheme,
	}
}

func (r *RouteReconciler) Reconcile(ctx context.Context, bestie *petsv1.Bestie) (ctrl.Result, error) {
	logInfo := types.NamespacedName{
		Namespace: bestie.Namespace,
		Name:      "RouteReconciler",
	}
	r.Log.WithValues("route", logInfo)
	r.Log.Info("deploy route or service if openshift or vanilla k8s")
	// If the cluster is OpenShift, add a route, else add an ingress.
	if util.IsRouteAPIAvailable() {
		// utilruntime.Must(routev1.AddToScheme(runtime.NewScheme()))
		route := &routev1.Route{}
		err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-route", Namespace: bestie.Namespace}, route)
		if err != nil && errors.IsNotFound(err) {
			r.Log.Info("Creating a new route for bestie")
			fileName := "config/resources/bestie-route.yaml"
			err := r.applyManifests(ctx, bestie, route, fileName)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("Error during Manifests apply - %w", err)
			}
		}
	} else {
		r.Log.Info("Creating an ingress for bestie")
		ingress := &networkv1.Ingress{}
		err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-ingress", Namespace: bestie.Namespace}, ingress)
		if err != nil && errors.IsNotFound(err) {
			r.Log.Info("Creating a new ingress for bestie")
			fileName := "config/resources/bestie-ingress.yaml"
			err = r.applyManifests(ctx, bestie, ingress, fileName)
			if err != nil {
				r.Log.Error(err, "Failed to get ingress.")
				return ctrl.Result{Requeue: true}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *RouteReconciler) applyManifests(ctx context.Context, bestie *petsv1.Bestie, obj client.Object, fileName string) error {
	err := util.ApplyManifests(r.client, r.Scheme, ctx, bestie, obj, fileName)
	if err != nil {
		return err
	}
	return err
}

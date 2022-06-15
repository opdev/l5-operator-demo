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

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opdev/l5-operator-demo/internal/util"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
)

type ServiceMonitorReconciler struct {
	client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func NewServiceMonitorReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *ServiceMonitorReconciler {
	return &ServiceMonitorReconciler{
		client: client,
		Log:    log,
		Scheme: scheme,
	}
}

func (r *ServiceMonitorReconciler) Reconcile(ctx context.Context, bestie *petsv1.Bestie) (ctrl.Result, error) {
	logInfo := types.NamespacedName{
		Namespace: bestie.Namespace,
		Name:      "ServiceMonitorReconciler",
	}
	r.Log.WithValues("servicemonitor", logInfo)
	// Reconcile servicemonitor.
	r.Log.Info("reconcile bestie servicemonitor if it does not exist")
	servicemonitorv1 := &monitoringv1.ServiceMonitor{}
	err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-servicemonitor", Namespace: bestie.Namespace}, servicemonitorv1)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("Creating a new servicemonitor for bestie")
			fileName := "config/resources/bestie-metrics-servicemonitor.yaml"
			err := r.applyManifests(ctx, bestie, servicemonitorv1, fileName)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("Error during Manifests apply - %w", err)
			}
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}
	return ctrl.Result{}, err
}

func (r *ServiceMonitorReconciler) applyManifests(ctx context.Context, bestie *petsv1.Bestie, obj client.Object, fileName string) error {
	err := util.ApplyManifests(r.client, r.Scheme, ctx, bestie, obj, fileName)
	if err != nil {
		return err
	}
	return err
}

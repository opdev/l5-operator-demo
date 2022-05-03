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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
	"github.com/opdev/l5-operator-demo/internal/util"
)

type ServiceReconciler struct {
	client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func NewServiceReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *ServiceReconciler {
	return &ServiceReconciler{
		client: client,
		Log:    log,
		Scheme: scheme,
	}
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, bestie *petsv1.Bestie) (ctrl.Result, error) {
	logInfo := types.NamespacedName{
		Namespace: bestie.Namespace,
		Name:      "ServiceReconciler",
	}
	r.Log.WithValues("service", logInfo)
	// Reconcile service.
	r.Log.Info("reconcile bestie service if it does not exist")
	svc := &corev1.Service{}
	err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-service", Namespace: bestie.Namespace}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("Creating a new service for bestie")
			fileName := "config/resources/bestie-svc.yaml"
			err := r.applyManifests(ctx, bestie, svc, fileName)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("Error during Manifests apply - %w", err)
			}
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}
	return ctrl.Result{}, err
}

func (r *ServiceReconciler) applyManifests(ctx context.Context, bestie *petsv1.Bestie, obj client.Object, fileName string) error {
	err := util.ApplyManifests(r.client, r.Scheme, ctx, bestie, obj, fileName)
	if err != nil {
		return err
	}
	return err
}

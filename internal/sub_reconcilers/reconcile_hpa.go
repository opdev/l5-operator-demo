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
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
	hpa "github.com/opdev/l5-operator-demo/internal/hpa"
)

type HPAReconciler struct {
	client k8sclient.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func NewHPAReconciler(client k8sclient.Client, log logr.Logger, scheme *runtime.Scheme) *HPAReconciler {
	return &HPAReconciler{
		client: client,
		Log:    log,
		Scheme: scheme,
	}
}

func (r *HPAReconciler) Reconcile(ctx context.Context, bestie *petsv1.Bestie) (ctrl.Result, error) {
	logInfo := types.NamespacedName{
		Namespace: bestie.Namespace,
		Name:      "HPAReconciler",
	}
	log := r.Log.WithValues("hpa", logInfo)
	// Reconciling HPA.
	// check if deployment exists
	bestieDeployment := &appsv1.Deployment{}
	err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-app", Namespace: bestie.Namespace}, bestieDeployment)
	if err != nil && errors.IsNotFound(err) {
		log.Info("No deployment yet, requeue for hpa creation")
	}
	log.Info("Reconcile hpa if it does not exist and maxreplicas is set")
	autoScaler := &autoscalingv2.HorizontalPodAutoscaler{}
	err = r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-hpa", Namespace: bestie.Namespace}, autoScaler)
	if err != nil && errors.IsNotFound(err) && bestie.Spec.MaxReplicas != nil {
		log.Info("Creating New HPA Instance")
		err = r.createHorizontalPodAutoscaler(ctx, *bestie.DeepCopy(), *bestieDeployment.DeepCopy())
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}
	err = r.updateHorizontalPodAutoScaler(ctx, *bestie.DeepCopy(), *autoScaler.DeepCopy())
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{}, nil
}

func (r *HPAReconciler) createHorizontalPodAutoscaler(ctx context.Context, bestie petsv1.Bestie, bestieDeployment appsv1.Deployment) error {
	r.Log.Info("MaxReplicas is set, enabling HPA")
	target := hpa.AutoScaler(r.Log, bestieDeployment, bestie)
	if err := controllerutil.SetControllerReference(&bestie, &target, r.Scheme); err != nil {
		return fmt.Errorf("failed to set the controller reference: %w", err)
	}
	err := r.client.Create(ctx, &target)
	if err != nil {
		r.Log.Error(err, "Failed to create HPA")
		return err
	}
	r.Log.Info("created", "HPA.Name", target.Name, "HPA.NameSpace", target.Namespace)
	return nil
}

func (r *HPAReconciler) updateHorizontalPodAutoScaler(ctx context.Context, bestie petsv1.Bestie, hpa autoscalingv2.HorizontalPodAutoscaler) error {
	hpa.Spec.MinReplicas = &bestie.Spec.Size
	err := r.client.Update(ctx, &hpa)
	if err != nil {
		r.Log.Error(err, "Failed to update HPA min replicas to match bestie deployment size")
		return err
	}
	return nil
}

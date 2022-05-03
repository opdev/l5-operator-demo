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
	"time"

	"github.com/opdev/l5-operator-demo/internal/bestie_errors"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
)

type DeploymentSizeReconciler struct {
	client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func NewDeploymentSizeReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *DeploymentSizeReconciler {
	return &DeploymentSizeReconciler{
		client: client,
		Log:    log,
		Scheme: scheme,
	}
}

func (r *DeploymentSizeReconciler) Reconcile(ctx context.Context, bestie *petsv1.Bestie) (ctrl.Result, error) {
	logInfo := types.NamespacedName{
		Namespace: bestie.Namespace,
		Name:      "DeploymentSizeReconciler",
	}
	log := r.Log.WithValues("deployment_size", logInfo)
	// Ensure the deployment size is the same as the spec
	log.Info("reconcile deployment to appropriate size if HPA is not enabled")
	bestieDeployment := &appsv1.Deployment{}
	HorizontalPodAutoScalar := &autoscalingv2.HorizontalPodAutoscaler{}
	// get latest instance of deployment
	err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-app", Namespace: bestie.Namespace}, bestieDeployment)
	if err != nil && errors.IsNotFound(err) {
		log.Info("No deployment initiated, deployment has not completed. Requeue.")
		delay := time.Second * time.Duration(5)
		log.Info(fmt.Sprintf("will retry after waiting for %s", delay))
		return ctrl.Result{RequeueAfter: delay}, nil
	}
	// Validate that MaxReplicas is greater than or equal to size.
	if bestie.Spec.MaxReplicas != nil && !(*bestie.Spec.MaxReplicas >= bestie.Spec.Size) {
		log.Error(bestie_errors.InvalidDeploymentSizeValue, "Invalid Deployment Size Value")
		return ctrl.Result{}, err
	}
	size := bestie.Spec.Size
	// TODO check if autoscaling is enabled in a better way ?
	err = r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-hpa", Namespace: bestie.Namespace}, HorizontalPodAutoScalar)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Horizontal pod autoscaler is not enabled proceeding with setting deployment to cr spec size")
		if *bestieDeployment.Spec.Replicas != size {
			*bestieDeployment.Spec.Replicas = size
			err = r.client.Update(ctx, bestieDeployment)
			if err != nil {
				r.Log.Error(err, "Failed to update Deployment", "Deployment.Namespace", bestieDeployment.Namespace, "Deployment.Name", bestieDeployment.Name, "Deployment.Spec", bestieDeployment.Spec)
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

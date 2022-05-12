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
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/opdev/l5-operator-demo/internal/util"

	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
)

type DeploymentReconciler struct {
	client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func NewDeploymentReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *DeploymentReconciler {
	return &DeploymentReconciler{
		client: client,
		Log:    log,
		Scheme: scheme,
	}
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, bestie *petsv1.Bestie) (ctrl.Result, error) {
	logInfo := types.NamespacedName{
		Namespace: bestie.Namespace,
		Name:      "DeploymentReconciler",
	}
	log := r.Log.WithValues("deployment", logInfo)

	// wait for db to be seeded
	job := &batchv1.Job{}
	err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-job", Namespace: bestie.Namespace}, job)
	if err != nil && errors.IsNotFound(err) {
		log.Info("No job initiated, job has not completed. Requeue to create deployment.")
		delay := time.Second * time.Duration(5)
		log.Info(fmt.Sprintf("will retry after waiting for %s", delay))
		return ctrl.Result{RequeueAfter: delay}, nil
	}
	if !(job.Status.Succeeded >= *job.Spec.Completions) {
		delay := time.Second * time.Duration(15)
		log.Info(fmt.Sprintf("database seeding job has not yet completed, waiting %s seconds", delay))
		return ctrl.Result{RequeueAfter: delay}, nil
	}
	_, err = r.updateDatabaseSeededCondition(ctx, *bestie)
	if err != nil {
		return ctrl.Result{}, err
	}
	// reconcile deployment.
	log.Info("reconcile deployment if it does not exist")
	dp := &appsv1.Deployment{}
	err = r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-app", Namespace: bestie.Namespace}, dp)
	if err != nil && errors.IsNotFound(err) && (job.Status.Succeeded >= *job.Spec.Completions) {
		log.Info("Creating a new app for bestie")
		fileName := "config/resources/bestie-deploy.yaml"
		err := r.applyDeploymentFromFile(ctx, bestie, *dp.DeepCopy(), fileName)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to apply deployment manifest - %w", err)
		}
		return r.createDeploymentCondition(ctx, *bestie)
	}
	return ctrl.Result{}, nil
}

func (r *DeploymentReconciler) applyDeploymentFromFile(ctx context.Context, bestie *petsv1.Bestie, obj appsv1.Deployment, fileName string) error {

	yaml, err := os.ReadFile(fileName)
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("Couldn't read manifest file for: %s", fileName))
		return err
	}
	if err = yamlutil.Unmarshal(yaml, &obj); err != nil {
		r.Log.Error(err, fmt.Sprintf("Couldn't unmarshall yaml file for: %s", fileName))
		return err
	}
	//apply values from the cr to the deployment object read from the template file
	obj.SetNamespace(bestie.Namespace)
	obj.Spec.Replicas = &bestie.Spec.Size
	containerPosition := 0
	for pos, container := range obj.Spec.Template.Spec.Containers {
		if container.Name == "bestie" {
			containerPosition = pos
		}
	}
	obj.Spec.Template.Spec.Containers[containerPosition].Image = fmt.Sprintf("%s:%s", bestie.Spec.Image, bestie.Spec.Version)
	err = controllerutil.SetControllerReference(bestie, &obj, r.Scheme)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference on the %s", obj.Name)
		return err
	}
	err = r.client.Create(ctx, &obj)
	if err != nil {
		r.Log.Error(err, "Failed to create object", "object", obj.GetName())
		return err
	}
	return nil
}

func (r *DeploymentReconciler) createDeploymentCondition(ctx context.Context, bestie petsv1.Bestie) (ctrl.Result, error) {
	err := util.RefreshCustomResource(ctx, r.client, &bestie)
	if err != nil {
		return ctrl.Result{}, err
	}
	meta.SetStatusCondition(&bestie.Status.Conditions, NewDeploymentCreatedCondition())
	err = r.client.Status().Update(ctx, &bestie)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *DeploymentReconciler) updateDatabaseSeededCondition(ctx context.Context, bestie petsv1.Bestie) (ctrl.Result, error) {
	err := util.RefreshCustomResource(ctx, r.client, &bestie)
	if err != nil {
		return ctrl.Result{}, err
	}
	databaseSeededCondition := NewDatabaseSeededCondition()
	databaseSeededCondition.Status = metav1.ConditionTrue
	databaseSeededCondition.Reason = "SeedJobCompleted"
	databaseSeededCondition.Message = "Database seed job completed successfully"
	meta.SetStatusCondition(&bestie.Status.Conditions, databaseSeededCondition)
	err = r.client.Status().Update(ctx, &bestie)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

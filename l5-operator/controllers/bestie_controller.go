/*
Copyright 2022.

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
	"fmt"
	"reflect"
	"time"

	//"time"
	pgov1 "github.com/crunchydata/postgres-operator/pkg/apis/postgres-operator.crunchydata.com/v1beta1"
	petsv1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = ctrllog.Log.WithName("controller_bestie")

// BestieReconciler reconciles a Bestie object
type BestieReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	BestieDefaultImage   = "quay.io/mkong/bestiev2"
	BestieDefaultVersion = "1.1"
	BestieName           = "bestie"
)

//+kubebuilder:rbac:groups=pets.bestie.com,resources=besties,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pets.bestie.com,resources=besties/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pets.bestie.com,resources=besties/finalizers,verbs=update

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
	log := ctrllog.FromContext(ctx)
	log.Info("Reconciling Bestie")

	// Fetch the Bestie instance
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

	// reconcile Postgres
	pgo := &pgov1.PostgresCluster{}

	err = r.Get(ctx, types.NamespacedName{Name: BestieName + "-pgo", Namespace: bestie.Namespace}, pgo)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating a new PGC for bestie")
			fileName := "config/resources/postgrescluster.yaml"
			r.applyManifests(ctx, req, bestie, pgo, fileName)
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}

	// reconcile Deployment
	dp := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: BestieName + "-app", Namespace: bestie.Namespace}, dp)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating a new app for bestie")
			fileName := "config/resources/bestie-deploy.yaml"
			r.applyManifests(ctx, req, bestie, dp, fileName)
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}

	//isAppRunning
	bestieRunning := r.isRunning(ctx, bestie)

	if !bestieRunning {
		// If bestie-app isn't running yet, requeue the reconcile
		// to run again after a delay
		delay := time.Second * time.Duration(15)

		log.Info(fmt.Sprintf("bestie-app is instantiating, waiting for %s", delay))
		return ctrl.Result{RequeueAfter: delay}, nil
	}

	//Level 2 : update Operand
	err = r.upgradeOperand(ctx, bestie)
	if err != nil {
		log.Error(err, "Failed to upgrade the operand")
		return ctrl.Result{Requeue: true}, err
	}

	//Level 1 : update appVersion status
	appVersion := r.reportappversion(bestie)
	if !reflect.DeepEqual(appVersion, bestie.Status.AppVersion) {
		bestie.Status.AppVersion = appVersion
		log.Info("update app version status")
		err := r.Status().Update(ctx, bestie)
		if err != nil {
			log.Error(err, "Failed to update app-version status")
			return ctrl.Result{Requeue: true}, err
		}
	}

	//Level 1 : update application status
	err = r.updateApplicationStatus(ctx, bestie)
	if err != nil {
		log.Error(err, "Failed to update bestie application status")
		return ctrl.Result{Requeue: true}, err
	}

	//seed the database - as long as the postgres app is up and running this can run
	job := &batchv1.Job{}

	err = r.Get(ctx, types.NamespacedName{Name: BestieName + "-job", Namespace: bestie.Namespace}, job)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating a new job for bestie")
			fileName := "config/resources/bestie-job.yaml"
			r.applyManifests(ctx, req, bestie, job, fileName)
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}

	// reconcile service
	svc := &corev1.Service{}

	err = r.Get(ctx, types.NamespacedName{Name: BestieName + "-service", Namespace: bestie.Namespace}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating a new service for bestie")
			fileName := "config/resources/bestie-svc.yaml"
			r.applyManifests(ctx, req, bestie, svc, fileName)
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}

	// reconcile route
	route := &routev1.Route{}

	err = r.Get(ctx, types.NamespacedName{Name: BestieName + "-route", Namespace: bestie.Namespace}, route)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating a new route for bestie")
			fileName := "config/resources/bestie-route.yaml"
			r.applyManifests(ctx, req, bestie, route, fileName)
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BestieReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&petsv1.Bestie{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&routev1.Route{}).
		Complete(r)
}

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
	"fmt"
	"reflect"
	"time"

	pgov1 "github.com/crunchydata/postgres-operator/pkg/apis/postgres-operator.crunchydata.com/v1beta1"

	petsv1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	controller "sigs.k8s.io/controller-runtime/pkg/controller"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = ctrllog.Log.WithName("controller_bestie")

// BestieReconciler reconciles a Bestie object.
type BestieReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	BestieDefaultImage   = "quay.io/mkong/bestiev2"
	BestieDefaultVersion = "1.2.3"
	BestieName           = "bestie"
)

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
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

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
			err := r.applyManifests(ctx, bestie, pgo, fileName)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("Error during Manifests apply - %w", err)
			}
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}

	// wait for postgres to come up
	var postgresReady = false
	if pgo.Status.InstanceSets != nil && len(pgo.Status.InstanceSets) > 0 && pgo.Status.InstanceSets[0].ReadyReplicas >= 1 {
		postgresReady = true
	}
	if !postgresReady {
		// If postgres is not ready yet, requeue after delay seconds.
		delay := time.Second * time.Duration(15)
		log.Info(fmt.Sprintf("postgres is instantiating, waiting for %s", delay))
		return ctrl.Result{RequeueAfter: delay}, nil
	}

	// reconcile Deployment.
	dp := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: BestieName + "-app", Namespace: bestie.Namespace}, dp)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating a new app for bestie")
			fileName := "config/resources/bestie-deploy.yaml"
			err := r.applyManifests(ctx, bestie, dp, fileName)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("Error during Manifests apply - %w", err)
			}
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}

	// Ensure the deployment size is the same as the spec
	size := bestie.Spec.Size
	if *dp.Spec.Replicas != size {
		dp.Spec.Replicas = &size
		err = r.Update(ctx, dp)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", dp.Namespace, "Deployment.Name", dp.Name)
			return ctrl.Result{}, err
		}
		// Ask to requeue after 1 minute in order to give enough time for the
		// pods be created on the cluster side and the operand be able
		// to do the next update step accurately.
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	//isAppRunning
	bestieRunning := r.isBestieRunning(ctx, bestie)

	if !bestieRunning {
		// If bestie-app isn't running yet, requeue the reconcile
		// to run again after a delay.
		delay := time.Second * time.Duration(15)

		log.Info(fmt.Sprintf("bestie-app is instantiating, waiting for %s", delay))
		return ctrl.Result{RequeueAfter: delay}, nil
	}

	//Level 2 : update Operand.
	err = r.upgradeOperand(ctx, bestie)
	if err != nil {
		log.Error(err, "Failed to upgrade the operand")
		return ctrl.Result{Requeue: true}, err
	}

	//Level 1 : update appVersion status.
	appVersion := r.getDeployedBestieVersion(ctx, bestie)
	if !reflect.DeepEqual(appVersion, bestie.Status.AppVersion) {
		bestie.Status.AppVersion = appVersion
		log.Info("update app version status")
		err := r.Status().Update(ctx, bestie)
		if err != nil {
			log.Error(err, "Failed to update app-version status")
			return ctrl.Result{}, err
		}
	}

	//Level 1 : update application status.
	_, err = r.updateApplicationStatus(ctx, bestie)
	if err != nil {
		log.Error(err, "Failed to update bestie application status")
		return ctrl.Result{Requeue: true}, err
	}

	//seed the database - as long as the postgres app is up and running this can run.
	job := &batchv1.Job{}

	err = r.Get(ctx, types.NamespacedName{Name: BestieName + "-job", Namespace: bestie.Namespace}, job)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating a new job for bestie")
			fileName := "config/resources/bestie-job.yaml"
			err := r.applyManifests(ctx, bestie, job, fileName)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("Error during Manifests apply - %w", err)
			}
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}

	// reconcile service.
	svc := &corev1.Service{}

	err = r.Get(ctx, types.NamespacedName{Name: BestieName + "-service", Namespace: bestie.Namespace}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating a new service for bestie")
			fileName := "config/resources/bestie-svc.yaml"
			err := r.applyManifests(ctx, bestie, svc, fileName)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("Error during Manifests apply - %w", err)
			}
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}

	//Reconciling HPA.
	hpa := &autoscalingv1.HorizontalPodAutoscaler{}

	err = r.Get(ctx, types.NamespacedName{Name: BestieName + "-hpa", Namespace: bestie.Namespace}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating New HPA Instance")
			isHpa := horizontalpodautoscalers(ctx, *bestie.DeepCopy(), r.Client, r.Scheme)
			if isHpa != nil {
				return ctrl.Result{Requeue: true}, isHpa
			}
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	}

	// Checking to see if cluster is an OpenShift cluster.
	// Checks for this api "route.openshift.io/v1".
	isOpenShiftCluster, err := verifyOpenShiftCluster(routev1.GroupName, routev1.SchemeGroupVersion.Version)
	if err != nil {
		return ctrl.Result{}, err
	}

	// If the cluster is OpenShift, add a route, else add an ingress.
	if isOpenShiftCluster {

		utilruntime.Must(routev1.AddToScheme(runtime.NewScheme()))

		route := &routev1.Route{}
		err = r.Get(ctx, types.NamespacedName{Name: bestie.Name + "-route", Namespace: bestie.Namespace}, route)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Info("Creating a new route for bestie")
				fileName := "config/resources/bestie-route.yaml"
				err := r.applyManifests(ctx, bestie, route, fileName)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("Error during Manifests apply - %w", err)
				}
			} else {
				log.Error(err, "Failed to get route.")
				return ctrl.Result{Requeue: true}, err
			}
			// TODO: should we update then?
		}
	} else {
		ingress := &networkv1.Ingress{}
		err = r.Get(ctx, types.NamespacedName{Name: bestie.Name + "-ingress", Namespace: bestie.Namespace}, ingress)
		if err != nil && errors.IsNotFound(err) {

			log.Info("Creating a new ingress for bestie")
			fileName := "config/resources/bestie-ingress.yaml"
			err = r.applyManifests(ctx, bestie, ingress, fileName)

			if err != nil {
				log.Error(err, "Failed to get ingress.")
				return ctrl.Result{Requeue: true}, err
			}

			log.Info("Ingress Created Successfully", "Ingress.Namespace", ingress.Namespace, "Ingress.Name", ingress.Name)
			return ctrl.Result{Requeue: true}, nil

		} else if err != nil {
			log.Error(err, "Failed to get Ingress")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BestieReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&petsv1.Bestie{})
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.Service{})
	builder.Owns(&networkv1.Ingress{})
	builder.Owns(&autoscalingv1.HorizontalPodAutoscaler{})
	if IsRouteAPIAvailable() {
		builder.Owns(&routev1.Route{})
	}
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: 2})

	return builder.Complete(r)
}

var routeAPIFound = false

func IsRouteAPIAvailable() bool {
	err := verifyRouteAPI()
	if err != nil {
		return false
	}
	return routeAPIFound
}

func verifyRouteAPI() error {
	found, err := verifyOpenShiftCluster(routev1.GroupName, routev1.SchemeGroupVersion.Version)
	if err != nil {
		return err
	}
	routeAPIFound = found
	return nil
}

func verifyOpenShiftCluster(group string, version string) (bool, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return false, err
	}

	k8s, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, err
	}

	gv := schema.GroupVersion{
		Group:   group,
		Version: version,
	}

	if err = discovery.ServerSupportsVersion(k8s, gv); err != nil {
		return false, nil
	}

	return true, nil
}

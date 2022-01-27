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

	petsv1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *BestieReconciler) ensureJob(ctx context.Context,
	bestie *petsv1.Bestie,
	job *batchv1.Job) (*reconcile.Result, error) {
	// Check if the Job already exists, if not create a new one
	found := &batchv1.Job{}
	err := r.Get(ctx, types.NamespacedName{Name: job.Name, Namespace: bestie.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Job
		log.Info("Creating a new Job for", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		err = r.Create(ctx, job)
		if err != nil {
			log.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return &reconcile.Result{}, err
		} else {
			// Job created successfully - return and requeue
			return nil, nil
		}
	} else if err != nil {
		log.Error(err, "Failed to get Job")
		return &reconcile.Result{}, err
	}
	return nil, nil
}

func (r *BestieReconciler) ensureDeployment(ctx context.Context,
	bestie *petsv1.Bestie,
	dep *appsv1.Deployment) (*reconcile.Result, error) {
	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: bestie.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		log.Info("Creating a new Deployment for", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return &reconcile.Result{}, err
		} else {
			// Deployment created successfully - return and requeue
			return nil, nil
		}
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return &reconcile.Result{}, err
	}
	return nil, nil
}

func (r *BestieReconciler) ensureService(ctx context.Context,
	bestie *petsv1.Bestie,
	serv *corev1.Service,
) (*reconcile.Result, error) {
	found := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: serv.Name, Namespace: bestie.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the service
		log.Info("Creating a new service", "Service.Namespace", serv.Namespace, "Service.Name", serv.Name)
		err = r.Create(ctx, serv)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Service", "Service.Namespace", serv.Namespace, "Service.Name", serv.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get Service")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *BestieReconciler) ensureRoute(ctx context.Context,
	bestie *petsv1.Bestie,
	route *routev1.Route,
) (*reconcile.Result, error) {
	found := &routev1.Route{}
	err := r.Get(ctx, types.NamespacedName{Name: route.Name, Namespace: bestie.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the route
		log.Info("Creating a new Route", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
		err = r.Create(ctx, route)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Route", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the route not existing
		log.Error(err, "Failed to get Route")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func labels(v *petsv1.Bestie, tier string) map[string]string {
	return map[string]string{
		"app":         "Bestie",
		"demosite_cr": v.Name,
		"tier":        tier,
	}
}

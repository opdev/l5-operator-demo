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
	petsv1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

const bestiePort = 8080

func bestieAppServiceName(bestie *petsv1.Bestie) string {
	return bestie.Name + "-service"
}

func bestieAppDeploymentName(bestie *petsv1.Bestie) string {
	return bestie.Name + "-deployment"
}

//route for bestie application
func (r *BestieReconciler) bestieRoute(bestie *petsv1.Bestie) *routev1.Route {
	labels := labels(bestie, "bestie")

	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bestieAppServiceName(bestie),
			Namespace: bestie.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: bestieAppServiceName(bestie),
			},
		},
	}
	ctrl.SetControllerReference(bestie, route, r.Scheme)
	return route
}

// service for bestie application
func (r *BestieReconciler) bestieAppService(bestie *petsv1.Bestie) *corev1.Service {
	labels := labels(bestie, "bestie")

	serv := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bestieAppServiceName(bestie),
			Namespace: bestie.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       5000,
				TargetPort: intstr.FromInt(8080),
				NodePort:   0,
			}},
		},
	}
	ctrl.SetControllerReference(bestie, serv, r.Scheme)
	return serv
}

// bestieDeployment returns a bestie Deployment object
func (r *BestieReconciler) bestieAppDeployment(bestie *petsv1.Bestie) *appsv1.Deployment {
	labels := labels(bestie, "bestie")
	size := bestie.Spec.Size

	userSecret := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: mysqlAuthName()},
			Key:                  "username",
		},
	}

	passwordSecret := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: mysqlAuthName()},
			Key:                  "password",
		},
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bestieAppDeploymentName(bestie),
			Namespace: bestie.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "quay.io/rocrisp/cakedemo:v1",
						Name:  "bestie-demo",
						Ports: []corev1.ContainerPort{{
							ContainerPort: bestiePort,
							Name:          "bestie",
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "DATABASE_SERVICE_NAME",
								Value: "mysql",
							},
							{
								Name:  "DATABASE_NAME",
								Value: "cakephp",
							},
							{
								Name:  "FIRST_LASTNAME",
								Value: bestie.Spec.AgencyName,
							},
							{
								Name:  "MYSQL_SERVICE_HOST",
								Value: "mysql",
							},
							{
								Name:  "SESSION_DEFAULTS",
								Value: "database",
							},
							{
								Name:      "DATABASE_USER",
								ValueFrom: userSecret,
							},
							{
								Name:      "DATABASE_PASSWORD",
								ValueFrom: passwordSecret,
							},
						},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(bestie, dep, r.Scheme)
	return dep
}

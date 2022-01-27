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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

const bestiePort = 8000
const databaseServiceName = "mysql"
const databaseName = "bestie"
const sessionDefaults = "database"
const bestieImage = "quay.io/skattoju/bestie:v1"

var weight int32 = 100

func bestieAppServiceName(bestie *petsv1.Bestie) string {
	return bestie.Name + "-service"
}

func bestieAppDeploymentName(bestie *petsv1.Bestie) string {
	return bestie.Name + "-deployment"
}

func bestieJob(bestie *petsv1.Bestie) string {
	return bestie.Name + "-job"
}

// newJob returns a new Job instance.
func (r *BestieReconciler) bestieJob(bestie *petsv1.Bestie) *batchv1.Job {
	labels := labels(bestie, "bestie")

	host := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "hippo-pguser-hippo"},
			Key:                  "host",
		},
	}

	port := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "hippo-pguser-hippo"},
			Key:                  "port",
		},
	}

	dbname := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "hippo-pguser-hippo"},
			Key:                  "dbname",
		},
	}

	user := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "hippo-pguser-hippo"},
			Key:                  "user",
		},
	}

	password := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "hippo-pguser-hippo"},
			Key:                  "password",
		},
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bestieJob(bestie),
			Namespace: bestie.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           bestieImage,
						ImagePullPolicy: corev1.PullAlways,
						Name:            "bestie-demo",
						Ports: []corev1.ContainerPort{{
							ContainerPort: bestiePort,
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "GUNICORN_CMD_ARGS",
								Value: "--bind=0.0.0.0 --workers=3",
							},
							{
								Name:  "FLASK_APP",
								Value: "app",
							},
							{
								Name:  "FLASK_ENV",
								Value: "development",
							},
							{
								Name:  "SECRET_KEY",
								Value: "secret-key",
							},
							{
								Name:      "DB_ADDR",
								ValueFrom: host,
							},
							{
								Name:      "DB_PORT",
								ValueFrom: port,
							},
							{
								Name:      "DB_DATABASE",
								ValueFrom: dbname,
							},
							{
								Name:      "DB_USER",
								ValueFrom: user,
							},
							{
								Name:      "DB_PASSWORD",
								ValueFrom: password,
							},
							{
								Name:  "DATABASE_URL",
								Value: "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_ADDR):$(DB_PORT)/$(DB_DATABASE)",
							},
						},
						Command: []string{
							"/bin/sh",
							"-c",
						},
						Args: []string{
							"flask db migrate",
							"flask db upgrade",
							"flask seed all",
						},
					}},
					RestartPolicy: "OnFailure",
				},
			},
		},
	}
	ctrl.SetControllerReference(bestie, job, r.Scheme)
	return job
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
				Kind:   "Service",
				Name:   bestieAppServiceName(bestie),
				Weight: &weight,
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
			Type:     "LoadBalancer",
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromInt(8000),
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

	host := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "hippo-pguser-hippo"},
			Key:                  "host",
		},
	}

	port := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "hippo-pguser-hippo"},
			Key:                  "port",
		},
	}

	dbname := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "hippo-pguser-hippo"},
			Key:                  "dbname",
		},
	}

	user := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "hippo-pguser-hippo"},
			Key:                  "user",
		},
	}

	password := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "hippo-pguser-hippo"},
			Key:                  "password",
		},
	}
	//log.Info("password is :", password)

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
						Image: bestieImage,
						Name:  "bestie-demo",
						Ports: []corev1.ContainerPort{{
							ContainerPort: bestiePort,
							Name:          "http",
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "GUNICORN_CMD_ARGS",
								Value: "--bind=0.0.0.0 --workers=3",
							},
							{
								Name:  "FLASK_APP",
								Value: "app",
							},
							{
								Name:  "FLASK_ENV",
								Value: "development",
							},
							{
								Name:  "SECRET_KEY",
								Value: "secret-key",
							},
							{
								Name:      "DB_ADDR",
								ValueFrom: host,
							},
							{
								Name:      "DB_PORT",
								ValueFrom: port,
							},
							{
								Name:      "DB_DATABASE",
								ValueFrom: dbname,
							},
							{
								Name:      "DB_USER",
								ValueFrom: user,
							},
							{
								Name:      "DB_PASSWORD",
								ValueFrom: password,
							},
							{
								Name:  "DATABASE_URL",
								Value: "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_ADDR):$(DB_PORT)/$(DB_DATABASE)",
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

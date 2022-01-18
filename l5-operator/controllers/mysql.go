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

	//"time"
	"context"

	petsv1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

const sqlPort = 3306

func mysqlDeploymentName() string {
	return "mysql"
}

func mysqlServiceName() string {
	return "mysql"
}

func mysqlAuthName() string {
	return "mysql-auth"
}

func (r *BestieReconciler) mysqlAuthSecret(bestie *petsv1.Bestie) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysqlAuthName(),
			Namespace: bestie.Namespace,
		},
		Type: "Opaque",
		StringData: map[string]string{
			"username": "bestieadmin",
			"password": "cakephp",
		},
	}
	ctrl.SetControllerReference(bestie, secret, r.Scheme)
	return secret
}

func (r *BestieReconciler) mysqlService(bestie *petsv1.Bestie) *corev1.Service {
	labels := labels(bestie, "mysql")

	serv := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysqlServiceName(),
			Namespace: bestie.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Port:       3306,
				TargetPort: intstr.FromInt(sqlPort),
			}},
		},
	}
	ctrl.SetControllerReference(bestie, serv, r.Scheme)
	return serv
}

func (r *BestieReconciler) isMysqlUp(bestie *petsv1.Bestie) bool {
	deployment := &appsv1.Deployment{}

	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      mysqlDeploymentName(),
		Namespace: bestie.Namespace,
	}, deployment)

	if err != nil {
		log.Error(err, "Mysql Deployment is not found")
		return false
	}

	if deployment.Status.ReadyReplicas == 1 {
		return true
	}

	return false
}

// deploymentForMemcached returns a memcached Deployment object
func (r *BestieReconciler) mysqlDeployment(bestie *petsv1.Bestie) *appsv1.Deployment {
	labels := labels(bestie, "mysql")
	size := int32(1)

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
			Name:      mysqlDeploymentName(),
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
					Volumes: []corev1.Volume{{
						Name: "mysql-data",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}},
					Containers: []corev1.Container{{
						Image: "mysql:5.7",
						Name:  "mysql-server",
						Ports: []corev1.ContainerPort{{
							ContainerPort: sqlPort,
							Name:          "mysql",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "mysql-data",
							MountPath: "/var/lib/mysql",
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "MYSQL_ROOT_PASSWORD",
								Value: "cakephp",
							},
							{
								Name:  "MYSQL_DATABASE",
								Value: "cakephp",
							},
							{
								Name:      "MYSQL_USER",
								ValueFrom: userSecret,
							},
							{
								Name:      "MYSQL_PASSWORD",
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

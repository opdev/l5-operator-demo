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
	"os"
	"reflect"
	"strings"

	petsv1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Returns true if readyReplicas=1
func (r *BestieReconciler) isRunning(ctx context.Context, bestie *petsv1.Bestie) bool {
	dp := &appsv1.Deployment{}

	err := r.Get(ctx, types.NamespacedName{Name: BestieName + "-app", Namespace: bestie.Namespace}, dp)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(BestieName + " deployment is not found.")
			return false
		}
	}
	if dp.Status.ReadyReplicas == 1 {
		return true
	}
	return false
}

func (r *BestieReconciler) reportappversion(bestie *petsv1.Bestie) string {

	tag := BestieDefaultVersion
	if len(bestie.Spec.Version) > 0 {
		tag = bestie.Spec.Version
	}
	return tag
}

func (r *BestieReconciler) applyManifests(ctx context.Context, req ctrl.Request, bestie *petsv1.Bestie, obj client.Object, fileName string) error {

	Log := ctrllog.FromContext(ctx)

	b, err := os.ReadFile(fileName)
	if err != nil {
		Log.Error(err, fmt.Sprintf("Couldn't read manifest file for: %s", fileName))
		return err
	}

	if err = yamlutil.Unmarshal(b, &obj); err != nil {
		Log.Error(err, fmt.Sprintf("Couldn't unmarshall yaml file for: %s", fileName))
		return err
	}

	obj.SetNamespace(bestie.GetNamespace())

	controllerutil.SetControllerReference(bestie, obj, r.Scheme)

	err = r.Client.Create(ctx, obj)
	if err != nil {
		Log.Error(err, "Failed to create object", "object", obj.GetName())
		return err
	}

	return nil
}

// getBestieContainerImage will return the container image for the Bestie App Image.
func getBestieContainerImage(bestie *petsv1.Bestie) string {
	img := BestieDefaultImage
	if len(bestie.Spec.Image) > 0 {
		img = bestie.Spec.Image
	}

	tag := BestieDefaultVersion
	if len(bestie.Spec.Version) > 0 {
		tag = bestie.Spec.Version
	}

	return CombineImageTag(img, tag)
}

// CombineImageTag will return the combined image and tag in the proper format for tags and digests.
func CombineImageTag(img string, tag string) string {
	if strings.Contains(tag, ":") {
		return fmt.Sprintf("%s@%s", img, tag) // Digest
	} else if len(tag) > 0 {
		return fmt.Sprintf("%s:%s", img, tag) // Tag
	}
	return img // No tag, use default
}

//Create a container object with the image from cr
func createContainer(bestie *petsv1.Bestie) corev1.Container {
	host := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "bestie-pgo-pguser-bestie-pgo"},
			Key:                  "host",
		},
	}

	port := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "bestie-pgo-pguser-bestie-pgo"},
			Key:                  "port",
		},
	}

	dbname := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "bestie-pgo-pguser-bestie-pgo"},
			Key:                  "dbname",
		},
	}

	user := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "bestie-pgo-pguser-bestie-pgo"},
			Key:                  "user",
		},
	}

	password := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "bestie-pgo-pguser-bestie-pgo"},
			Key:                  "password",
		},
	}

	container := corev1.Container{
		Name:  "bestie",
		Image: getBestieContainerImage(bestie),
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: 8000,
				Name:          "http",
			},
		},
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
	}
	return container
}

func (r *BestieReconciler) upgradeOperand(ctx context.Context, bestie *petsv1.Bestie) error {
	dp := &appsv1.Deployment{}
	desired := &appsv1.Deployment{}
	container := createContainer(bestie)

	err := r.Get(ctx, types.NamespacedName{Name: BestieName + "-app", Namespace: bestie.Namespace}, dp)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("bestie-app not found")
			return err
		}
	}

	desired.Spec.Template.Spec.Containers = append(desired.Spec.Template.Spec.Containers, container)

	bestieImageDifferent := !reflect.DeepEqual(dp.Spec.Template.Spec.Containers[0].Image, desired.Spec.Template.Spec.Containers[0].Image)

	if bestieImageDifferent {
		log.Info("Upgrade Operand")
		dp.Spec.Template.Spec.Containers = desired.Spec.Template.Spec.Containers
		err = r.Client.Update(ctx, dp)
		if err != nil {
			log.Error(err, "Need to update, but failed to update bestie image")
			return err
		}
	}
	return nil
}

// getPodNames returns the pod names of the array of pods passed in
func (r *BestieReconciler) updateApplicationStatus(ctx context.Context, bestie *petsv1.Bestie) error {
	var bestiePodStatus string

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(bestie.Namespace),
		client.MatchingLabels{"app": "bestie"},
	}
	if err := r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "bestie.Namespace", bestie.Namespace, "Bestie.Name", bestie.Name)
		return err
	}

	for _, pod := range podList.Items {
		podName := pod.GetName()

		if strings.Contains(podName, "-app") {
			bestiePodStatus = string(pod.Status.Phase)
			bestieStatusDifferent := !reflect.DeepEqual(bestie.Status.AppStatus, bestiePodStatus)
			if bestieStatusDifferent {
				log.Info("Update bestie application status")
				bestie.Status.AppStatus = bestiePodStatus
				err := r.Status().Update(ctx, bestie)
				if err != nil {
					log.Error(err, "Failed to update bestie application status")
					return err
				}
			}
		} else {
			fmt.Println("The substring is not present in the string.")
		}
	}
	return nil
}
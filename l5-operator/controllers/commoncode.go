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
	"os"
	"reflect"
	"strings"

	petsv1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Returns true if readyReplicas=1.
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

func (r *BestieReconciler) applyManifests(ctx context.Context, bestie *petsv1.Bestie, obj client.Object, fileName string) error {

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

	err = controllerutil.SetControllerReference(bestie, obj, r.Scheme)
	if err != nil {
		return err
	}

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

func (r *BestieReconciler) upgradeOperand(ctx context.Context, bestie *petsv1.Bestie) error {
	dp := &appsv1.Deployment{}

	err := r.Get(ctx, types.NamespacedName{Name: BestieName + "-app", Namespace: bestie.Namespace}, dp)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("bestie-app not found")
			return err
		}
	}

	//compare deployment container image to bestie spec image+version
	bestieImageDifferent := !reflect.DeepEqual(dp.Spec.Template.Spec.Containers[0].Image, getBestieContainerImage(bestie))

	if bestieImageDifferent {
		log.Info("Upgrade Operand")
		dp.Spec.Template.Spec.Containers[0].Image = getBestieContainerImage(bestie)
		err = r.Client.Update(ctx, dp)
		if err != nil {
			log.Error(err, "Need to update, but failed to update bestie image")
			return err
		}
	}
	return nil
}

// getPodNames returns the pod names of the array of pods passed in.
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

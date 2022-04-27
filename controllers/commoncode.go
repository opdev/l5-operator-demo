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

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"

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

// Returns true if readyReplicas=1.
func (r *BestieReconciler) isBestieRunning(ctx context.Context, bestie *petsv1.Bestie) bool {
	dp := &appsv1.Deployment{}

	err := r.Get(ctx, types.NamespacedName{Name: BestieName + "-app", Namespace: bestie.Namespace}, dp)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(BestieName + " deployment is not found.")
			return false
		}
	}
	if dp.Status.ReadyReplicas >= 1 {
		return true
	}
	return false
}

func (r *BestieReconciler) getDeployedBestieVersion(ctx context.Context, bestie *petsv1.Bestie) string {

	dp := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: BestieName + "-app", Namespace: bestie.Namespace}, dp)
	if err != nil {
		log.Error(err, "unable to retrieve deployment")
		return "unknown"
	}
	imageUri := dp.Spec.Template.Spec.Containers[0].Image
	version := strings.Split(imageUri, ":")[1]
	return version
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

	//compare current container image to spec image
	bestieImageDifferent := !reflect.DeepEqual(dp.Spec.Template.Spec.Containers[0].Image, getBestieContainerImage(bestie))

	if bestieImageDifferent {
		if bestieImageDifferent {
			log.Info("Upgrade Operand")
			dp.Spec.Template.Spec.Containers[0].Image = getBestieContainerImage(bestie)
		}
		err = r.Client.Update(ctx, dp)
		if err != nil {
			log.Error(err, "Deployment failed.")
			return err
		}
		// Level 4 Add metrics
		applicationUpgradeCounter.Inc()
	}
	return nil
}

// getPodNames returns the pod names of the array of pods passed in.
func (r *BestieReconciler) updateApplicationStatus(ctx context.Context, bestie *petsv1.Bestie) (ctrl.Result, error) {

	podList := &corev1.PodList{}

	listOpts := []client.ListOption{
		client.InNamespace(bestie.Namespace),
		client.MatchingLabels{"app": "bestie"},
	}

	if err := r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "bestie.Namespace", bestie.Namespace, "Bestie.Name", bestie.Name)
		return ctrl.Result{}, err
	}

	//Be Careful When Listing Pods... Some May Be in Terminating Status..
	var nonTerminatedPodList []corev1.Pod
	for _, pod := range podList.Items {
		if pod.ObjectMeta.DeletionTimestamp == nil {
			nonTerminatedPodList = append(nonTerminatedPodList, pod)
		}
	}

	//Update status if needed
	appStatus := getPodNamesandStatuses(nonTerminatedPodList)
	log.Info(fmt.Sprintf("The pod status is %v", appStatus))

	bestieStatusDifferent := !reflect.DeepEqual(appStatus, bestie.Status.PodStatus)
	if bestieStatusDifferent {
		log.Info("Update bestie application status")
		bestie.Status.PodStatus = appStatus
		err := r.Status().Update(ctx, bestie)
		if err != nil {
			log.Error(err, "Failed to update bestie application status")
			return ctrl.Result{}, err
		}
		//requeue
		return ctrl.Result{Requeue: true}, nil
	}

	//Level 4 - Update metric
	rc := getPodstatusReason(podList.Items)
	log.Info(fmt.Sprintf("return code for getPodstatusReason %f", rc))
	applicationUpgradeFailure.Set(rc)
	return ctrl.Result{}, nil
}

// getPodNameandStatuses returns the pod names+status of the array of pods passed in.
func getPodNamesandStatuses(pods []corev1.Pod) []string {
	var podNamesStatus []string
	var podStat string
	for _, pod := range pods {
		podStat = pod.Name + " : " + string(pod.Status.Phase) + " : " + string(pod.Spec.Containers[0].Image)
		podNamesStatus = append(podNamesStatus, podStat)
	}
	return podNamesStatus
}

// getPodNameandStatuses returns the pod names+status of the array of pods passed in
func getPodstatusReason(pods []corev1.Pod) float64 {
	// return 0 if not found, otherwise return 1
	for _, pod := range pods {
		if string(pod.Status.Phase) == "Pending" {
			stat := pod.Status.ContainerStatuses[0].State.Waiting.Reason
			log.Info(stat)
			if string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ErrImagePull" ||
				string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ImagePullBackOff" {
				return 1
			}
		}
	}
	return 0
}

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
	"reflect"
	"sort"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/opdev/l5-operator-demo/internal/util"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
	"github.com/opdev/l5-operator-demo/internal/bestie_metrics"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeploymentImageReconciler struct {
	client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func NewDeploymentImageReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *DeploymentImageReconciler {
	return &DeploymentImageReconciler{
		client: client,
		Log:    log,
		Scheme: scheme,
	}
}

const (
	BestieDefaultImage   = "quay.io/opdev/bestie"
	BestieDefaultVersion = "1.3"
)

func (r *DeploymentImageReconciler) Reconcile(ctx context.Context, bestie *petsv1.Bestie) (ctrl.Result, error) {
	logInfo := types.NamespacedName{
		Namespace: bestie.Namespace,
		Name:      "DeploymentImageReconciler",
	}
	log := r.Log.WithValues("deployment_image", logInfo)

	if !r.isBestieRunning(ctx, *bestie) {
		// If bestie-app isn't running yet, requeue the reconcile
		// to run again after a delay.
		delay := time.Second * time.Duration(15)
		log.Info(fmt.Sprintf("bestie-app is instantiating, requeue reconcile for %s seconds", delay))
		return ctrl.Result{RequeueAfter: delay}, nil
	}

	err := r.updateDeploymentReadyCondition(ctx, *bestie)
	if err != nil {
		log.Error(err, "Unable to update DeploymentReady status condition")
		return ctrl.Result{}, err
	}

	// Level 2 : update Operand.
	r.Log.Info("reconcile bestie version")
	err = r.upgradeOperand(ctx, *bestie)
	if err != nil {
		log.Error(err, "Failed to upgrade the operand")
		return ctrl.Result{Requeue: true}, err
	}

	// Level 2 : update appVersion status.
	log.Info("update bestie version status")
	err = util.RefreshCustomResource(ctx, r.client, bestie)
	if err != nil {
		log.Error(err, "Unable to refresh bestie custom resource")
		return ctrl.Result{}, err
	}
	appVersion := r.getDeployedBestieVersion(ctx, *bestie)
	if !reflect.DeepEqual(appVersion, bestie.Status.AppVersion) {
		bestie.Status.AppVersion = appVersion
		log.Info("update app version status")
		err := r.client.Status().Update(ctx, bestie)
		if err != nil {
			r.Log.Error(err, "Failed to update app-version status")
			return ctrl.Result{}, err
		}
	}

	// Level 2 : update application status.
	log.Info("update bestie pods status")
	err = util.RefreshCustomResource(ctx, r.client, bestie)
	if err != nil {
		log.Error(err, "Unable to refresh bestie custom resource")
		return ctrl.Result{}, err
	}
	_, err = r.updateApplicationStatus(ctx, *bestie)
	if err != nil {
		log.Error(err, "Failed to update bestie application status")
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

func (r *DeploymentImageReconciler) upgradeOperand(ctx context.Context, bestie petsv1.Bestie) error {
	dp := &appsv1.Deployment{}

	err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-app", Namespace: bestie.Namespace}, dp)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("bestie-app not found")
			return err
		}
	}

	// compare current container image to spec image
	imageInDeployment := "unknown"
	positionInContainerArray := 0
	for pos, container := range dp.Spec.Template.Spec.Containers {
		if container.Name == "bestie" {
			imageInDeployment = container.Image
			positionInContainerArray = pos
		}
	}
	if imageInDeployment != getBestieContainerImage(bestie) {
		r.Log.Info("Updating deployment")
		dp.Spec.Template.Spec.Containers[positionInContainerArray].Image = getBestieContainerImage(bestie)
		err := r.client.Update(ctx, dp)
		if err != nil {
			r.Log.Error(err, "Failed to update deployment")
			return err
		}
		// Level 4 Add metrics
		r.Log.Info("update metric : ApplicationUpgradeCounter")
		bestie_metrics.ApplicationUpgradeCounter.Inc()
	}
	return nil
}

// Returns true if readyReplicas=1.
func (r *DeploymentImageReconciler) isBestieRunning(ctx context.Context, bestie petsv1.Bestie) bool {
	dp := &appsv1.Deployment{}
	err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-app", Namespace: bestie.Namespace}, dp)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info(dp.Name + " deployment is not found.")
			return false
		}
	}
	if dp.Status.ReadyReplicas >= 1 {
		return true
	}
	return false
}

// getBestieContainerImage will return the container image for the Bestie App Image.
func getBestieContainerImage(bestie petsv1.Bestie) string {
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

func (r *DeploymentImageReconciler) getDeployedBestieVersion(ctx context.Context, bestie petsv1.Bestie) string {
	dp := &appsv1.Deployment{}
	err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-app", Namespace: bestie.Namespace}, dp)
	if err != nil {
		r.Log.Error(err, "unable to retrieve deployment")
		return "unknown"
	}
	imageUri := dp.Spec.Template.Spec.Containers[0].Image
	version := strings.Split(imageUri, ":")[1]
	return version
}

// getPodNames returns the pod names of the array of pods passed in.
func (r *DeploymentImageReconciler) updateApplicationStatus(ctx context.Context, bestie petsv1.Bestie) (ctrl.Result, error) {
	err := util.RefreshCustomResource(ctx, r.client, &bestie)
	if err != nil {
		r.Log.Error(err, "Unable to refresh bestie custom resource")
		return ctrl.Result{}, err
	}
	podList := &corev1.PodList{}

	listOpts := []client.ListOption{
		client.InNamespace(bestie.Namespace),
		client.MatchingLabels{"app": "bestie"},
	}

	if err := r.client.List(ctx, podList, listOpts...); err != nil {
		r.Log.Error(err, "Failed to list pods", "bestie.Namespace", bestie.Namespace, "Bestie.Name", bestie.Name)
		return ctrl.Result{}, err
	}

	// Be Careful When Listing Pods... Some May Be in Terminating Status..
	var nonTerminatedPodList []corev1.Pod
	for _, pod := range podList.Items {
		if pod.ObjectMeta.DeletionTimestamp == nil {
			nonTerminatedPodList = append(nonTerminatedPodList, pod)
		}
	}

	//Level 4 - Update metric.
	if len(nonTerminatedPodList) > 0 {
		rc := getPodstatusReason(nonTerminatedPodList)
		r.Log.Info(fmt.Sprintf("update metric : ApplicationUpgradeFailure %f", rc))
		bestie_metrics.ApplicationUpgradeFailure.Set(rc)
	}

	// Update status if needed.
	appStatus := getPodNamesandStatuses(nonTerminatedPodList)
	sort.Strings(appStatus)
	sort.Strings(bestie.Status.PodStatus)

	bestieStatusDifferent := !reflect.DeepEqual(appStatus, bestie.Status.PodStatus)

	if bestieStatusDifferent {
		r.Log.Info("Pod Status different, update bestie application status")
		bestie.Status.PodStatus = appStatus
		err := r.client.Status().Update(ctx, &bestie)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// getPodNameandStatuses returns the pod names+status of the array of pods passed in.
func getPodNamesandStatuses(pods []corev1.Pod) []string {
	var podNamesStatus []string
	var podStat string
	for _, pod := range pods {
		podStat = pod.Name + " : " + string(pod.Status.Phase) + " : " + pod.Spec.Containers[0].Image
		podNamesStatus = append(podNamesStatus, podStat)
	}
	return podNamesStatus
}

// getPodNameandStatuses returns the pod names+status of the array of pods passed in.
func getPodstatusReason(pods []corev1.Pod) float64 {
	// return 0 if not found, otherwise return 1.
	for _, pod := range pods {
		pendingState := string(pod.Status.Phase)
		if pendingState == "Pending" &&
			len(pod.Status.ContainerStatuses) > 0 &&
			len(pod.Status.ContainerStatuses[0].State.Waiting.Reason) > 0 {
			errorImagePull := string(pod.Status.ContainerStatuses[0].State.Waiting.Reason)
			imagePullBackOff := string(pod.Status.ContainerStatuses[0].State.Waiting.Reason)
			if errorImagePull == "ErrImagePull" ||
				imagePullBackOff == "ImagePullBackOff" {
				return 1
			}
		}
	}
	return float64(0)
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

func (r *DeploymentImageReconciler) updateDeploymentReadyCondition(ctx context.Context, bestie petsv1.Bestie) error {
	err := util.RefreshCustomResource(ctx, r.client, &bestie)
	if err != nil {
		return err
	}
	deploymentReadyCondition := NewDeploymentCreatedCondition()
	deploymentReadyCondition.Status = metav1.ConditionTrue
	deploymentReadyCondition.Reason = "HasMinReadyReplicas"
	deploymentReadyCondition.Message = "Deployment has min ready replicas"
	meta.SetStatusCondition(&bestie.Status.Conditions, deploymentReadyCondition)
	err = r.client.Status().Update(ctx, &bestie)
	if err != nil {
		return err
	}
	return nil
}

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

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v2beta2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/opdev/l5-operator-demo/api/v1"
	"github.com/opdev/l5-operator-demo/internal/hpa"
)

func horizontalpodautoscalers(ctx context.Context, bestieDeployment appsv1.Deployment, bestie v1.Bestie, client cli.Client, r *runtime.Scheme) error {
	desired := []autoscalingv1.HorizontalPodAutoscaler{}

	if bestie.Spec.MaxReplicas != nil {
		log.Info("MaxReplicas is set, enabling HPA")
		desired = append(desired, hpa.AutoScaler(ctrllog.Log, bestieDeployment, bestie))
	}

	if err := applyHorizontalPodAutoscalers(ctx, bestie, client, r, desired); err != nil {
		log.Error(err, "failed to reconcile the expected horizontal pod autoscalers")
		return err
	}
	return nil
}

func applyHorizontalPodAutoscalers(ctx context.Context, bestie v1.Bestie, client cli.Client, r *runtime.Scheme, expected []autoscalingv1.HorizontalPodAutoscaler) error {

	log := ctrllog.FromContext(ctx)

	for _, rec := range expected {
		desired := rec

		if err := controllerutil.SetControllerReference(bestie.DeepCopy(), &desired, r); err != nil {
			return fmt.Errorf("failed to find the controller reference: %w", err)
		}
		existing := &autoscalingv1.HorizontalPodAutoscaler{}
		dNameSpace := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		err := client.Get(ctx, dNameSpace, existing)
		if k8serrors.IsNotFound(err) {
			if err := client.Create(ctx, &desired); err != nil {
				return fmt.Errorf("failed to create: %w", err)
			}
			log.V(2).Info("created", "HPA.Name", desired.Name, "HPA.NameSpace", desired.Namespace)
			continue
		} else if err != nil {
			return fmt.Errorf("failed to get %w", err)
		}

		updated := existing.DeepCopy()
		if updated.Annotations == nil {
			updated.Annotations = map[string]string{}
		}
		if updated.Labels == nil {
			updated.Labels = map[string]string{}
		}

		updated.OwnerReferences = desired.OwnerReferences
		updated.Spec.MinReplicas = &bestie.Spec.Size
		if bestie.Spec.MaxReplicas != nil {
			updated.Spec.MaxReplicas = *bestie.Spec.MaxReplicas
		}

		for k, v := range desired.Annotations {
			updated.Annotations[k] = v
		}

		for k, v := range desired.Labels {
			updated.Labels[k] = v
		}

		patch := cli.MergeFrom(existing)

		if err := client.Patch(ctx, updated, patch); err != nil {
			return fmt.Errorf("failed to apply changes: %w", err)
		}

		log.V(2).Info("applied", "HPA.Name", desired.Name, "HPA.NameSpace", desired.Namespace)
	}
	return nil
}

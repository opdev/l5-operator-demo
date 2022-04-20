package controllers

import (
	"context"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"
)

func (r *BestieReconciler) applyHorizontalPodAutoscalers(ctx context.Context, bestie v1.Bestie, expected []autoscalingv2.HorizontalPodAutoscaler) error {
	log := ctrllog.FromContext(ctx)

	for _, rec := range expected {
		desired := rec 

		if err := controllerutil.SetControllerReference(bestie.DeepCopy(), &desired, r.Scheme); err != nil {
			fmt.Errorf("failed to find the controller reference: %w", err)
		}
		existing := &autoscalingv2.HorizontalPodAutoscaler{}
		dNameSpace := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		err := r.Client.Get(ctx, dNameSpace, existing)
		if k8serrors.IsNotFound(err) {
			if err := r.Client.Create(ctx, &desired); err != nil {
				return fmt.Errorf("failed to create: %w", err)
			}
			log.V(2).Info("created", "HPA.Name", desired.Name, "HPA.NameSpace", desired.Namespace)
			continue
		} else if err != nil  {
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
		updated.Spec.MinReplicas = bestie.Spec.Replicas
		if bestie.Spec.MaxReplicas != nil {
			updated.Spec.MaxReplicas = *bestie.Spec.MaxReplicas
		}
		
		for k, v := range desired.Annotations {
			updated.Annotations[k] = v
		}

		for k, v := range desired.Labels {
			updated.Labels[k] = v
		}

		patch := client.MergeFrom(existing)

		if err := r.Client.Patch(ctx, updated, patch); err != nil {
			return fmt.Errorf("failed to apply changes: %w", err)
		}

		log.V(2).Info("applied", "HPA.Name", desired.Name, "HPA.NameSpace", desired.Namespace)
	}
	return nil
}

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

package hpa

import (
	"github.com/go-logr/logr"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"
)

const (
	defaultCPUTarget = int32(90)
)

// Autoscalers returns a list of HPAs based on specs.
func AutoScaler(logger logr.Logger, bestie v1.Bestie) autoscalingv1.HorizontalPodAutoscaler {
	cpuTarget := defaultCPUTarget
	return autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bestie.Name + "hpa",
			Namespace: bestie.Namespace,
			Labels:    bestie.Labels,
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       bestie.Name,
			},
			MinReplicas:                    bestie.Spec.Replicas,
			MaxReplicas:                    *bestie.Spec.MaxReplicas,
			TargetCPUUtilizationPercentage: &cpuTarget,
		},
	}
}

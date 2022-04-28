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
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/opdev/l5-operator-demo/api/v1"
)

const (
	defaultCPUTarget = int32(30)
)

// Autoscalers returns an HPAs based on specs.
func AutoScaler(logger logr.Logger, bestieDeployment appsv1.Deployment, bestie v1.Bestie) autoscalingv2.HorizontalPodAutoscaler {
	targetCpuUtilization := defaultCPUTarget
	cpuTarget := autoscalingv2.ResourceMetricSource{
		Name: "cpu",
		Target: autoscalingv2.MetricTarget{
			Type:               autoscalingv2.UtilizationMetricType,
			AverageUtilization: &targetCpuUtilization,
		},
	}
	targetMetrics := []autoscalingv2.MetricSpec{
		{
			Type:     "Resource",
			Resource: &cpuTarget,
		},
	}

	return autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bestie.Name + "-hpa",
			Namespace: bestieDeployment.Namespace,
			Labels:    bestieDeployment.Labels,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       bestieDeployment.Name,
			},
			MinReplicas: bestieDeployment.Spec.Replicas,
			MaxReplicas: *bestie.Spec.MaxReplicas,
			Metrics:     targetMetrics,
		},
	}
}

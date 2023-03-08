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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
)

func NewDatabaseCreatedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    petsv1.DatabaseReady,
		Status:  metav1.ConditionFalse,
		Reason:  "PostgresClusterCreated",
		Message: "Created PostgresCluster resource",
	}
}

func NewDatabaseSeededCondition() metav1.Condition {
	return metav1.Condition{
		Type:    petsv1.DatabaseSeeded,
		Status:  metav1.ConditionFalse,
		Reason:  "DatabaseSeedJobCreated",
		Message: "Database seed job has been created",
	}
}

func NewDeploymentCreatedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    petsv1.DeploymentReady,
		Status:  metav1.ConditionFalse,
		Reason:  "DeploymentCreated",
		Message: "Created Deployment resource",
	}
}

func NewServiceCreatedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    petsv1.ServiceCreated,
		Status:  metav1.ConditionTrue,
		Reason:  "ServiceCreated",
		Message: "The Service has been created",
	}
}

func NewHPACreatedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    petsv1.HPACreated,
		Status:  metav1.ConditionTrue,
		Reason:  "HPACreated",
		Message: "A horizontal pod autoscaling object has been created",
	}
}

func NewRouteCreatedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    petsv1.RouteCreated,
		Status:  metav1.ConditionTrue,
		Reason:  "RouteCreated",
		Message: "The Route has been created",
	}
}

func NewIngressCreatedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    petsv1.IngressCreated,
		Status:  metav1.ConditionTrue,
		Reason:  "ServiceCreated",
		Message: "The Ingress has been created",
	}
}

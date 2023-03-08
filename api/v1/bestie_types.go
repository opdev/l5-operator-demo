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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BestieSpec defines the desired state of Bestie.
type BestieSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster.
	// Important: Run "make" to regenerate code after modifying this file.

	// The size of the deployment
	Size int32 `json:"size"`

	// The bestie app image
	Image string `json:"image,omitempty"`

	// The bestie app version
	Version string `json:"version,omitempty"`

	// MaxReplicas sets an upper bound to the autoscaling feature. If MaxReplicas is set autoscaling is enabled.
	// +optional
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
}

// Status Condition Types.
const (
	DatabaseReady   string = "DatabaseReady"
	DatabaseSeeded  string = "DatabaseSeeded"
	DeploymentReady string = "DeploymentReady"
	ServiceCreated  string = "ServiceCreated"
	HPACreated      string = "HPACreated"
	RouteCreated    string = "RouteCreated"
	IngressCreated  string = "IngressCreated"
)

// BestieStatus defines the observed state of Bestie.
type BestieStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster.
	// Important: Run "make" to regenerate code after modifying this file.

	// List of pods, their status and the image they use
	PodStatus []string `json:"podstatus,omitempty"`

	// Current version deployed
	AppVersion string `json:"appversion,omitempty"`

	// The operators status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Bestie is the Schema for the besties API.
type Bestie struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BestieSpec   `json:"spec,omitempty"`
	Status BestieStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BestieList contains a list of Bestie.
type BestieList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bestie `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bestie{}, &BestieList{})
}

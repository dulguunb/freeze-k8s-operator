/*
Copyright 2025.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DeploymentFreezerSpec defines the desired state of DeploymentFreezer
// +kubebuilder:subresource:status
type DeploymentFreezerSpec struct {
	// Name of the target Deployment
	DeploymentName string `json:"deploymentName"`
	// Namespace of the target Deployment
	DeploymentNamespace string `json:"deploymentNamespace"`
	// Duration in seconds to freeze the deployment
	DurationSeconds int64 `json:"durationSeconds"`
}

// DeploymentFreezerStatus defines the observed state of DeploymentFreezer
// +kubebuilder:subresource:status
type DeploymentFreezerStatus struct {
	// When the deployment was frozen
	FrozenSince *metav1.Time `json:"frozenSince,omitempty"`
	// How long the deployment has been frozen (human readable)
	FrozenDuration string `json:"frozenDuration,omitempty"`
	// Whether the deployment is currently frozen
	IsFrozen bool `json:"isFrozen"`
	// Reason for the current state
	Reason string `json:"reason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DeploymentFreezer is the Schema for the deploymentfreezers API
type DeploymentFreezer struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of DeploymentFreezer
	// +required
	Spec DeploymentFreezerSpec `json:"spec"`

	// status defines the observed state of DeploymentFreezer
	// +optional
	Status DeploymentFreezerStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// DeploymentFreezerList contains a list of DeploymentFreezer
type DeploymentFreezerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeploymentFreezer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeploymentFreezer{}, &DeploymentFreezerList{})
}

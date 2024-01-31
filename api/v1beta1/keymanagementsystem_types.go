/*
Copyright The Ratify Authors.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KeyManagementSystemSpec defines the desired state of KeyManagementSystem
type KeyManagementSystemSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Name of the key management system provider
	Type string `json:"type,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	// Parameters of the key management system
	Parameters runtime.RawExtension `json:"parameters,omitempty"`
}

// KeyManagementSystemStatus defines the observed state of KeyManagementSystem
type KeyManagementSystemStatus struct {
	// Important: Run "make manifests" to regenerate code after modifying this file

	// Is successful in loading certificate/key files
	IsSuccess bool `json:"issuccess"`
	// Error message if operation was unsuccessful
	// +optional
	Error string `json:"error,omitempty"`
	// Truncated error message if the message is too long
	// +optional
	BriefError string `json:"brieferror,omitempty"`
	// The time stamp of last successful certificate/key fetch operation. If operation failed, last fetched time shows the time of error
	// +optional
	LastFetchedTime *metav1.Time `json:"lastfetchedtime,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// provider specific properties of the each individual certificate/key
	// +optional
	Properties runtime.RawExtension `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="IsSuccess",type=boolean,JSONPath=`.status.issuccess`
// +kubebuilder:printcolumn:name="Error",type=string,JSONPath=`.status.brieferror`
// +kubebuilder:printcolumn:name="LastFetchedTime",type=date,JSONPath=`.status.lastfetchedtime`
// KeyManagementSystem is the Schema for the keymanagementsystems API
type KeyManagementSystem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeyManagementSystemSpec   `json:"spec,omitempty"`
	Status KeyManagementSystemStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// KeyManagementSystemList contains a list of KeyManagementSystem
type KeyManagementSystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeyManagementSystem `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeyManagementSystem{}, &KeyManagementSystemList{})
}

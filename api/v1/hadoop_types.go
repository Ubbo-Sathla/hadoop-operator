/*
Copyright 2022.

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

// HadoopSpec defines the desired state of Hadoop
type HadoopSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Quantity of hadoop slave instances
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	ClusterSize int32 `json:"clusterSize"`
}

// HadoopStatus defines the observed state of Hadoop
type HadoopStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Hadoop is the Schema for the hadoops API
type Hadoop struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HadoopSpec   `json:"spec,omitempty"`
	Status HadoopStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HadoopList contains a list of Hadoop
type HadoopList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hadoop `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Hadoop{}, &HadoopList{})
}

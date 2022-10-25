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

// StepJobSpec defines the desired state of StepJob
type StepJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name string `json:"name,omitempty"`
	Steps       []Step `json:"steps,omitempty"`
	CurrentStep string `json:"current_step"`
}

type Step struct {
	Image           string `json:"image"`
	StepName        string `json:"step_name"`
	RetriesCount    int    `json:"retries_count"`
	RetriesInterval int    `json:"retries_interval"`
	CMD             string `json:"cmd"`
	HTTP            string `json:"http"`
}

// StepJobStatus defines the observed state of StepJob
type StepJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// StepJob is the Schema for the stepjobs API
type StepJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StepJobSpec   `json:"spec,omitempty"`
	Status StepJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StepJobList contains a list of StepJob
type StepJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StepJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StepJob{}, &StepJobList{})
}

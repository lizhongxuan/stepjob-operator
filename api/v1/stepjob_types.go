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

type StepCondition string

const (
	PendingStepCondition StepCondition = "Pending"
	RunningStepCondition StepCondition = "Running"
	SuccessStepCondition StepCondition = "Success"
	FailedStepCondition  StepCondition = "Failed"
	NextStepCondition    StepCondition = "Next"
	NoneStepCondition    StepCondition = "None"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StepJobSpec defines the desired state of StepJob
type StepJobSpec struct {
	Steps    []Step `json:"steps"`
	NodeName string `json:"node_name,omitempty"`
	Times    int    `json:"times,omitempty"`
}

type Step struct {
	Image           string   `json:"image,omitempty"`
	StepName        string   `json:"step_name"`
	RetriesCount    int      `json:"retries_count,omitempty"`
	RetriesInterval int      `json:"retries_interval,omitempty"`
	CMD             []string `json:"cmd"`
	HTTP            string   `json:"http,omitempty"`
}

// StepJobStatus defines the observed state of StepJob
type StepJobStatus struct {
	CurrentStep string                `json:"current_step,omitempty"`
	Condition   StepCondition         `json:"condition,omitempty"`
	Steps       map[string]StepStatus `json:"steps"`
	EndTime     int64                 `json:"end_time,omitempty"`
}
type StepStatus struct {
	BeginTime int64         `json:"begin_time,omitempty"`
	EndTime   int64         `json:"end_time,omitempty"`
	Condition StepCondition `json:"condition,omitempty"`
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

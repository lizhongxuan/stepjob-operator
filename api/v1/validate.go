package v1

import (
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *StepJob) Validate() error {
	if len(s.Spec.Steps) == 0 {
		return errors.New("steps is nil")
	}
	if s.Spec.Times == 0 {
		s.Spec.Times = 1
	}
	for i,_ := range s.Spec.Steps{
		if s.Spec.Steps[i].Image == "" {
			s.Spec.Steps[i].Image = "busybox:latest"
		}
		if s.Spec.Steps[i].RetriesCount == 0 {
			s.Spec.Steps[i].RetriesCount = 1
		}
		if s.Spec.Steps[i].RetriesInterval == 0 {
			s.Spec.Steps[i].RetriesCount = 5
		}
	}
	if s.Status.Steps == nil {
		s.Status.Steps = make(map[string]StepStatus)
	}
	return nil
}

func (c *StepJob) ownerReferences() []metav1.OwnerReference {
	controller := true
	if c == nil {
		return []metav1.OwnerReference{}
	}
	return []metav1.OwnerReference{
		{
			UID:        c.ObjectMeta.UID,
			APIVersion: "step.io/v1",
			Kind:       "stepjob",
			Name:       c.Name,
			Controller: &controller,
		},
	}
}
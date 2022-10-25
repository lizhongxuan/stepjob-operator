package v1

import (
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *StepJob) Validate() error {
	if len(s.Spec.Steps) == 0 {
		return errors.New("steps is nil")
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
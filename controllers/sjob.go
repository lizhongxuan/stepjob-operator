package controllers

import (
	"context"
	stepiov1 "github.com/vega-punk/stepjob-operator/api/v1"
	v1 "k8s.io/api/batch/v1"
	v13 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *StepJobReconciler) EnsureSJob(ctx context.Context, nsname types.NamespacedName, currentStep *stepiov1.Step, currentStepIndex int, OwnerRefs []v12.OwnerReference, nodeName string, stepCount int) (stepiov1.StepCondition, error) {
	logger := log.FromContext(ctx)
	sjob := v1.Job{}
	skey := types.NamespacedName{
		Name:      getJobName(nsname.Name, currentStepIndex),
		Namespace: nsname.Namespace,
	}
	// 获取job
	if err := r.Get(ctx, skey, &sjob); err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Get sjob")
			return stepiov1.NoneStepCondition, err
		}

		// 创建job
		newJob := generatedJob(skey.Name, skey.Namespace, OwnerRefs, v13.PodTemplateSpec{
			Spec: v13.PodSpec{
				Containers: []v13.Container{
					{
						Name:    currentStep.StepName,
						Image:   currentStep.Image,
						Command: currentStep.CMD,
					},
				},
				NodeName: nodeName,
			},
		})
		if err := r.Create(ctx, &newJob); err != nil {
			logger.Error(err, "Create sjob")
			return stepiov1.NoneStepCondition, err
		}
		return stepiov1.RunningStepCondition, nil
	}

	// 获取当前step的job
	if sjob.Status.Failed > 0 {
		// job失败
		return stepiov1.FailedStepCondition, nil
	} else if sjob.Status.Succeeded > 0 {
		if stepCount <= currentStepIndex+1 {
			return stepiov1.SuccessStepCondition, nil
		}
		return stepiov1.NextStepCondition, nil
	}
	// 当前job还在运行,继续等待
	return stepiov1.RunningStepCondition, nil
}

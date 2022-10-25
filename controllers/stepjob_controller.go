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

package controllers

import (
	"context"
	"fmt"
	v1 "k8s.io/api/batch/v1"
	v13 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"errors"
	stepiov1 "github.com/vega-punk/stepjob-operator/api/v1"
)

// StepJobReconciler reconciles a StepJob object
type StepJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=step.io.vega-punk.io,resources=stepjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=step.io.vega-punk.io,resources=stepjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=step.io.vega-punk.io,resources=stepjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the StepJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *StepJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Begin stepjob reconcile")

	// get new copy of crd object
	release := stepiov1.StepJob{}
	if err := r.Get(ctx, req.NamespacedName, &release); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch stepjob", "stepjob", &req.NamespacedName)
		return ctrl.Result{}, err
	}

	if err := release.Validate(); err != nil {
		logger.Error(err, "validate err")
		return ctrl.Result{}, err
	}

	if release.Status.Condition == stepiov1.SuccessStepCondition || release.Status.Condition == stepiov1.FailedStepCondition {
		return ctrl.Result{}, nil
	}

	// 获取执行步骤
	currentStep, currentStepIndex, err := getCurrentStep(release.Status.CurrentStep, release.Spec.Steps)
	if err != nil {
		logger.Error(err, "getCurrentStep")
		return ctrl.Result{}, err
	}

	stepStatus, ok := release.Status.Steps[release.Status.CurrentStep]
	if !ok {
		stepStatus = stepiov1.StepStatus{
			BeginTime: time.Now(),
			Condition: stepiov1.RunningStepCondition,
		}
	}

	sjob := v1.Job{}
	skey := types.NamespacedName{
		Name:      getJobName(release.Name, currentStepIndex),
		Namespace: release.Namespace,
	}
	if err := r.Get(ctx, skey, &sjob); err != nil {
		if apierrors.IsNotFound(err) {
			// 创建job
			newJob := generatedJob(skey.Name, skey.Namespace, release.GetOwnerReferences(), v13.PodTemplateSpec{
				Spec: v13.PodSpec{
					Containers: []v13.Container{
						{
							Name:    currentStep.StepName,
							Image:   currentStep.Image,
							Command: currentStep.CMD,
						},
					},
					NodeName: release.Spec.NodeName,
				},
			})
			if err := r.Create(ctx, &newJob); err != nil {
				logger.Error(err, "Create job")
				return ctrl.Result{}, err
			}
			release.Status.Steps[release.Status.CurrentStep] = stepStatus
			if err := r.UpdateStatus(ctx, release, stepiov1.RunningStepCondition); err != nil {
				logger.Error(err, "Status Update")
				return ctrl.Result{}, err
			}
			return ctrl.Result{
				RequeueAfter: 5 * time.Second,
			}, nil
		}
		logger.Error(err, "Get job")
		return ctrl.Result{}, err
	}
	// 获取当前step的job
	if sjob.Status.CompletionTime == nil {
		// 当前job还在运行,继续等待
		return ctrl.Result{
			RequeueAfter: 5 * time.Second,
		}, nil
	} else if sjob.Status.Succeeded == 0 {
		// job失败
		// stop step
		stepStatus.Condition = stepiov1.FailedStepCondition
		release.Status.Steps[release.Status.CurrentStep] = stepStatus
		if err := r.UpdateStatus(ctx, release, stepiov1.FailedStepCondition); err != nil {
			logger.Error(err, "Status Update")
			return ctrl.Result{}, err
		}
		logger.Info("End all step reconcile")
		return ctrl.Result{}, nil
	}

	// job成功
	if currentStepIndex+1 == len(release.Spec.Steps) {
		// end all step
		stepStatus.EndTime = time.Now()
		stepStatus.Condition = stepiov1.SuccessStepCondition
		release.Status.Steps[currentStep.StepName] = stepStatus
		if err := r.UpdateStatus(ctx, release, stepiov1.SuccessStepCondition); err != nil {
			logger.Error(err, "Status Update")
			return ctrl.Result{}, err
		}
		logger.Info("End all step reconcile")
		return ctrl.Result{}, nil
	}

	// next step
	stepStatus.EndTime = time.Now()
	stepStatus.Condition = stepiov1.SuccessStepCondition
	step := release.Spec.Steps[currentStepIndex+1]
	release.Status.CurrentStep = step.StepName
	release.Status.Steps[currentStep.StepName] = stepStatus
	if err := r.UpdateStatus(ctx, release, stepiov1.RunningStepCondition); err != nil {
		logger.Error(err, "Status Update")
		return ctrl.Result{}, err
	}
	logger.Info("Next step reconcile")
	return ctrl.Result{
		Requeue: true,
	}, nil
}

func (r *StepJobReconciler) UpdateStatus(ctx context.Context, release stepiov1.StepJob, condition stepiov1.StepCondition) error {
	if condition == stepiov1.SuccessStepCondition {
		release.Status.EndTime = time.Now()
	}
	release.Status.Condition = condition
	if err := r.Status().Update(ctx, &release); err != nil {
		return err
	}
	return nil
}

func generatedJob(ns, name string, OwnerRefs []v12.OwnerReference, podTemplateSpec v13.PodTemplateSpec) v1.Job {
	return v1.Job{
		ObjectMeta: ctrl.ObjectMeta{
			Name:            name,
			Namespace:       ns,
			OwnerReferences: OwnerRefs,
		},
		Spec: v1.JobSpec{
			Template: podTemplateSpec,
		},
	}
}

func getJobName(releaseName string, stepIndex int) string {
	return fmt.Sprintf("%s-step-%d", releaseName, stepIndex)
}

func getCurrentStep(currentStep string, steps []stepiov1.Step) (*stepiov1.Step, int, error) {
	if currentStep == "" {
		return &steps[0], 0, nil
	}
	for i, s := range steps {
		if s.StepName == currentStep {
			return &steps[i], i, nil
		}
	}
	return nil, 0, errors.New("not find current step")
}

// SetupWithManager sets up the controller with the Manager.
func (r *StepJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&stepiov1.StepJob{}).
		Complete(r)
}

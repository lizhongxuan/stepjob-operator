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
	release.Status.CurrentStep = currentStep.StepName

	// 执行job
	trueVar := true
	condition, err := r.EnsureSJob(ctx, req.NamespacedName, currentStep, currentStepIndex, []v12.OwnerReference{
		{
			APIVersion: release.APIVersion,
			Kind:       release.Kind,
			Name:       release.Name,
			Controller: &trueVar,
			UID:        release.UID,
		},
	}, release.Spec.NodeName, len(release.Spec.Steps))
	if err != nil {
		logger.Error(err, "EnsureSJob")
		return ctrl.Result{}, err
	}

	// 更新状态
	stepStauts, ok := release.Status.Steps[release.Status.CurrentStep]
	if !ok {
		stepStauts = stepiov1.StepStatus{
			BeginTime: time.Now().String(),
		}
	}
	stepStauts.Condition = condition
	release.Status.Steps[release.Status.CurrentStep] = stepStauts
	if condition == stepiov1.NextStepCondition && len(release.Spec.Steps) > currentStepIndex+1 {
		stepStauts.EndTime = time.Now().String()
		release.Status.CurrentStep = release.Spec.Steps[currentStepIndex+1].StepName
	} else if condition == stepiov1.SuccessStepCondition || condition == stepiov1.FailedStepCondition {
		stepStauts.EndTime = time.Now().String()
		release.Status.EndTime = time.Now().String()
	}
	if err := r.Status().Update(ctx, &release); err != nil {
		logger.Error(err, "UpdateStatus")
		return ctrl.Result{}, err
	}

	// 回调
	if condition == stepiov1.RunningStepCondition || condition == stepiov1.PendingStepCondition {
		logger.Info("Wait stepjob reconcile")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	} else if condition == stepiov1.NextStepCondition {
		logger.Info("Next stepjob reconcile")
		return ctrl.Result{Requeue: true}, nil
	}
	logger.Info("End stepjob reconcile")
	return ctrl.Result{}, nil
}

func generatedJob(ns, name string, OwnerRefs []v12.OwnerReference, podTemplateSpec v13.PodTemplateSpec) v1.Job {
	var parallelism int32 = 1
	var completions int32 = 1
	var backoffLimit int32 = 0
	return v1.Job{
		ObjectMeta: ctrl.ObjectMeta{
			Name:            name,
			Namespace:       ns,
			OwnerReferences: OwnerRefs,
		},
		Spec: v1.JobSpec{
			Parallelism:  &parallelism,
			Completions:  &completions,
			BackoffLimit: &backoffLimit,
			Template:     podTemplateSpec,
		},
	}
}

func getJobName(releaseName string, stepIndex int) string {
	return fmt.Sprintf("%s-step-%d", releaseName, stepIndex)
}

func getCurrentStep(currentStepName string, steps []stepiov1.Step) (*stepiov1.Step, int, error) {
	if currentStepName == "" {
		return &steps[0], 0, nil
	}
	for i, s := range steps {
		if s.StepName == currentStepName {
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

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
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	stepiov1 "github.com/vega-punk/step-operator/api/v1"
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
	logger.Info("begin stepjob reconcile")

	// get new copy of crd object
	release := stepiov1.StepJob{}
	if err := r.Get(ctx, req.NamespacedName, &release); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch stepjob", "stepjob", &req.NamespacedName)
		return ctrl.Result{}, err
	}

	if err := release.Validate();err != nil{
		logger.Error(err,"validate err")
		return ctrl.Result{}, err
	}

	// 获取执行步骤



	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StepJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&stepiov1.StepJob{}).
		Complete(r)
}

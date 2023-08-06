/*
Copyright 2023.

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

package dtd

import (
	"context"
	"fmt"

	twinevent "ktwin/operator/internal/resources/event"
	twinservice "ktwin/operator/internal/resources/service"

	kEventing "knative.dev/eventing/pkg/apis/eventing/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dtdv0 "ktwin/operator/api/dtd/v0"
)

// TwinInstanceReconciler reconciles a TwinInstance object
type TwinInstanceReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	TwinService twinservice.TwinService
	TwinEvent   twinevent.TwinEvent
}

//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininstances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TwinInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *TwinInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	twinInstance := &dtdv0.TwinInstance{}
	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, twinInstance)

	// Delete scenario
	if err != nil {
		if errors.IsNotFound(err) {
			return r.deleteTwinInstance(ctx, req, req.NamespacedName)
		}
		logger.Error(err, fmt.Sprintf("Unexpected error while deleting TwinInstance %s", req.Name))
		return ctrl.Result{}, err
	}

	// Get parent TwinInterface
	twinInterface := &dtdv0.TwinInterface{}
	twinInterfaceName := twinInstance.Spec.Interface
	err = r.Get(ctx, types.NamespacedName{Name: twinInterfaceName, Namespace: twinInstance.Namespace}, twinInterface)

	if err != nil {
		logger.Error(err, fmt.Sprintf("Unexpected error while getting TwinInterface %s", twinInterfaceName))
		return ctrl.Result{}, err
	}

	return r.createUpdateTwinInstance(ctx, req, twinInstance, twinInterface)
}

func (r *TwinInstanceReconciler) createUpdateTwinInstance(ctx context.Context, req ctrl.Request, twinInstance *dtdv0.TwinInstance, twinInterface *dtdv0.TwinInterface) (ctrl.Result, error) {
	twinInterfaceName := twinInstance.ObjectMeta.Name

	var resultErrors []error
	//logger := log.FromContext(ctx)

	// Get TwinInterface Trigger

	// Merge with new Trigger

	// Get Event Store Trigger

	// Create Triggers
	// triggers := r.TwinEvent.GetTriggers(twinInstance, twinInterface)

	// for _, trigger := range triggers {
	// 	err := r.Create(ctx, &trigger, &client.CreateOptions{})
	// 	if err != nil && !errors.IsAlreadyExists(err) {
	// 		logger.Error(err, fmt.Sprintf("Error while creating Twin Events %s", twinInterfaceName))
	// 		resultErrors = append(resultErrors, err)
	// 	}
	// }

	if len(resultErrors) > 0 {
		twinInstance.Status.Status = dtdv0.TwinInstancePhaseFailed
		return ctrl.Result{}, resultErrors[0]
	} else {
		twinInstance.Status.Status = dtdv0.TwinInstancePhaseRunning
	}

	twinInstance.Labels = map[string]string{
		"ktwin/twin-interface": twinInterfaceName,
	}

	// Update Status for Running or Failed
	_, err := r.updateTwinInstance(ctx, req, twinInstance)

	if err != nil {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *TwinInstanceReconciler) updateTwinInstance(ctx context.Context, req ctrl.Request, twinInstance *dtdv0.TwinInstance) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	err := r.Update(ctx, twinInstance, &client.UpdateOptions{})

	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while updating TwinInstance %s", twinInstance.ObjectMeta.Name))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TwinInstanceReconciler) deleteTwinInstance(ctx context.Context, req ctrl.Request, namespacedName types.NamespacedName) (ctrl.Result, error) {
	var errorsResult []error
	logger := log.FromContext(ctx)

	// Delete Triggers
	deletionTriggerLabels := r.TwinEvent.GetTriggersDeletionFilterCriteria(namespacedName)

	triggerList := kEventing.TriggerList{}
	triggerListOptions := []client.ListOption{
		client.InNamespace(namespacedName.Namespace),
		client.MatchingLabels(deletionTriggerLabels),
	}

	err := r.List(ctx, &triggerList, triggerListOptions...)

	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while getting triggers %s", namespacedName.Name))
		return ctrl.Result{}, err
	}

	for _, trigger := range triggerList.Items {
		err := r.Delete(ctx, &trigger, &client.DeleteOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				logger.Error(err, fmt.Sprintf("Error while deleting trigger %s", namespacedName.Name))
				errorsResult = append(errorsResult, err)
			}
		}
	}

	if len(errorsResult) > 0 {
		return ctrl.Result{}, errorsResult[0]
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TwinInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dtdv0.TwinInstance{}).
		Complete(r)
}

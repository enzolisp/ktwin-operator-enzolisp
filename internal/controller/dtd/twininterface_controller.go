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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dtdv0 "ktwin/operator/api/dtd/v0"
)

// TwinInterfaceReconciler reconciles a TwinInterface object
type TwinInterfaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininterfaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininterfaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininterfaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TwinInterface object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *TwinInterfaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	twinInterface := &dtdv0.TwinInterface{}
	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, twinInterface)

	// Delete scenario
	if err != nil {
		if errors.IsNotFound(err) {
			return r.deleteTwinInterface(ctx, req, req.NamespacedName)
		}
		logger.Error(err, fmt.Sprintf("Unexpected error while deleting TwinInstance %s", req.Name))
		return ctrl.Result{}, err
	}

	// Create Mew Schema in Event Store

	return r.createUpdateTwinInterface(ctx, req, twinInterface)
}

func (r *TwinInterfaceReconciler) deleteTwinInterface(ctx context.Context, req ctrl.Request, namespacedName types.NamespacedName) (ctrl.Result, error) {
	var errorsResult []error

	if len(errorsResult) > 0 {
		return ctrl.Result{}, errorsResult[0]
	}

	return ctrl.Result{}, nil
}

func (r *TwinInterfaceReconciler) createUpdateTwinInterface(ctx context.Context, req ctrl.Request, twinInstance *dtdv0.TwinInterface) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TwinInterfaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dtdv0.TwinInterface{}).
		Complete(r)
}

/*
Copyright 2021.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	astrolabeiov1 "github.com/dsu-igeek/astrolabe-demo/cmd/astrolabe-controller/api/v1"
)

// PSQLRestoreReconciler reconciles a PSQLRestore object
type PSQLRestoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Pem astrolabe.ProtectedEntityManager
}

//+kubebuilder:rbac:groups=astrolabe.io,resources=psqlrestores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=astrolabe.io,resources=psqlrestores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=astrolabe.io,resources=psqlrestores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PSQLRestore object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *PSQLRestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Getting PSQLRestore")
	logger.Info("Getting PSQLRestore")
	restore := &astrolabeiov1.PSQLRestore{}
	if err := r.Client.Get(ctx, req.NamespacedName, restore); err != nil {
		logger.Info("Unable to find PSQLRestore - likely deleted")
		// TODO: Call through to deletion of snapshot if the DeletionPolicy allows for it
		// Do we have access to the object at this point to be able to delete?
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	patcher, err := patch.NewHelper(restore, r.Client)
	if err != nil {
		logger.Error(err, "unable to initialize patch helper")
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to Patch the object and status after each reconciliation.
		if err := patcher.Patch(ctx, restore); err != nil {
			logger.Error(err, "Error updating PSQL Restore")
			return
		}
	}()

	// Check for the status, if empty, start restore
	if restore.Status == (astrolabeiov1.PSQLRestoreStatus{}) {
		var snapshotIDstr string
		if restore.Spec.Source.SnapshotID != nil {
			snapshotIDstr = *restore.Spec.Source.SnapshotID
		} else if restore.Spec.Source.Name != (corev1.LocalObjectReference{}) {
			return ctrl.Result{Requeue: false}, fmt.Errorf("unsupported source")
		} else {
			return ctrl.Result{Requeue: false}, fmt.Errorf("spec.peid field missing")
		}

		protectedEntityID := astrolabe.NewProtectedEntityIDWithSnapshotID("psql", restore.Spec.Name, astrolabe.NewProtectedEntitySnapshotID(snapshotIDstr))
		protectedEntity, err := r.Pem.GetProtectedEntity(ctx, protectedEntityID)
		if err != nil {
			logger.Error(err, "Error getting ProtectedEntity")
			errTime := metav1.Now()
			errMessage := fmt.Sprintf("error getting PSQL Protected Entity %q with snapshot %q: %v", protectedEntityID, snapshotIDstr, err)
			restore.Status.Error = &astrolabeiov1.PSQLRestoreError{
				Time:    &errTime,
				Message: &errMessage,
			}
			return ctrl.Result{Requeue: false}, err
		}

		logger.Info("Retrieved ProtectedEntity", "protectedEntity", protectedEntity.GetID().String())
		snapshotID, err := protectedEntity.Snapshot(ctx, map[string]map[string]interface{}{})

		logger.Info("Created ProtectedEntity snapshot", "snapshot id", snapshotID.String())
		if err != nil {
			logger.Error(err, "Creating ProtectedEntity snapshot")
			errTime := metav1.Now()
			errMessage := fmt.Sprintf("error snapshotting PSQL Protected Entity %q: %v", protectedEntityID, err)
			restore.Status.Error = &astrolabeiov1.PSQLRestoreError{
				Time:    &errTime,
				Message: &errMessage,
			}
		} else {
			// restore was successful, update the status
		}
	} else {
		// We have already done a reconciliation of this object, need to check the status and update accordingly
		// We are currently doing a blocking action on the snapshot operation
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PSQLRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&astrolabeiov1.PSQLRestore{}).
		Complete(r)
}

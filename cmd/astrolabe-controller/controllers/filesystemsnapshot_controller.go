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
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/cluster-api/util/patch"

	astrolabeiov1 "github.com/dsu-igeek/astrolabe-demo/cmd/astrolabe-controller/api/v1"
)

// PSQLSnapshotReconciler reconciles a FileSystemSnapshot object
type PSQLSnapshotReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Pem astrolabe.ProtectedEntityManager
}

const fsSnapshotFinalizer = "astrolabe.io/finalizer"

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
//+kubebuilder:rbac:groups=astrolabe.io,resources=filesystemsnapshots,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=astrolabe.io,resources=filesystemsnapshots/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=astrolabe.io,resources=filesystemsnapshots/finalizers,verbs=update
func (r *PSQLSnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Getting FileSystemSnapshot")
	fsSnapshot := &astrolabeiov1.FileSystemSnapshot{}
	if err := r.Client.Get(ctx, req.NamespacedName, fsSnapshot); err != nil {
		logger.Info("Unable to find FileSystemSnapshot - likely deleted")
		// TODO: Call through to deletion of snapshot if the DeletionPolicy allows for it
		// Do we have access to the object at this point to be able to delete?
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	patcher, err := patch.NewHelper(fsSnapshot, r.Client)
	if err != nil {
		logger.Error(err, "unable to initialize patch helper")
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to Patch the object and status after each reconciliation.
		if err := patcher.Patch(ctx, fsSnapshot); err != nil {
			logger.Error(err, "Error updating FS Snapshot")
			return
		}
	}()

	if fsSnapshot.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(fsSnapshot.GetFinalizers(), fsSnapshotFinalizer) {
			controllerutil.AddFinalizer(fsSnapshot, fsSnapshotFinalizer)
		}
	} else {
		// The object is being deleted
		if containsString(fsSnapshot.GetFinalizers(), fsSnapshotFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			// TODO need to fetch the PE here to be able to delete the snapshot
			if err := r.deleteSnapshot(fsSnapshot); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(fsSnapshot, fsSnapshotFinalizer)
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Check for the status, if empty, start snapshot
	if fsSnapshot.Status == (astrolabeiov1.FileSystemSnapshotStatus{}) {
		var peIDstr string
		if fsSnapshot.Spec.Source.PEID != nil {
			peIDstr = *fsSnapshot.Spec.Source.PEID
		} else if fsSnapshot.Spec.Source.Name != (corev1.LocalObjectReference{}) {
			return ctrl.Result{Requeue: false}, fmt.Errorf("unsupported source")
		} else {
			return ctrl.Result{Requeue: false}, fmt.Errorf("spec.peid field missing")
		}

		protectedEntityID := astrolabe.NewProtectedEntityID("psql", peIDstr)
		protectedEntity, err := r.Pem.GetProtectedEntity(ctx, protectedEntityID)
		if err != nil {
			logger.Error(err, "Error getting ProtectedEntity")
			errTime := metav1.Now()
			errMessage := fmt.Sprintf("error getting FileSystem Protected Entity %q: %v", protectedEntityID, err)
			fsSnapshot.Status.Error = &astrolabeiov1.FileSystemSnapshotError{
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
			errMessage := fmt.Sprintf("error snapshotting FileSystem Protected Entity %q: %v", protectedEntityID, err)
			fsSnapshot.Status.Error = &astrolabeiov1.FileSystemSnapshotError{
				Time:    &errTime,
				Message: &errMessage,
			}
		} else {
			// snapshot was successful, update the status
			snapshotIDStr := snapshotID.String()
			logger.Info("updating the status")
			readyToUse := true
			fsSnapshot.Status.SnapshotID = &snapshotIDStr
			fsSnapshot.Status.ReadyToUse = &readyToUse
		}
	} else {
		// We have already done a reconciliation of this object, need to check the status and update accordingly
		// We are currently doing a blocking action on the snapshot operation
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PSQLSnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&astrolabeiov1.FileSystemSnapshot{}).
		Complete(r)
}

// deleteSnapshot deletes the given snapshot
func (r *PSQLSnapshotReconciler) deleteSnapshot(s *astrolabeiov1.FileSystemSnapshot) error {
	return nil
}

// containsString returns true if the given string exists in the slice
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PSQLSnapshot is the Schema for the PSQLSnapshots API
type PSQLSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec PSQLSnapshotSpec `json:"spec,omitempty" protobuf:"bytes,name=spec"`

	Status PSQLSnapshotStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

//+kubebuilder:object:root=true

// PSQLSnapshotList contains a list of PSQLSnapshot
type PSQLSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PSQLSnapshot `json:"items"`
}

// PSQLSnapshotSpec defines the desired state of PSQLSnapshot
type PSQLSnapshotSpec struct {
	// source specifies where a snapshot will be created from.
	// This field is immutable after creation.
	// Required.
	Source PSQLSnapshotSource `json:"source" protobuf:"bytes,1,opt,name=source"`

	// deletionPolicy determines whether this PSQLSnapshotContent and its physical snapshot on
	// the underlying storage system should be deleted when its bound PSQLSnapshot is deleted.
	// Supported values are "Retain" and "Delete".
	// "Retain" means that the PSQLSnapshotContent and its physical snapshot on underlying storage system are kept.
	// "Delete" means that the PSQLSnapshotContent and its physical snapshot on underlying storage system are deleted.
	// For dynamically provisioned snapshots, this field will automatically be filled in by the
	// CSI snapshotter sidecar with the "DeletionPolicy" field defined in the corresponding
	// PSQLSnapshotClass.
	// For pre-existing snapshots, users MUST specify this field when creating the
	//  PSQLSnapshotContent object.
	// Required.
	DeletionPolicy DeletionPolicy `json:"deletionPolicy"`
}

// PSQLSnapshotSource represents the source that should be used to take the snapshot
// Can reference an object of the type in the namespace or a ProtectedEntity ID
type PSQLSnapshotSource struct {
	// Name specifies the name of an astrolabe compatible object from which a snapshot should be created.
	// This field is immutable.
	// +optional
	Name corev1.LocalObjectReference `json:"name,omitempty" protobuf:"bytes,2,opt,name=name"`

	// PEID is the ID for the PSQL ProtectedEntity from which a snapshot should be created.
	// This field is immutable.
	// +optional
	PEID *string `json:"peid" protobuf:"bytes,1,name=peid"`
}

// PSQLSnapshotStatus defines the observed state of PSQLSnapshot
type PSQLSnapshotStatus struct {
	// SnapshotID is the Snapshot ID for the Protected Entity
	SnapshotID *string `json:"snapshotID,omitempty" protobuf:"bytes,1,opt,name=snapshotID"`

	// creationTime is the timestamp when the point-in-time snapshot is taken
	// by the underlying storage system.
	// In dynamic snapshot creation case, this field will be filled in by the
	// snapshot controller with the "creation_time" value returned from CSI
	// "CreateSnapshot" gRPC call.
	// TODO(bridget): Update the above to reference astrolabe
	// TODO(bridget): Also check - does Astrolabe record snapshot time or is that up to each implementation?
	// For a pre-existing snapshot, this field will be filled with the "creation_time"
	// value returned from the CSI "ListSnapshots" gRPC call if the driver supports it.
	// If not specified, it may indicate that the creation time of the snapshot is unknown.
	// +optional
	// CreationTime *metav1.Time `json:"creationTime,omitempty"`

	// RestoreSize represents the minimum size of volume required to create a volume
	// from this snapshot.
	// In dynamic snapshot creation case, this field will be filled in by the
	// snapshot controller with the "size_bytes" value returned from CSI
	// "CreateSnapshot" gRPC call.
	// For a pre-existing snapshot, this field will be filled with the "size_bytes"
	// value returned from the CSI "ListSnapshots" gRPC call if the driver supports it.
	// When restoring a volume from this snapshot, the size of the volume MUST NOT
	// be smaller than the restoreSize if it is specified, otherwise the restoration will fail.
	// If not specified, it indicates that the size is unknown.
	// +optional
	// RestoreSize *resource.Quantity `json:"restoreSize,omitempty"`

	// ReadyToUse indicates if the snapshot is ready to be used to restore a volume.
	// In dynamic snapshot creation case, this field will be filled in by the
	// snapshot controller with the "ready_to_use" value returned from CSI
	// "CreateSnapshot" gRPC call.
	// For a pre-existing snapshot, this field will be filled with the "ready_to_use"
	// value returned from the CSI "ListSnapshots" gRPC call if the driver supports it,
	// otherwise, this field will be set to "True".
	// If not specified, it means the readiness of a snapshot is unknown.
	// +optional
	ReadyToUse *bool `json:"readyToUse,omitempty" protobuf:"bytes,2,opt,name=readyToUse"`

	// Error is the last observed error during snapshot creation, if any.
	// This field could be helpful to upper level controllers(i.e., application controller)
	// to decide whether they should continue on waiting for the snapshot to be created
	// based on the type of error reported.
	// The snapshot controller will keep retrying when an error occurrs during the
	// snapshot creation. Upon success, this error field will be cleared.
	// +optional
	Error *PSQLSnapshotError `json:"error,omitempty" protobuf:"bytes,3,opt,name=error,casttype=PSQLSnapshotError"`
}

// DeletionPolicy describes a policy for end-of-life maintenance of volume snapshot contents
// +kubebuilder:validation:Enum=Delete;Retain
type DeletionPolicy string

const (
	// PSQLSnapshotContentDelete means the snapshot will be deleted from the
	// underlying storage system on release from its PSQL snapshot.
	PSQLSnapshotContentDelete DeletionPolicy = "Delete"

	// PSQLSnapshotContentRetain means the snapshot will be left in its current
	// state on release from its PSQL snapshot.
	PSQLSnapshotContentRetain DeletionPolicy = "Retain"
)

// PSQLSnapshotError describes an error encountered during snapshot creation.
type PSQLSnapshotError struct {
	// time is the timestamp when the error was encountered.
	// +optional
	Time *metav1.Time `json:"time,omitempty"`

	// message is a string detailing the encountered error during snapshot
	// creation if specified.
	// NOTE: message may be logged, and it should not contain sensitive
	// information.
	// +optional
	Message *string `json:"message,omitempty"`
}

func init() {
	SchemeBuilder.Register(&PSQLSnapshot{}, &PSQLSnapshotList{})
}

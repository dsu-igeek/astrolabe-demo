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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PSQLRestoreSpec defines the desired state of PSQLRestore
type PSQLRestoreSpec struct {
	// Name is the name of the resource to restore
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// source specifies where a snapshot will be created from.
	// This field is immutable after creation.
	// Required.
	Source PSQLRestoreSource `json:"source" protobuf:"bytes,1,opt,name=source"`
}

// PSQLRestoreSource represents the source that should be used to restore from
// Can reference a snapshot object in the namespace or a ProtectedEntity Snapshot ID
type PSQLRestoreSource struct {
	// Name specifies the name of an psqlsnapshot object to restore from.
	// This field is immutable.
	// +optional
	Name corev1.LocalObjectReference `json:"name,omitempty" protobuf:"bytes,2,opt,name=name"`

	// SnapshotID is the snapshot ID for the PSQL ProtectedEntity to restore from
	// This field is immutable.
	// +optional
	SnapshotID *string `json:"snapshotId" protobuf:"bytes,1,name=snapshotId"`
}

// PSQLRestoreStatus defines the observed state of PSQLRestore
type PSQLRestoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Error is the last observed error during snapshot creation, if any.
	// This field could be helpful to upper level controllers(i.e., application controller)
	// to decide whether they should continue on waiting for the snapshot to be created
	// based on the type of error reported.
	// The snapshot controller will keep retrying when an error occurrs during the
	// snapshot creation. Upon success, this error field will be cleared.
	// +optional
	Error *PSQLRestoreError `json:"error,omitempty" protobuf:"bytes,3,opt,name=error,casttype=PSQLSnapshotError"`
}

// PSQLRestoreError describes an error encountered during restore from a snapshot
type PSQLRestoreError struct {
	// time is the timestamp when the error was encountered.
	// +optional
	Time *metav1.Time `json:"time,omitempty"`

	// message is a string detailing the encountered error during restore
	// NOTE: message may be logged, and it should not contain sensitive
	// information.
	// +optional
	Message *string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PSQLRestore is the Schema for the psqlrestores API
type PSQLRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PSQLRestoreSpec   `json:"spec,omitempty"`
	Status PSQLRestoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PSQLRestoreList contains a list of PSQLRestore
type PSQLRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PSQLRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PSQLRestore{}, &PSQLRestoreList{})
}

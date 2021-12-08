//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileSystemSnapshot) DeepCopyInto(out *FileSystemSnapshot) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileSystemSnapshot.
func (in *FileSystemSnapshot) DeepCopy() *FileSystemSnapshot {
	if in == nil {
		return nil
	}
	out := new(FileSystemSnapshot)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FileSystemSnapshot) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileSystemSnapshotError) DeepCopyInto(out *FileSystemSnapshotError) {
	*out = *in
	if in.Time != nil {
		in, out := &in.Time, &out.Time
		*out = (*in).DeepCopy()
	}
	if in.Message != nil {
		in, out := &in.Message, &out.Message
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileSystemSnapshotError.
func (in *FileSystemSnapshotError) DeepCopy() *FileSystemSnapshotError {
	if in == nil {
		return nil
	}
	out := new(FileSystemSnapshotError)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileSystemSnapshotList) DeepCopyInto(out *FileSystemSnapshotList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FileSystemSnapshot, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileSystemSnapshotList.
func (in *FileSystemSnapshotList) DeepCopy() *FileSystemSnapshotList {
	if in == nil {
		return nil
	}
	out := new(FileSystemSnapshotList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FileSystemSnapshotList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileSystemSnapshotSource) DeepCopyInto(out *FileSystemSnapshotSource) {
	*out = *in
	out.Name = in.Name
	if in.PEID != nil {
		in, out := &in.PEID, &out.PEID
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileSystemSnapshotSource.
func (in *FileSystemSnapshotSource) DeepCopy() *FileSystemSnapshotSource {
	if in == nil {
		return nil
	}
	out := new(FileSystemSnapshotSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileSystemSnapshotSpec) DeepCopyInto(out *FileSystemSnapshotSpec) {
	*out = *in
	in.Source.DeepCopyInto(&out.Source)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileSystemSnapshotSpec.
func (in *FileSystemSnapshotSpec) DeepCopy() *FileSystemSnapshotSpec {
	if in == nil {
		return nil
	}
	out := new(FileSystemSnapshotSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileSystemSnapshotStatus) DeepCopyInto(out *FileSystemSnapshotStatus) {
	*out = *in
	if in.SnapshotID != nil {
		in, out := &in.SnapshotID, &out.SnapshotID
		*out = new(string)
		**out = **in
	}
	if in.ReadyToUse != nil {
		in, out := &in.ReadyToUse, &out.ReadyToUse
		*out = new(bool)
		**out = **in
	}
	if in.Error != nil {
		in, out := &in.Error, &out.Error
		*out = new(FileSystemSnapshotError)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileSystemSnapshotStatus.
func (in *FileSystemSnapshotStatus) DeepCopy() *FileSystemSnapshotStatus {
	if in == nil {
		return nil
	}
	out := new(FileSystemSnapshotStatus)
	in.DeepCopyInto(out)
	return out
}

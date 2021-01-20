/*
Copyright 2020 The Kubernetes Authors.

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

package ignition

import (
	"sigs.k8s.io/cluster-api/exp/kubeadm-ignition/api/v1alpha4"
	//     types "sigs.k8s.io/cluster-api/exp/kubeadm-ignition/types/v1beta1"
)

const (
	DefaultFileMode = 0644
	DefaultDirMode  = 0755
)

type Node struct {
	Files    []v1alpha4.File
	Services []ServiceUnit
	Version  string
}

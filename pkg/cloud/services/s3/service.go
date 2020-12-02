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

package s3

import (
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"sigs.k8s.io/cluster-api-provider-aws/pkg/cloud"
	"sigs.k8s.io/cluster-api-provider-aws/pkg/cloud/scope"
	kubeadmignition "sigs.k8s.io/cluster-api-provider-aws/pkg/cloud/services/s3/ignition"
)

var (
	userDataBucket string = "ignition-userdata-bucket"
	userDataDir    string = "ignition-userdata-dir"
)

// Service holds a collection of interfaces.
// The interfaces are broken down like this to group functions together.
// One alternative is to have a large list of functions from the ec2 client.
type Service struct {
	scope           cloud.ClusterScoper
	S3Client        s3iface.S3API
	IgnitionFactory *kubeadmignition.Factory
	Node            *kubeadmignition.Node
}

// NewService returns a new service given the api clients.
func NewService(secretsScope cloud.ClusterScoper) *Service {
	templateBackend, err := kubeadmignition.NewS3TemplateBackend(userDataDir, userDataBucket)
	if err != nil {
		return nil
	}

	s3Client := scope.NewS3Client(secretsScope, secretsScope, secretsScope.InfraCluster())
	templateBackend.S3Client = s3Client

	return &Service{
		scope:           secretsScope,
		S3Client:        s3Client,
		IgnitionFactory: kubeadmignition.NewFactory(templateBackend),
		Node:            NewFakeNode(),
	}
}

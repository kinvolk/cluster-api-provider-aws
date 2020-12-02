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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"sigs.k8s.io/cluster-api-provider-aws/pkg/cloud/scope"
	kubeadmignition "sigs.k8s.io/cluster-api-provider-aws/pkg/cloud/services/s3/ignition"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/types/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/exp/kubeadm-ignition/api/v1alpha4"
)

var (
	bucketName      = "capi-ignition"
	userdataDirName = "node-userdata"
)

// UserData creates a multi-part MIME document including a script boothook to
// download userdata from AWS Systems Manager and then restart ignition, and an include part
// specifying the on disk location of the new userdata
func (s *Service) UserData(secretPrefix string, chunks int32, region string, endpoints []scope.ServiceEndpoint) ([]byte, error) {
	ignData, err := s.IgnitionFactory.GenerateUserData(s.Node)
	if err != nil {
		fmt.Printf("failed to generate ignition for bootstrap control plane\n")
		return nil, err
	}

	userData, err := json.Marshal(ignData)
	if err != nil {
		fmt.Printf("failed to marshal ignition file\n")
		return nil, err
	}

	return userData, nil
}

// Create stores data in AWS SSM for a given machine, chunking at 4kb per secret. The prefix of the secret
// ARN and the number of chunks are returned.
func (s *Service) Create(m *scope.MachineScope, data []byte) (string, int32, error) {
	if _, err := s.IgnitionFactory.ApplyConfig(s.Node, data); err != nil {
		fmt.Printf("failed to apply ignition config for bootstrap control plane\n")
		return "", 0, err
	}

	return "", 0, nil
}

// Delete the secret belonging to a machine from AWS SSM
func (s *Service) Delete(m *scope.MachineScope) error {
	var err error
	_, err = s.S3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.IgnitionFactory.GetUserDataBucket()),
		Key:    aws.String(s.IgnitionFactory.GetFilePath()),
	})
	if err != nil {
		return fmt.Errorf("failed to get object from bucket")
	}

	return err
}

func NewFakeNode() *kubeadmignition.Node {
	kubeadmConfig := bootstrapv1.KubeadmIgnitionConfig{}

	initdata, err := kubeadmv1beta1.ConfigurationToYAML(kubeadmConfig.Spec.InitConfiguration)
	if err != nil {
		fmt.Printf("failed to marshal init configuration\n")
		return nil
	}

	clusterdata, err := kubeadmv1beta1.ConfigurationToYAML(kubeadmConfig.Spec.ClusterConfiguration)
	if err != nil {
		fmt.Printf("failed to marshal cluster configuration\n")
		return nil
	}

	verbosityFlag := fmt.Sprintf("--v %s", strconv.Itoa(int(*kubeadmConfig.Spec.Verbosity)))

	return &kubeadmignition.Node{
		Files: append([]bootstrapv1.File{}, bootstrapv1.File{
			Path:        kubeadmignition.KubeadmIgnitionConfigPath,
			Permissions: "0640",
			Content:     strings.Join([]string{initdata, clusterdata}, "\n---\n"),
		}),
		Services: []kubeadmignition.ServiceUnit{
			{
				Name:    "kubeinit.service",
				Content: fmt.Sprintf(kubeadmignition.InitUnitTemplate, verbosityFlag, kubeadmignition.KubeadmIgnitionConfigPath),
				Enabled: true,
				Dropins: kubeadmignition.GetCommandsDropins(kubeadmConfig.Spec.PreKubeadmCommands, kubeadmConfig.Spec.PostKubeadmCommands),
			},
		},
		Version: kubeadmConfig.Spec.ClusterConfiguration.KubernetesVersion,
	}
}

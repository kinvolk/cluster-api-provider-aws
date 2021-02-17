/*
Copyright 2021 The Kubernetes Authors.

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
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/pkg/errors"

	"sigs.k8s.io/cluster-api-provider-aws/pkg/cloud/scope"
)

// Service holds a collection of interfaces.
// The interfaces are broken down like this to group functions together.
// One alternative is to have a large list of functions from the ec2 client.
type Service struct {
	scope     scope.S3Scope
	s3Client  s3iface.S3API
	stsClient stsiface.STSAPI
}

// NewService returns a new service given the api clients.
func NewService(s3Scope scope.S3Scope) *Service {
	s3Client := scope.NewS3Client(s3Scope, s3Scope, s3Scope.InfraCluster())
	stsClient := scope.NewSTSClient(s3Scope, s3Scope, s3Scope.InfraCluster())

	return &Service{
		scope:     s3Scope,
		s3Client:  s3Client,
		stsClient: stsClient,
	}
}

func (s *Service) bucketManagementEnabled() bool {
	return s.scope.Bucket().Enabled
}

func (s *Service) DeleteBucket() error {
	if !s.bucketManagementEnabled() {
		return nil
	}

	s.scope.Info("Deleting S3 Bucket", "name", s.bucketName())

	if _, err := s.s3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(s.bucketName()),
	}); err != nil {
		return errors.Wrap(err, "deleting S3 bucket")
	}

	return nil
}

func (s *Service) ReconcileBucket() error {
	if !s.bucketManagementEnabled() {
		return nil
	}

	if _, err := s.s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(s.bucketName()),
	}); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			// If bucket already exists, all good.
			// TODO: This will fail if bucket is shared with other cluster.
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				break
			default:
				return errors.Wrap(aerr, "creating S3 bucket")
			}
		} else {
			return errors.Wrap(err, "creating S3 bucket")
		}
	}

	accountID, err := s.stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return errors.Wrap(err, "getting account ID")
	}

	// Create a policy using map interface. Filling in the bucket as the
	// resource.
	readOnlyAnonUserPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Sid":    "Stmt1613551032800",
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					// TODO: Document that if user specifies their own IAM role for nodes, they must also include access to the bucket
					// using user role.
					"AWS": fmt.Sprintf("arn:aws:iam::%s:role/nodes.cluster-api-provider-aws.sigs.k8s.io", *accountID.Account),
				},
				"Action": []string{
					"s3:GetObject",
				},
				"Resource": fmt.Sprintf("arn:aws:s3:::%s/node/*", s.bucketName()),
			},
			{
				"Sid":    "Stmt1613551032801",
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					"AWS": fmt.Sprintf("arn:aws:iam::%s:role/control-plane.cluster-api-provider-aws.sigs.k8s.io", *accountID.Account),
				},
				"Action": []string{
					"s3:GetObject",
				},
				"Resource": fmt.Sprintf("arn:aws:s3:::%s/control-plane/*", s.bucketName()),
			},
		},
	}

	// Marshal the policy into a JSON value so that it can be sent to S3.
	policy, err := json.Marshal(readOnlyAnonUserPolicy)
	if err != nil {
		return errors.Wrap(err, "building bucket policy")
	}

	if _, err := s.s3Client.PutBucketPolicy(&s3.PutBucketPolicyInput{
		Bucket: aws.String(s.bucketName()),
		Policy: aws.String(string(policy)),
	}); err != nil {
		return errors.Wrap(err, "creating S3 bucket policy")
	}

	return nil
}

func (s *Service) bucketName() string {
	if name := s.scope.Bucket().Name; name != "" {
		return name
	}

	return s.scope.KubernetesClusterName()
}

func (s *Service) bootstrapDataKey(m *scope.MachineScope) string {
	// Use machine name as object key.
	return path.Join(m.Role(), m.Name())
}

func (s *Service) Delete(m *scope.MachineScope) error {
	if _, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName()),
		Key:    aws.String(s.bootstrapDataKey(m)),
	}); err != nil {
		return errors.Wrap(err, "deleting object")
	}

	return nil
}

func (s *Service) Create(m *scope.MachineScope, data []byte) (string, error) {
	bucket := s.bucketName()
	key := s.bootstrapDataKey(m)

	if _, err := s.s3Client.PutObject(&s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(bytes.NewReader(data)),
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}); err != nil {
		return "", errors.Wrap(err, "putting object")
	}

	objectURL := &url.URL{
		Scheme: "s3",
		Host:   bucket,
		Path:   key,
	}

	return objectURL.String(), nil
}

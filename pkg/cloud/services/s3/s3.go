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

	infrav1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
	"sigs.k8s.io/cluster-api-provider-aws/pkg/cloud/scope"
)

// Service holds a collection of interfaces.
// The interfaces are broken down like this to group functions together.
// One alternative is to have a large list of functions from the ec2 client.
type Service struct {
	scope     scope.S3Scope
	S3Client  s3iface.S3API
	stsClient stsiface.STSAPI
}

// NewService returns a new service given the api clients.
func NewService(s3Scope scope.S3Scope) *Service {
	s3Client := scope.NewS3Client(s3Scope, s3Scope, s3Scope, s3Scope.InfraCluster())
	stsClient := scope.NewSTSClient(s3Scope, s3Scope, s3Scope, s3Scope.InfraCluster())

	return &Service{
		scope:     s3Scope,
		S3Client:  s3Client,
		stsClient: stsClient,
	}
}

func (s *Service) ReconcileBucket() error {
	if !s.bucketManagementEnabled() {
		return nil
	}

	if err := s.createBucketIfNotExist(); err != nil {
		return errors.Wrap(err, "ensuring bucket exists")
	}

	if err := s.ensureBucketPolicy(); err != nil {
		return errors.Wrap(err, "ensuring bucket policy")
	}

	return nil
}

func (s *Service) DeleteBucket() error {
	if !s.bucketManagementEnabled() {
		return nil
	}

	s.scope.Info("Deleting S3 Bucket", "name", s.bucketName())

	if _, err := s.S3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(s.bucketName()),
	}); err != nil {
		return errors.Wrap(err, "deleting S3 bucket")
	}

	return nil
}

func (s *Service) Create(m *scope.MachineScope, data []byte) (string, error) {
	if !s.bucketManagementEnabled() {
		return "", errors.New("requested object creation but bucket management is not enabled")
	}

	if m == nil {
		return "", errors.New("machine scope can't be nil")
	}

	if len(data) == 0 {
		return "", errors.New("got empty data")
	}

	bucket := s.bucketName()
	key := s.bootstrapDataKey(m)

	if _, err := s.S3Client.PutObject(&s3.PutObjectInput{
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

func (s *Service) Delete(m *scope.MachineScope) error {
	if !s.bucketManagementEnabled() {
		return errors.New("requested object creation but bucket management is not enabled")
	}

	if m == nil {
		return errors.New("machine scope can't be nil")
	}

	if _, err := s.S3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName()),
		Key:    aws.String(s.bootstrapDataKey(m)),
	}); err != nil {
		return errors.Wrap(err, "deleting object")
	}

	return nil
}

func (s *Service) createBucketIfNotExist() error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(s.bucketName()),
	}

	_, err := s.S3Client.CreateBucket(input)
	if err == nil {
		return nil
	}

	aerr, ok := err.(awserr.Error)
	if !ok {
		return errors.Wrap(err, "creating S3 bucket")
	}

	switch aerr.Code() {
	// If bucket already exists, all good.
	// TODO: This will fail if bucket is shared with other cluster.
	case s3.ErrCodeBucketAlreadyOwnedByYou:
		return nil
	default:
		return errors.Wrap(aerr, "creating S3 bucket")
	}
}

func (s *Service) ensureBucketPolicy() error {
	bucketPolicy, err := s.bucketPolicy()
	if err != nil {
		return errors.Wrap(err, "generating Bucket policy")
	}

	input := &s3.PutBucketPolicyInput{
		Bucket: aws.String(s.bucketName()),
		Policy: aws.String(bucketPolicy),
	}

	if _, err := s.S3Client.PutBucketPolicy(input); err != nil {
		return errors.Wrap(err, "creating S3 bucket policy")
	}

	return nil
}

func (s *Service) bucketPolicy() (string, error) {
	accountID, err := s.stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", errors.Wrap(err, "getting account ID")
	}

	readOnlyAnonUserPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Sid":    "Stmt1613551032800",
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					// TODO: Document that if user specifies their own IAM role for nodes, they must also include
					// access to the bucket using user role.
					"AWS": fmt.Sprintf("arn:aws:iam::%s:role/nodes%s", *accountID.Account, infrav1.DefaultNameSuffix),
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
					"AWS": fmt.Sprintf("arn:aws:iam::%s:role/control-plane%s", *accountID.Account, infrav1.DefaultNameSuffix),
				},
				"Action": []string{
					"s3:GetObject",
				},
				"Resource": fmt.Sprintf("arn:aws:s3:::%s/control-plane/*", s.bucketName()),
			},
		},
	}

	policy, err := json.Marshal(readOnlyAnonUserPolicy)
	if err != nil {
		return "", errors.Wrap(err, "building bucket policy")
	}

	return string(policy), nil
}

func (s *Service) bucketManagementEnabled() bool {
	return s.scope.Bucket().Enabled
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

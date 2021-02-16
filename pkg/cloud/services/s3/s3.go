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
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/pkg/errors"

	"sigs.k8s.io/cluster-api-provider-aws/pkg/cloud/scope"
)

// Service holds a collection of interfaces.
// The interfaces are broken down like this to group functions together.
// One alternative is to have a large list of functions from the ec2 client.
type Service struct {
	scope    scope.S3Scope
	s3Client s3iface.S3API
}

// NewService returns a new service given the api clients.
func NewService(s3Scope scope.S3Scope) *Service {
	s3Client := scope.NewS3Client(s3Scope, s3Scope, s3Scope.InfraCluster())

	return &Service{
		scope:    s3Scope,
		s3Client: s3Client,
	}
}

func (s *Service) bucketManagementEnabled() bool {
	return s.scope.Bucket().Enabled
}

func (s *Service) DeleteBucket() error {
	if !s.bucketManagementEnabled() {
		return nil
	}

	s.scope.Info("Would delete S3 Bucket")

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

	s.scope.Info("Would reconcile S3 Bucket, maybe use %q as a name?", s.bucketName())

	if _, err := s.s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(s.bucketName()),
	}); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			// If bucket already exists, all good.
			// TODO: This will fail if bucket is shared with other cluster.
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				return nil
			default:
				return errors.Wrap(aerr, "creating S3 bucket")
			}
		}

		return errors.Wrap(err, "creating S3 bucket")
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
	return m.Name()
}

func (s *Service) Delete(m *scope.MachineScope) error {
	s.scope.Info("Would delete S3 object for machine %q", m.Name())

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
		ACL:    aws.String("public-read"), // TODO: We can do better, this is insecure.
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

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
	"bytes"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	ignTypes "github.com/coreos/ignition/config/v2_2/types"
	"github.com/google/uuid"
	"net/url"
	"strings"
)

var (
	baseIgnitionUri = map[string]string{
		"v1.15.11": "ignition-config/k8s-v1.15.11.ign",
		"v1.16.8":  "ignition-config/k8s-v1.16.8.ign",
		"v1.17.4":  "ignition-config/k8s-v1.17.4.ign",
	}
	userDataDirName = "node-userdata"
)

const (
	KubernetesDefaultVersion = "v1.17.4"
	IgnitionSchemaVersion    = "2.2.0"
)

func NewS3TemplateBackend(userDataDir string, userDataBucket string) (*S3TemplateBackend, error) {
	session, err := session.NewSession()
	if err != nil {
		ignitionLogger.Error(err, "failed to initialize s3 session")
		return nil, err
	}
	return &S3TemplateBackend{
		UserDataDir:    userDataDir,
		UserDataBucket: userDataBucket,
		session:        session,
	}, nil
}

type S3TemplateBackend struct {
	UserDataDir    string
	UserDataBucket string
	FilePath       string
	session        *session.Session
	S3Client       s3iface.S3API
}

func (factory *S3TemplateBackend) getUserDataDir() string {
	return factory.UserDataDir
}

func (factory *S3TemplateBackend) getUserDataBucket() string {
	return factory.UserDataBucket
}

func (factory *S3TemplateBackend) getFilePath() string {
	return factory.FilePath
}

func (factory *S3TemplateBackend) getIgnitionConfigTemplate(node *Node) (*ignTypes.Config, error) {
	templateConfigUri, ok := baseIgnitionUri[node.Version]
	if !ok {
		err := errors.New("kubernetes version is not supported.")
		ignitionLogger.Error(err, "kubernetes version is not supported.")
		templateConfigUri = baseIgnitionUri[KubernetesDefaultVersion]
	}
	baseIgnitionUrl := &url.URL{
		Scheme: "s3",
		Host:   factory.UserDataBucket,
		Path:   templateConfigUri,
	}
	out := factory.getIgnitionBaseConfig()
	out.Ignition.Config = ignTypes.IgnitionConfig{
		Append: []ignTypes.ConfigReference{
			{
				Source: baseIgnitionUrl.String(),
			},
		},
	}
	return out, nil
}

func (factory *S3TemplateBackend) applyConfig(userdata []byte) (*ignTypes.Config, error) {
	factory.FilePath = strings.Join([]string{userDataDirName, uuid.New().String()}, "/")

	var errPut error
	_, errPut = factory.S3Client.PutObject(&s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(bytes.NewReader(userdata)),
		Bucket: aws.String(factory.UserDataBucket),
		Key:    aws.String(factory.FilePath),
	})
	if errPut != nil {
		return nil, fmt.Errorf("failed to put object to bucket")
	}

	userDataUrl := url.URL{
		Scheme: "s3",
		Host:   factory.UserDataBucket,
		Path:   factory.FilePath,
	}
	out := factory.getIgnitionBaseConfig()
	out.Ignition.Config = ignTypes.IgnitionConfig{
		Replace: &ignTypes.ConfigReference{
			Source: userDataUrl.String(),
		},
	}
	return out, nil
}

func (factory *S3TemplateBackend) getIgnitionBaseConfig() *ignTypes.Config {
	return &ignTypes.Config{
		Ignition: ignTypes.Ignition{
			Version: IgnitionSchemaVersion,
		},
	}
}

module sigs.k8s.io/cluster-api-provider-aws

go 1.15

require (
	github.com/ajeddeloh/go-json v0.0.0-20200220154158-5ae607161559 // indirect
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/aws/amazon-vpc-cni-k8s v1.7.5
	github.com/aws/aws-sdk-go v1.35.30
	github.com/awslabs/goformation/v4 v4.15.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/coreos/ignition v0.35.0
	github.com/go-logr/logr v0.3.0
	github.com/golang/mock v1.4.4
	github.com/google/goexpect v0.0.0-20200816234442-b5b77125c2c5
	github.com/google/goterm v0.0.0-20200907032337-555d40f16ae2 // indirect
	github.com/google/uuid v1.1.2
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/sergi/go-diff v1.1.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1 // indirect
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/vincent-petithory/dataurl v0.0.0-20191104211930-d1553a71de50
	go4.org v0.0.0-20200411211856-f5505b9728dd // indirect
	golang.org/x/crypto v0.0.0-20200930160638-afb6bcd081ae
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6 // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.2
	k8s.io/apiextensions-apiserver v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/component-base v0.19.2
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	sigs.k8s.io/aws-iam-authenticator v0.5.1
	sigs.k8s.io/cluster-api v0.3.11-0.20210115191551-61dc332270dc
	sigs.k8s.io/controller-runtime v0.7.1-0.20201215171748-096b2e07c091
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/cluster-api => github.com/kinvolk/cluster-api v0.0.0-20201216133705-042e36b28a98

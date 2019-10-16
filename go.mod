module github.com/solo-io/valet

go 1.12

require (
	cloud.google.com/go v0.43.0
	github.com/aws/aws-sdk-go v1.21.10
	github.com/ghodss/yaml v1.0.0
	github.com/golang/mock v1.3.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/solo-io/go-utils v0.9.17
	github.com/spf13/cobra v0.0.3
	go.uber.org/zap v1.9.1
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/tools v0.0.0-20190628153133-6cdbf07be9d0
	google.golang.org/api v0.7.0
	google.golang.org/genproto v0.0.0-20190716160619-c506a9f90610
	google.golang.org/grpc v1.21.1
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20181221193117-173ce66c1e39+incompatible
	k8s.io/apimachinery v0.0.0-20190104073114-849b284f3b75+incompatible
)

replace (
	github.com/Sirupsen/logrus v1.0.5 => github.com/sirupsen/logrus v1.0.5
	github.com/Sirupsen/logrus v1.3.0 => github.com/Sirupsen/logrus v1.0.6
	github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.0.6
)

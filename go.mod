module github.com/solo-io/valet

go 1.13

require (
	cloud.google.com/go v0.40.0
	github.com/Masterminds/sprig/v3 v3.0.0
	github.com/avast/retry-go v2.2.0+incompatible
	github.com/aws/aws-sdk-go v1.26.5
	github.com/ghodss/yaml v1.0.1-0.20190212202910-dc05a4bc0ab4
	github.com/golang/mock v1.3.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/helm/helm v2.16.1+incompatible
	github.com/mitchellh/hashstructure v1.0.0
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	github.com/rotisserie/eris v0.1.1
	github.com/solo-io/go-utils v0.14.0
	github.com/spf13/cobra v0.0.5
	go.uber.org/zap v1.9.1
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/api v0.6.0
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6
	google.golang.org/grpc v1.24.0
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20191121015604-11707872ac1c
	k8s.io/apimachinery v0.0.0-20191121015412-41065c7a8c2a
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	//kube 1.16
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
)

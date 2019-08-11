module github.com/solo-io/kube-cluster

go 1.12

require (
	cloud.google.com/go v0.43.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/solo-io/go-utils v0.9.17
	github.com/spf13/cobra v0.0.3
	go.uber.org/zap v1.9.1
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/tools v0.0.0-20190628153133-6cdbf07be9d0
	google.golang.org/api v0.7.0
	google.golang.org/genproto v0.0.0-20190716160619-c506a9f90610
)

replace (
	github.com/Sirupsen/logrus v1.0.5 => github.com/sirupsen/logrus v1.0.5
	github.com/Sirupsen/logrus v1.3.0 => github.com/Sirupsen/logrus v1.0.6
	github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.0.6
)

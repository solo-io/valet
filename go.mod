module github.com/solo-io/kube-cluster

go 1.12

require (
	cloud.google.com/go v0.43.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/solo-io/go-utils v0.9.17
	go.uber.org/zap v1.9.1
	google.golang.org/api v0.7.0
)

replace (
	github.com/Sirupsen/logrus v1.0.5 => github.com/sirupsen/logrus v1.0.5
	github.com/Sirupsen/logrus v1.3.0 => github.com/Sirupsen/logrus v1.0.6
	github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.0.6
)

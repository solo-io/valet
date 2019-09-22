package cluster

import "context"

type KubeCluster interface {
	KubeVersion(ctx context.Context) string
	IsRunning(ctx context.Context) (bool, error)
	SetKubeContext(ctx context.Context) error
	Create(ctx context.Context) error
	Destroy(ctx context.Context) error
}

type Provisioner interface {
	// Checks to see if a cluster with the spec exists. If so, validates that it matches the
	// spec and returns it. Otherwise, it creates the cluster.
	Ensure(ctx context.Context) error

	GetCluster() KubeCluster
}


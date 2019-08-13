package gke

import (
	container "cloud.google.com/go/container/apiv1"
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/ensure/cluster/cluster"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	container2 "google.golang.org/genproto/googleapis/container/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var _ cluster.KubeCluster = new(gkeCluster)

func getParent(config options.Cluster) string {
	return fmt.Sprintf("projects/%s/locations/%s", config.GKE.Project, config.GKE.Location)
}

func getClusterIdentifier(config options.Cluster) string {
	return fmt.Sprintf("%s/clusters/%s", getParent(config), config.GKE.Name)
}

func getOperationIdentifier(config options.Cluster, opName string) string {
	return fmt.Sprintf("%s/operations/%s", getParent(config), opName)
}

func NewGkeClusterFromOpts(ctx context.Context, opts options.Cluster) (*gkeCluster, error) {
	client, err := getClusterManagerClient(ctx)
	if err != nil {
		return nil, err
	}
	return &gkeCluster{
		config: opts,
		client: client,
	}, nil
}

func getClusterManagerClient(ctx context.Context) (*container.ClusterManagerClient, error) {
	ts, err := google.DefaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error locating token", zap.Error(err))
		return nil, err
	}

	client, err := container.NewClusterManagerClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating cluster client", zap.Error(err))
		return nil, err
	}
	return client, nil
}

type gkeCluster struct {
	config options.Cluster
	client *container.ClusterManagerClient
}

func (c *gkeCluster) KubeVersion(ctx context.Context) string {
	return c.config.KubeVersion
}

func (c *gkeCluster) IsRunning(ctx context.Context) (bool, error) {
	cluster, err := c.getCluster(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error checking cluster status",
			zap.Error(err),
			zap.String("clusterName", getClusterIdentifier(c.config)))
		return false, err
	} else if cluster == nil {
		contextutils.LoggerFrom(ctx).Infow("Cluster not running")
		return false, nil
	}
	return cluster.GetStatus() == container2.Cluster_RUNNING, nil
}

func (c *gkeCluster) getCluster(ctx context.Context) (*container2.Cluster, error) {
	cluster, err := c.client.GetCluster(ctx, &container2.GetClusterRequest{Name: getClusterIdentifier(c.config)})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound{
			contextutils.LoggerFrom(ctx).Infow("Cluster not found")
			return nil, nil
		}
		// Use st.Message() and st.Code()
		contextutils.LoggerFrom(ctx).Errorw("Error getting cluster",
			zap.Error(err),
			zap.String("clusterName", getClusterIdentifier(c.config)))
		return nil, err
	}
	return cluster, nil
}

func (c *gkeCluster) SetKubeContext(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Setting kube context to GKE")
	out, err := internal.ExecuteCmd("gcloud",
		"container", "clusters", "get-credentials",
		"--project="+c.config.GKE.Project,
		"--zone="+c.config.GKE.Location,
		c.config.GKE.Name)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error setting kube context",
			zap.Error(err),
			zap.String("out", out))
	}
	return err
}

func (c *gkeCluster) Create(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Creating cluster", zap.String("name", c.config.GKE.Name))
	clusterToCreate := container2.Cluster{
		Name:             c.config.GKE.Name,
		InitialNodeCount: 3,
		ResourceLabels: map[string]string {
			"creator": "valet",
		},
	}
	req := container2.CreateClusterRequest{
		Parent:  getParent(c.config),
		Cluster: &clusterToCreate,
	}
	operation, err := c.client.CreateCluster(ctx, &req)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating cluster", zap.Error(err))
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("Create cluster led to operation", zap.Any("operation", operation))
	err = c.waitForOperation(ctx, getOperationIdentifier(c.config, operation.Name))
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error waiting for cluster creation", zap.Error(err))
	}
	return err
}

func (c *gkeCluster) waitForOperation(ctx context.Context, operationId string) error {
	ticker := time.NewTicker(5 * time.Second)
	getOp := container2.GetOperationRequest{
		Name: operationId,
	}
	for range ticker.C {
		operation, err := c.client.GetOperation(ctx, &getOp)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("Error monitoring operation", zap.Error(err))
			return err
		}
		contextutils.LoggerFrom(ctx).Infow("Status",
			zap.String("operation", operation.Status.String()))
		if operation.Status == container2.Operation_DONE {
			contextutils.LoggerFrom(ctx).Infow("Operation done")
			return nil
		}
	}
	panic("Operation status unknown")
}

func (c *gkeCluster) Destroy(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Deleting cluster", zap.String("name", c.config.GKE.Name))
	req := container2.DeleteClusterRequest{
		Name: getClusterIdentifier(c.config),
	}
	operation, err := c.client.DeleteCluster(ctx, &req)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error deleting cluster", zap.Error(err))
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("Delete cluster led to operation", zap.Any("operation", operation))
	err = c.waitForOperation(ctx, getOperationIdentifier(c.config, operation.Name))
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error waiting for cluster to be deleted", zap.Error(err))
	}
	return err
}

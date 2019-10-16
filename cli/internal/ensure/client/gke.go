package client

import (
	container "cloud.google.com/go/container/apiv1"
	"fmt"
	container2 "google.golang.org/genproto/googleapis/container/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"context"
	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	"time"
)

type GkeClient interface {
	Create(ctx context.Context, name, project, zone string) error
	Destroy(ctx context.Context, name, project, zone string) error
	IsRunning(ctx context.Context, name, project, zone string) (bool, error)
}

var _ GkeClient = new(gkeClient)

type gkeClient struct {
	client *container.ClusterManagerClient
}

func NewGkeClient(ctx context.Context) (*gkeClient, error) {
	client, err := getClusterManagerClient(ctx)
	if err != nil {
		return nil, err
	}
	return &gkeClient{
		client: client,
	}, nil
}

func getClusterManagerClient(ctx context.Context) (*container.ClusterManagerClient, error) {
	ts, err := google.DefaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error locating token", zap.Error(err))
		return nil, err
	}

	client, err := container.NewClusterManagerClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error creating cluster client", zap.Error(err))
		return nil, err
	}
	return client, nil
}

func (c *gkeClient) IsRunning(ctx context.Context, name, project, zone string) (bool, error) {
	cluster, err := c.getCluster(ctx, name, project, zone)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error checking cluster status",
			zap.Error(err),
			zap.String("clusterName", getClusterIdentifier(name, project, zone)))
		return false, err
	} else if cluster == nil {
		return false, nil
	}
	return cluster.GetStatus() == container2.Cluster_RUNNING, nil
}

func (c *gkeClient) getCluster(ctx context.Context, name, project, zone string) (*container2.Cluster, error) {
	cluster, err := c.client.GetCluster(ctx, &container2.GetClusterRequest{Name: getClusterIdentifier(name, project, zone)})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			contextutils.LoggerFrom(ctx).Infow("cluster not found")
			return nil, nil
		}
		// Use st.Message() and st.Code()
		contextutils.LoggerFrom(ctx).Errorw("error getting cluster",
			zap.Error(err),
			zap.String("clusterName", getClusterIdentifier(name, project, zone)))
		return nil, err
	}
	return cluster, nil
}


func (c *gkeClient) Create(ctx context.Context,name, project, zone string) error {
	contextutils.LoggerFrom(ctx).Infow("creating cluster", zap.String("name", name))
	nodePool := container2.NodePool{
		Name:             "pool-1",
		InitialNodeCount: 1,
		Autoscaling: &container2.NodePoolAutoscaling{
			Enabled:      true,
			MinNodeCount: 1,
			MaxNodeCount: 30,
		},
		Config: &container2.NodeConfig{
			MachineType: "n1-standard-4",
		},
		Management: &container2.NodeManagement{
			AutoUpgrade: false,
		},
	}
	clusterToCreate := container2.Cluster{
		Name:      name,
		NodePools: []*container2.NodePool{&nodePool},
		ResourceLabels: map[string]string{
			"creator": "valet",
		},
	}
	req := container2.CreateClusterRequest{
		Parent:  getParent(project, zone),
		Cluster: &clusterToCreate,
	}
	operation, err := c.client.CreateCluster(ctx, &req)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error creating cluster", zap.Error(err))
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("create cluster led to operation", zap.Any("operation", operation))
	err = c.waitForOperation(ctx, getOperationIdentifier(project, zone, operation.Name))
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error waiting for cluster creation", zap.Error(err))
	}
	return err
}

func (c *gkeClient) Destroy(ctx context.Context, name, project, zone string) error {
	contextutils.LoggerFrom(ctx).Infow("deleting cluster", zap.String("name", name))
	req := container2.DeleteClusterRequest{
		Name: getClusterIdentifier(name, project, zone),
	}
	operation, err := c.client.DeleteCluster(ctx, &req)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error deleting cluster", zap.Error(err))
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("delete cluster led to operation", zap.Any("operation", operation))
	err = c.waitForOperation(ctx, getOperationIdentifier(project, zone, operation.Name))
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error waiting for cluster to be deleted", zap.Error(err))
	}
	return err
}

func (c *gkeClient) waitForOperation(ctx context.Context, operationId string) error {
	ticker := time.NewTicker(5 * time.Second)
	getOp := container2.GetOperationRequest{
		Name: operationId,
	}
	for range ticker.C {
		operation, err := c.client.GetOperation(ctx, &getOp)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error monitoring operation", zap.Error(err))
			return err
		}
		contextutils.LoggerFrom(ctx).Infow("Status",
			zap.String("operation", operation.Status.String()))
		if operation.Status == container2.Operation_DONE {
			contextutils.LoggerFrom(ctx).Infow("operation done")
			return nil
		}
	}
	panic("Operation status unknown")
}

func getParent(project, zone string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, zone)
}

func getClusterIdentifier(name, project, zone string) string {
	return fmt.Sprintf("%s/clusters/%s", getParent(project, zone), name)
}

func getOperationIdentifier(project, zone, opName string) string {
	return fmt.Sprintf("%s/operations/%s", getParent(project, zone), opName)
}

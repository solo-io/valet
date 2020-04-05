package gke

import (
	container "cloud.google.com/go/container/apiv1"
	"fmt"
	container2 "google.golang.org/genproto/googleapis/container/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"context"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

type CreateOptions struct {
	InitialNodeCount int    `json:"initialNodeCount" valet:"template,key=InitialNodeCount,default=1"`
	KubeVersion      string `json:"version" valet:"template,key=KubeVersion,default=1.13.0"`
}

type Client interface {
	Create(ctx context.Context, name, project, zone string, opts *CreateOptions) error
	Destroy(ctx context.Context, name, project, zone string) error
	IsRunning(ctx context.Context, name, project, zone string) (bool, error)
}

var _ Client = new(client)

type client struct {
	clusterClient *container.ClusterManagerClient
}

func NewClient(ctx context.Context) (*client, error) {
	clusterClient, err := getClusterManagerClient(ctx)
	if err != nil {
		return nil, err
	}
	return &client{
		clusterClient: clusterClient,
	}, nil
}

func getClusterManagerClient(ctx context.Context) (*container.ClusterManagerClient, error) {
	ts, err := google.DefaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		return nil, err
	}

	clusterClient, err := container.NewClusterManagerClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}
	return clusterClient, nil
}

func (c *client) IsRunning(ctx context.Context, name, project, zone string) (bool, error) {
	cluster, err := c.getCluster(ctx, name, project, zone)
	if err != nil {
		return false, err
	} else if cluster == nil {
		return false, nil
	}
	return cluster.GetStatus() == container2.Cluster_RUNNING, nil
}

func (c *client) getCluster(ctx context.Context, name, project, zone string) (*container2.Cluster, error) {
	cluster, err := c.clusterClient.GetCluster(ctx, &container2.GetClusterRequest{Name: getClusterIdentifier(name, project, zone)})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	return cluster, nil
}

func (c *client) Create(ctx context.Context, name, project, zone string, opts *CreateOptions) error {
	nodePool := container2.NodePool{
		Name:             "pool-1",
		InitialNodeCount: int32(opts.InitialNodeCount),
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
		Name:                  name,
		NodePools:             []*container2.NodePool{&nodePool},
		InitialClusterVersion: opts.KubeVersion,
		ResourceLabels: map[string]string{
			"creator": "valet",
		},
	}
	req := container2.CreateClusterRequest{
		Parent:  getParent(project, zone),
		Cluster: &clusterToCreate,
	}
	operation, err := c.clusterClient.CreateCluster(ctx, &req)
	if err != nil {
		return err
	}
	return c.waitForOperation(ctx, getOperationIdentifier(project, zone, operation.Name))
}

func (c *client) Destroy(ctx context.Context, name, project, zone string) error {
	req := container2.DeleteClusterRequest{
		Name: getClusterIdentifier(name, project, zone),
	}
	operation, err := c.clusterClient.DeleteCluster(ctx, &req)
	if err != nil {
		return err
	}
	return c.waitForOperation(ctx, getOperationIdentifier(project, zone, operation.Name))
}

func (c *client) waitForOperation(ctx context.Context, operationId string) error {
	ticker := time.NewTicker(5 * time.Second)
	getOp := container2.GetOperationRequest{
		Name: operationId,
	}
	for range ticker.C {
		operation, err := c.clusterClient.GetOperation(ctx, &getOp)
		if err != nil {
			return err
		}
		if operation.Status == container2.Operation_DONE {
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

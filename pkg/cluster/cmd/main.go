package main

import (
	container "cloud.google.com/go/container/apiv1"
	"context"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/kube-cluster/pkg/internal"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	container2 "google.golang.org/genproto/googleapis/container/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func main() {
	ctx := internal.GetInitialContext()
	var config EnvConfig
	err := envconfig.Process("", &config)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Error parsing env config", zap.Error(err))
	}
	client := mustGetClusterManagerClient(ctx)
	cluster := getOrCreateCluster(ctx, client, config)
	contextutils.LoggerFrom(ctx).Infow("Cluster is running", zap.String("name", cluster.Name))
}

type EnvConfig struct {
	ClusterName string `split_words:"true",required:"true"`
	ClusterLocation string `split_words:"true",required:"true"`
	ClusterProject string `split_words:"true",required:"true"`
}

func getInitialContext() context.Context {
	return context.TODO()
}

func getParent(config EnvConfig) string {
	return fmt.Sprintf("projects/%s/locations/%s", config.ClusterProject, config.ClusterLocation)
}

func getClusterIdentifier(config EnvConfig) string {
	return fmt.Sprintf("%s/clusters/%s", getParent(config), config.ClusterName)
}

func getOperationIdentifier(config EnvConfig, opName string) string {
	return fmt.Sprintf("%s/operations/%s", getParent(config), opName)
}

func getOrCreateCluster(ctx context.Context, client *container.ClusterManagerClient, config EnvConfig) *container2.Cluster {
	contextutils.LoggerFrom(ctx).Infow("Getting cluster", zap.String("clusterIdentifier", getClusterIdentifier(config)))
	req := container2.GetClusterRequest{
		Name: getClusterIdentifier(config),
	}
	cluster, err := client.GetCluster(ctx, &req)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.NotFound {
			contextutils.LoggerFrom(ctx).Fatalw("Error getting cluster", zap.Error(err))
		}
		// create cluster
		contextutils.LoggerFrom(ctx).Infow("Could not find cluster", zap.String("name", config.ClusterName))
		return createCluster(ctx, client, config)
	}
	contextutils.LoggerFrom(ctx).Infow("Found cluster",
		zap.String("name", cluster.Name),
		zap.String("status", cluster.Status.String()))
	return cluster
}

func createCluster(ctx context.Context, client *container.ClusterManagerClient, config EnvConfig) *container2.Cluster {
	contextutils.LoggerFrom(ctx).Infow("Creating cluster", zap.String("name", config.ClusterName))
	cluster := container2.Cluster{
		Name: config.ClusterName,
		InitialNodeCount: 3,
	}
	req := container2.CreateClusterRequest{
		Parent: getParent(config),
		Cluster: &cluster,
	}
	operation, err := client.CreateCluster(ctx, &req)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Error creating cluster", zap.Error(err))
	}
	contextutils.LoggerFrom(ctx).Infow("Create cluster led to operation", zap.Any("operation", operation))
	ticker := time.NewTicker(5 * time.Second)
	getOp := container2.GetOperationRequest{
		Name: getOperationIdentifier(config, operation.Name),
	}
	for range ticker.C {
		operation, err = client.GetOperation(ctx, &getOp)
		if err != nil {
			contextutils.LoggerFrom(ctx).Fatalw("Error monitoring operation", zap.Error(err))
		}
		readCluster, err := client.GetCluster(ctx, &container2.GetClusterRequest{Name: getClusterIdentifier(config)})
		if err != nil {
			contextutils.LoggerFrom(ctx).Fatalw("Error monitoring cluster",
				zap.Error(err),
				zap.String("clusterName", getClusterIdentifier(config)))
		}
		contextutils.LoggerFrom(ctx).Infow("Status",
			zap.String("operation", operation.Status.String()),
			zap.String("cluster", readCluster.Status.String()))
		if operation.Status == container2.Operation_DONE {
			contextutils.LoggerFrom(ctx).Infow("Operation done")
			contextutils.LoggerFrom(ctx).Infow("", zap.Any("cluster", readCluster))
			return readCluster
		}
	}
	return nil
}

func mustGetClusterManagerClient(ctx context.Context) *container.ClusterManagerClient {
	ts, err := google.DefaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Error locating token", zap.Error(err))
	}

	client, err := container.NewClusterManagerClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Error creating cluster client", zap.Error(err))
	}
	return client
}

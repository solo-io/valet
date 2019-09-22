package resources

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/internal"
	"go.uber.org/zap"
)

func EnsureResources(ctx context.Context, resources []string) error {
	for _, resource := range resources {
		out, err := internal.ExecuteCmd("kubectl", "apply", "-f", resource)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("Error applying resources to cluster",
				zap.Error(err),
				zap.String("out", out),
				zap.String("resource", resource))
			return err
		}
	}
	return nil
}

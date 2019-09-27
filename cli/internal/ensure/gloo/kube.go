package gloo

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/githubutils"
	"go.uber.org/zap"
)

func getLatestTag(ctx context.Context, repo string) (string, error) {
	client := githubutils.GetClientWithOrWithoutToken(ctx)
	release, _, err := client.Repositories.GetLatestRelease(ctx, "solo-io", repo)
	if err != nil {
		wrapped := CouldNotDetermineVersionError(err)
		contextutils.LoggerFrom(ctx).Errorw(err.Error(), zap.Error(err))
		return "", wrapped
	}
	return release.GetTagName(), nil
}

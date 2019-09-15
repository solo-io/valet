package gloo

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	UnableToDownloadEnterpriseGlooctlError =
		errors.Errorf("Unable to download glooctl for enterprise gloo, set GITHUB_TOKEN in your environment")
)

type GlooctlEnsurer interface {
	Install(ctx context.Context, gloo options.Gloo) (string, error)
}

var _ GlooctlEnsurer = new(glooctlEnsurer)

func NewGlooctlEnsurer() *glooctlEnsurer {
	return &glooctlEnsurer{}
}

type glooctlEnsurer struct {
}

func (g *glooctlEnsurer) Install(ctx context.Context, gloo options.Gloo) (string, error) {
	return ensureGlooctl(ctx, gloo)
}

func ensureGlooctl(ctx context.Context, gloo options.Gloo) (string, error) {
	path, err := ensureGlooctlIsDownloaded(ctx, gloo)
	if err != nil {
		return "", err
	}
	defaultPath, err := getDefaultFilepath()
	if err != nil {
		return "", err
	}
	out, err := internal.ExecuteCmd("cp", path, defaultPath)
	if err != nil {
		contextutils.LoggerFrom(ctx).Warnw("could not replace default glooctl on path, falling back to full path", zap.Error(err), zap.String("out", out))
		return path, nil
	}
	out, err = internal.ExecuteCmd("command", "-v", "glooctl")
	if err != nil {
		contextutils.LoggerFrom(ctx).Warnw("error trying to find glooctl on path, falling back to full path",
			zap.Error(err), zap.String("out", out))
		return defaultPath, nil
	}
	glooctlOnPath := strings.TrimSpace(out)
	if glooctlOnPath != defaultPath {
		contextutils.LoggerFrom(ctx).Warnw("there is another glooctl on the path, falling back to full path",
			zap.String("glooctlOnPath", glooctlOnPath),
			zap.String("fullPath", defaultPath))
		return defaultPath, nil
	}
	contextutils.LoggerFrom(ctx).Infow("updated glooctl on path to be this version")
	// "glooctl" doesn't work here, so just use full default path
	return defaultPath, nil
}

func ensureGlooctlIsDownloaded(ctx context.Context, gloo options.Gloo) (string, error) {
	localPathToGlooctl, err := getFilepath(gloo)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(localPathToGlooctl); err == nil {
		return localPathToGlooctl, nil
	} else if !os.IsNotExist(err) {
		contextutils.LoggerFrom(ctx).Errorw("Error checking if glooctl was downloaded, attempting to download", zap.Error(err))
	}

	if gloo.Enterprise && os.Getenv("GITHUB_TOKEN") == "" {
		contextutils.LoggerFrom(ctx).Errorw(UnableToDownloadEnterpriseGlooctlError.Error())
		return "", UnableToDownloadEnterpriseGlooctlError
	}

	client := githubutils.GetClientWithOrWithoutToken(ctx)
	downloader := NewGithubArtifactDownloader(client, getRepo(gloo.Enterprise), gloo.Version)
	err = downloader.Download(ctx, getAssetName(), localPathToGlooctl)
	if err != nil {
		return "", err
	}
	return localPathToGlooctl, nil
}

func getRepo(enterprise bool) string {
	if enterprise {
		return "solo-projects"
	}
	return "gloo"
}

func getAssetName() string {
	return "glooctl-" + runtime.GOOS + "-amd64"
}

func getDefaultFilepath() (string, error) {
	dir, err := getBinaryDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "glooctl"), nil
}


func getFilepath(gloo options.Gloo) (string, error) {
	dir, err := getBinaryDir()
	if err != nil {
		return "", nil
	}
	enterpriseText := ""
	if gloo.Enterprise {
		enterpriseText = "-enterprise"
	}
	filename := fmt.Sprintf("glooctl%s-%s", enterpriseText, gloo.Version[1:])
	return filepath.Join(dir, filename), nil
}

func getBinaryDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	glooBinDir := filepath.Join(homeDir, ".gloo", "bin")
	err = os.MkdirAll(glooBinDir, os.ModePerm)
	if err != nil {
		return "", err
	}
	return glooBinDir, nil
}





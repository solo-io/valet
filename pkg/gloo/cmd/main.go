package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/kelseyhightower/envconfig"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/randutils"
	"github.com/solo-io/kube-cluster/pkg/internal"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	ctx := internal.GetInitialContext()
	config := mustGetConfig(ctx)
	mustInstallGloo(ctx, config)
}

type EnvConfig struct {
	GlooVersion string `split_words:"true",required:"true"`
	Enterprise bool `default:"false"`
	LicenseKey string `split_words:"true"`
}

func mustGetConfig(ctx context.Context) EnvConfig {
	var config EnvConfig
	err := envconfig.Process("", &config)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Error parsing env config", zap.Error(err))
	}
	if config.GlooVersion == "" {
		contextutils.LoggerFrom(ctx).Fatalw("Must set GLOO_VERSION in environment")
	}
	return config
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

func mustInstallGloo(ctx context.Context, config EnvConfig) {
	client := githubutils.GetClientWithOrWithoutToken(ctx)
	release := mustGetRelease(ctx, config, client)
	asset := mustGetAsset(ctx, config, release)
	downloadUrl := asset.GetBrowserDownloadURL()
	filepath := fmt.Sprintf("glooctl-%s-%s", config.GlooVersion, randutils.RandString(4))
	mustDownloadFile(ctx, filepath, downloadUrl)
	mustChmod(ctx, filepath)
	mustInstall(ctx, filepath, config)
}

func GlooctlOut(filepath string, args ...string) (string, error) {
	cmd := exec.Command(filepath, args...)
	cmd.Env = os.Environ()
	for i, pair := range cmd.Env {
		if strings.HasPrefix(pair, "DEBUG") {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i+1:]...)
			break
		}
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s (%v)", out, err)
	}
	return string(out), err
}

func mustInstall(ctx context.Context, filepath string, config EnvConfig) {
	args := []string {"install", "gateway"}
	if config.Enterprise {
		args = append(args, "--license-key", config.LicenseKey)
	}
	out, err := GlooctlOut("./" + filepath, args...)
	contextutils.LoggerFrom(ctx).Infow("Ran glooctl", zap.String("output", out))
	if err != nil {
		contextutils.LoggerFrom(ctx).Infow("Failed to install gloo",
			zap.Error(err))

	}
}

func mustGetRelease(ctx context.Context, config EnvConfig, client *github.Client) *github.RepositoryRelease {
	release, _, err := client.Repositories.GetReleaseByTag(ctx, "solo-io", getRepo(config.Enterprise), config.GlooVersion)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Could not download glooctl",
			zap.Error(err),
			zap.Bool("enterprise", config.Enterprise),
			zap.String("tag", config.GlooVersion))
	}
	return release
}

func mustGetAsset(ctx context.Context, config EnvConfig, release *github.RepositoryRelease) github.ReleaseAsset {
	desiredAsset := getAssetName()
	for _, asset := range release.Assets {
		if asset.GetName() == desiredAsset {
			return asset
		}
	}
	contextutils.LoggerFrom(ctx).Fatalw("Could not find asset",
		zap.Bool("enterprise", config.Enterprise),
		zap.String("tag", config.GlooVersion),
		zap.String("assetName", desiredAsset))
	return github.ReleaseAsset{}
}

func mustChmod(ctx context.Context, filepath string) {
	err := os.Chmod(filepath, os.ModePerm)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Could not make glooctl executable",
			zap.Error(err),
			zap.String("filepath", filepath))
	}
}

func mustDownloadFile(ctx context.Context, filepath string, url string) {
	err := downloadFile(filepath, url)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Could not find asset",
			zap.Error(err),
			zap.String("filepath", filepath),
			zap.String("url", url))
	}
}

func downloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
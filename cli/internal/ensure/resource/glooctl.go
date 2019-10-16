package resource

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/valet/cli/internal/ensure/client"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"runtime"
)

const (
	Owner = "solo-io"
)

var (
	CouldNotDetermineVersionError = func(err error) error {
		return errors.Wrapf(err, "Error determining latest release.")
	}
	GlooctlNotEnsuredError = errors.Errorf("glooctl not ensured")
)

type Glooctl struct {
	Version           string `yaml:"version"`
	LocalArtifactsDir string `yaml:"localArtifactsDir"`
	ValetArtifacts bool   `yaml:"valetArtifacts"`
	Enterprise        bool   `yaml:"enterprise"`
}

func (g *Glooctl) Ensure(ctx context.Context, command cmd.Factory) error {
	if g.LocalArtifactsDir != "" {
		// try to use local artifact if it exists
		glooctlPath := filepath.Join(g.LocalArtifactsDir, getAssetName())
		if _, err := os.Stat(glooctlPath); err == nil {
			command.SetLocalPath(cmd.GlooctlCmd, glooctlPath)
			return nil
		} else if !os.IsNotExist(err) {
			contextutils.LoggerFrom(ctx).Errorw("error checking if glooctl was in artifacts directory", zap.Error(err))
			return err
		}
	}

	var glooctlVersion string
	if g.Enterprise || g.Version == "" {
		// use latest OS artifact if no version set and / or installing enterprise
		tag, err := getLatestTag(ctx, GlooRepo)
		if err != nil {
			return err
		}
		glooctlVersion = tag[1:]
	} else {
		glooctlVersion = g.Version
	}
	localPath, err := prepareLocalPath(glooctlVersion)
	if err != nil {
		return err
	}
	command.SetLocalPath(cmd.GlooctlCmd, localPath)
	if _, err := os.Stat(localPath); err == nil {
		contextutils.LoggerFrom(ctx).Infow("glooctl already exists", zap.String("localPath", localPath))
		return nil
	}
	if g.ValetArtifacts {
		return g.ensureFromValet(ctx, glooctlVersion, localPath)
	}
	return g.ensureFromGithub(ctx, glooctlVersion, localPath)
}

func (g *Glooctl) ensureFromGithub(ctx context.Context, version, localPath string) error {
	githubClient := githubutils.GetClientWithOrWithoutToken(ctx)
	downloader := client.NewGithubArtifactDownloader(githubClient, GlooRepo, "v"+version)
	return downloader.Download(ctx, getAssetName(), localPath)
}

func (g *Glooctl) ensureFromValet(ctx context.Context, version, localPath string) error {
	downloader := client.NewUrlArtifactDownloader()
	remotePath := fmt.Sprintf("https://storage.googleapis.com/valet/artifacts/gloo/%s/glooctl-%s-amd64", version, runtime.GOOS)
	return downloader.Download(ctx, remotePath, localPath)
}

func getAssetName() string {
	return fmt.Sprintf("glooctl-%s-amd64", runtime.GOOS)
}

func prepareLocalPath(version string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".gloo", "bin")
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("glooctl-%s", version)
	return filepath.Join(dir, filename), nil
}

func (g *Glooctl) Teardown(ctx context.Context) error {
	return nil
}

func getLatestTag(ctx context.Context, repo string) (string, error) {
	githubClient := githubutils.GetClientWithOrWithoutToken(ctx)
	release, _, err := githubClient.Repositories.GetLatestRelease(ctx, Owner, repo)
	if err != nil {
		wrapped := CouldNotDetermineVersionError(err)
		contextutils.LoggerFrom(ctx).Errorw(err.Error(), zap.Error(err))
		return "", wrapped
	}
	return release.GetTagName(), nil
}

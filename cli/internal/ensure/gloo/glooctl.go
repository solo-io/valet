package gloo

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/valet/cli/api"
	"github.com/solo-io/valet/cli/internal"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"runtime"
)

var _ Glooctl = new(glooctl)

var (
	GlooctlNotInstalledError = errors.Errorf("Glooctl not installed")
)

type Glooctl interface {
	Install(ctx context.Context) error
	Execute(args ...string) (string, error)
}

func NewGlooctl(valet *api.Valet, gloo *api.Gloo) Glooctl {
	return &glooctl{
		valet: valet,
		gloo:  gloo,
	}
}

type glooctl struct {
	valet *api.Valet
	gloo  *api.Gloo

	localPathToGlooctl string
}

func (g *glooctl) Execute(args ...string) (string, error) {
	if g.localPathToGlooctl == "" {
		return "", GlooctlNotInstalledError
	}
	return internal.ExecuteCmd(g.localPathToGlooctl, args...)
}

func (g *glooctl) Install(ctx context.Context) error {
	if g.valet.LocalArtifactsDir != "" {
		// try to use local artifact if it exists
		glooctlPath := filepath.Join(g.valet.LocalArtifactsDir, getAssetName())
		if _, err := os.Stat(glooctlPath); err == nil {
			g.localPathToGlooctl = glooctlPath
			return nil
		} else if !os.IsNotExist(err) {
			contextutils.LoggerFrom(ctx).Errorw("Error checking if glooctl was in artifacts directory", zap.Error(err))
			return err
		}
	}

	var glooctlVersion string
	if g.gloo.Enterprise || g.gloo.Version == "" {
		// use latest OS artifact if no version set and / or installing enterprise
		tag, err := getLatestTag(ctx, GlooRepo)
		if err != nil {
			return err
		}
		glooctlVersion = tag[1:]
	} else {
		glooctlVersion = g.gloo.Version
	}

	localPath, err := determineGlooctlFilepath(glooctlVersion)
	if err != nil {
		return err
	}
	g.localPathToGlooctl = localPath
	if _, err := os.Stat(g.localPathToGlooctl); err == nil {
		contextutils.LoggerFrom(ctx).Infow("Glooctl already exists", zap.String("localPath", localPath))
		return nil
	}
	if g.valet != nil && g.valet.ValetArtifacts {
		return downloadFromValet(ctx, glooctlVersion, localPath)
	} else {
		return downloadFromGithub(ctx, "v" + glooctlVersion, localPath)
	}
}

func getAssetName() string {
	return fmt.Sprintf("glooctl-%s-amd64", runtime.GOOS)
}

func getBinaryDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gloo", "bin"), nil
}

func determineGlooctlFilepath(version string) (string, error) {
	dir, err := getBinaryDir()
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("glooctl-%s", version)
	return filepath.Join(dir, filename), nil
}

func downloadFromValet(ctx context.Context, version, path string) error {
	downloader := NewUrlArtifactDownloader()
	remotePath := fmt.Sprintf("https://storage.googleapis.com/valet/artifacts/gloo/%s/glooctl-%s-amd64", version, runtime.GOOS)
	return downloader.Download(ctx, remotePath, path)
}

func downloadFromGithub(ctx context.Context, tag, path string) error {
	client := githubutils.GetClientWithOrWithoutToken(ctx)
	downloader := NewGithubArtifactDownloader(client, GlooRepo, tag)
	return downloader.Download(ctx, getAssetName(), path)
}

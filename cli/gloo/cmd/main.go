package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/kelseyhightower/envconfig"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/randutils"
	"github.com/solo-io/go-utils/testutils/kube"
	"github.com/solo-io/kube-cluster/cli/internal"
	"go.uber.org/zap"
	"io"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

func main() {
	ctx := internal.GetInitialContext()
	config := mustGetConfig(ctx)
	mustInstallGloo(ctx, config)
}

type EnvConfig struct {
	GlooVersion string `split_words:"true" required:"true"`
	Enterprise bool `default:"false"`
	LicenseKey string `split_words:"true"`
	GlooNamespace string `split_words:"true" default:"gloo-system"`
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

func mustWaitUntilPodsRunning(ctx context.Context, config EnvConfig) {
	contextutils.LoggerFrom(ctx).Infow("Waiting for pods")
	pods := kube.MustKubeClient().CoreV1().Pods(config.GlooNamespace)
	podsReady := func() (bool, error) {
		list, err := pods.List(v1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, pod := range list.Items {
			var podReady bool
			for _, cond := range pod.Status.Conditions {
				if cond.Type == v12.ContainersReady && cond.Status == v12.ConditionTrue {
					podReady = true
					break
				}
			}
			if !podReady {
				return false, nil
			}
		}
		return true, nil
	}
	failed := time.After(5 * time.Minute)
	notYetRunning := make(map[string]struct{})
	for {
		select {
		case <-failed:
			contextutils.LoggerFrom(ctx).Fatalf("timed out waiting for pods to come online: %v", notYetRunning)
		case <-time.After(time.Second / 2):
			notYetRunning = make(map[string]struct{})
			ready, err := podsReady()
			if err != nil {
				contextutils.LoggerFrom(ctx).Fatalw("error checking for ready pods", zap.Error(err))
			}
			if ready {
				return
			}
		}
	}
}

func mustInstallGloo(ctx context.Context, config EnvConfig) {
	glooInstalled := checkForGlooInstall(ctx, config)
	if glooInstalled {
		contextutils.LoggerFrom(ctx).Infow("Gloo is installed at the desired version")
		return
	}
	client := githubutils.GetClientWithOrWithoutToken(ctx)
	release := mustGetRelease(ctx, config, client)
	asset := mustGetAsset(ctx, config, release)
	filepath := fmt.Sprintf("glooctl-%s-%s", config.GlooVersion, randutils.RandString(4))
	mustDownloadFile(ctx, client, asset, filepath, config)
	mustChmod(ctx, filepath)
	mustInstall(ctx, filepath, config)
	mustWaitUntilPodsRunning(ctx, config)
}

func checkForGlooInstall(ctx context.Context, config EnvConfig) bool {
	kubeClient := kube.MustKubeClient()
	ns, err := kubeClient.CoreV1().Namespaces().Get(config.GlooNamespace, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false
		}
		contextutils.LoggerFrom(ctx).Fatalw("Error trying to get namespace", zap.Error(err), zap.String("ns", config.GlooNamespace))
	}
	if ns.Status.Phase != v12.NamespaceActive {
		contextutils.LoggerFrom(ctx).Fatalw("Namespace is not active", zap.Error(err), zap.Any("phase", ns.Status.Phase))
	}
	pods, err := kubeClient.CoreV1().Pods(config.GlooNamespace).List(v1.ListOptions{LabelSelector: "gloo"})
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Error listing pods", zap.Error(err))
	}
	if len(pods.Items) == 0 {
		contextutils.LoggerFrom(ctx).Fatalw("No Gloo pods")
	}
	for _, pod := range pods.Items {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == v12.ContainersReady && cond.Status != v12.ConditionTrue {
				contextutils.LoggerFrom(ctx).Fatalw("Gloo pods not ready")
			}
		}
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if strings.Contains(container.Image, config.GlooVersion[1:]) {
				return true
			}
		}
	}
	contextutils.LoggerFrom(ctx).Fatalw("Did not find any containers with the expected version",
		zap.String("expected", config.GlooVersion[1:]))
	return false
}

func mustInstall(ctx context.Context, filepath string, config EnvConfig) {
	contextutils.LoggerFrom(ctx).Infow("Running glooctl install")
	args := []string {"install", "gateway"}
	if config.Enterprise {
		args = append(args, "--license-key", config.LicenseKey)
	}
	out, err := internal.ExecuteCmd("./" + filepath, args...)
	if err != nil {
		contextutils.LoggerFrom(ctx).Infow("Failed to install gloo",
			zap.Error(err),
			zap.String("out", out))
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

func mustDownloadFile(ctx context.Context, client *github.Client, asset github.ReleaseAsset, filepath string, config EnvConfig) {
	contextutils.LoggerFrom(ctx).Infow("Downloading glooctl")
	rc, redirectUrl, err := client.Repositories.DownloadReleaseAsset(ctx, "solo-io", getRepo(config.Enterprise), asset.GetID())
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Could not download asset",
			zap.Error(err),
			zap.String("filepath", filepath),
			zap.Int64("assetId", asset.GetID()))
	}
	if rc != nil {
		err = copyReader(filepath, rc)
	} else {
		err = downloadFile(filepath, redirectUrl)
	}
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Could not download asset",
			zap.Error(err),
			zap.String("filepath", filepath),
			zap.Int64("assetId", asset.GetID()))
	}
}

func copyReader(filepath string, rc io.ReadCloser) error {
	defer rc.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, rc)
	return err
}

func downloadFile(filepath, url string) error {
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
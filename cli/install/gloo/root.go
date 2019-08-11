package gloo

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/randutils"
	"github.com/solo-io/go-utils/testutils/kube"
	"github.com/solo-io/kube-cluster/cli/internal"
	"github.com/solo-io/kube-cluster/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"io"
	v12 "k8s.io/api/core/v1"
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

func GlooCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gloo",
		Short: "ensures gloo is installed to namespace",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOptions(opts.Gloo); err != nil {
				return err
			}
			return ensureGloo(opts.Top.Ctx, opts.Gloo)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.Gloo.GlooVersion, "version", "v", "", "gloo version")
	cmd.PersistentFlags().StringVarP(&opts.Gloo.GlooNamespace, "namespace", "n", "gloo-system", "gloo namespace")
	cmd.PersistentFlags().BoolVarP(&opts.Gloo.Enterprise, "enterprise", "e", false, "install enterprise gloo")
	cmd.PersistentFlags().StringVar(&opts.Gloo.LicenseKey, "license-key", "", "enterprise gloo license key")
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func ensureGloo(ctx context.Context, config options.Gloo) error {

	glooInstalled, err := checkForGlooInstall(ctx, config)
	if err != nil {
		return err
	}
	if glooInstalled {
		contextutils.LoggerFrom(ctx).Infow("Gloo is installed at the desired version")
		return nil
	}
	client := githubutils.GetClientWithOrWithoutToken(ctx)
	release, err := getRelease(ctx, config, client)
	if err != nil {
		return err
	}
	asset, err := getAsset(ctx, config, release)
	if err != nil {
		return err
	}
	filepath := fmt.Sprintf("glooctl-%s-%s", config.GlooVersion, randutils.RandString(4))
	err = downloadAsset(ctx, client, asset, filepath, config)
	if err != nil {
		return err
	}
	err = chmod(ctx, filepath)
	if err != nil {
		return err
	}
	err = install(ctx, filepath, config)
	if err != nil {
		return err
	}
	return waitUntilPodsRunning(ctx, config)
}

func validateOptions(config options.Gloo) error {
	if config.GlooVersion == "" {
		return errors.Errorf("must specify a version to install")
	}
	if config.Enterprise && config.LicenseKey == "" {
		return errors.Errorf("must specify a license-key when installing enterprise gloo")
	}
	return nil
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

func waitUntilPodsRunning(ctx context.Context, config options.Gloo) error {
	contextutils.LoggerFrom(ctx).Infow("Waiting for pods")
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return err
	}
	pods := kubeClient.CoreV1().Pods(config.GlooNamespace)
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
			contextutils.LoggerFrom(ctx).Errorf("timed out waiting for pods to come online: %v", notYetRunning)
			return errors.Errorf("timed out waiting for pods to come online: %v", notYetRunning)
		case <-time.After(time.Second / 2):
			notYetRunning = make(map[string]struct{})
			ready, err := podsReady()
			if err != nil {
				contextutils.LoggerFrom(ctx).Errorw("error checking for ready pods", zap.Error(err))
				return err
			}
			if ready {
				return nil
			}
		}
	}
}

func checkForGlooInstall(ctx context.Context, config options.Gloo) (bool, error) {
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return false, err
	}
	ns, err := kubeClient.CoreV1().Namespaces().Get(config.GlooNamespace, v1.GetOptions{})
	if err != nil {
		if kubeerrs.IsNotFound(err) {
			return false, nil
		}
		contextutils.LoggerFrom(ctx).Errorw("Error trying to get namespace", zap.Error(err), zap.String("ns", config.GlooNamespace))
		return false, err
	}
	if ns.Status.Phase != v12.NamespaceActive {
		contextutils.LoggerFrom(ctx).Errorw("Namespace is not active", zap.Any("phase", ns.Status.Phase))
	}
	pods, err := kubeClient.CoreV1().Pods(config.GlooNamespace).List(v1.ListOptions{LabelSelector: "gloo"})
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error listing pods", zap.Error(err))
		return false, err
	}
	if len(pods.Items) == 0 {
		contextutils.LoggerFrom(ctx).Infow("No Gloo pods")
		return false, nil
	}
	for _, pod := range pods.Items {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == v12.ContainersReady && cond.Status != v12.ConditionTrue {
				contextutils.LoggerFrom(ctx).Infow("Gloo pods not ready")
				return false, nil
			}
		}
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if strings.Contains(container.Image, config.GlooVersion[1:]) {
				return true, nil
			}
		}
	}
	contextutils.LoggerFrom(ctx).Infow("Did not find any containers with the expected version",
		zap.String("expected", config.GlooVersion[1:]))
	return false, nil
}

func install(ctx context.Context, filepath string, config options.Gloo) error {
	contextutils.LoggerFrom(ctx).Infow("Running glooctl install")
	args := []string{"install", "gateway"}
	if config.Enterprise {
		args = append(args, "--license-key", config.LicenseKey)
	}
	out, err := internal.ExecuteCmd("./"+filepath, args...)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to install gloo",
			zap.Error(err),
			zap.String("out", out))
	}
	return err
}

func getRelease(ctx context.Context, config options.Gloo, client *github.Client) (*github.RepositoryRelease, error) {
	release, _, err := client.Repositories.GetReleaseByTag(ctx, "solo-io", getRepo(config.Enterprise), config.GlooVersion)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not download glooctl",
			zap.Error(err),
			zap.Bool("enterprise", config.Enterprise),
			zap.String("tag", config.GlooVersion))
		return nil, err
	}
	return release, nil
}

func getAsset(ctx context.Context, config options.Gloo, release *github.RepositoryRelease) (*github.ReleaseAsset, error) {
	desiredAsset := getAssetName()
	for _, asset := range release.Assets {
		if asset.GetName() == desiredAsset {
			return &asset, nil
		}
	}
	contextutils.LoggerFrom(ctx).Errorw("Could not find asset",
		zap.Bool("enterprise", config.Enterprise),
		zap.String("tag", config.GlooVersion),
		zap.String("assetName", desiredAsset))
	return nil, errors.Errorf("Could not find asset")
}

func chmod(ctx context.Context, filepath string) error {
	err := os.Chmod(filepath, os.ModePerm)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not make glooctl executable",
			zap.Error(err),
			zap.String("filepath", filepath))
	}
	return err
}

func downloadAsset(ctx context.Context, client *github.Client, asset *github.ReleaseAsset, filepath string, config options.Gloo) error {
	contextutils.LoggerFrom(ctx).Infow("Downloading glooctl")
	rc, redirectUrl, err := client.Repositories.DownloadReleaseAsset(ctx, "solo-io", getRepo(config.Enterprise), asset.GetID())
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not download asset",
			zap.Error(err),
			zap.String("filepath", filepath),
			zap.Int64("assetId", asset.GetID()))
		return err
	}
	if rc != nil {
		err = copyReader(filepath, rc)
	} else {
		err = downloadFile(filepath, redirectUrl)
	}
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not download asset",
			zap.Error(err),
			zap.String("filepath", filepath),
			zap.Int64("assetId", asset.GetID()))
	}
	return err
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

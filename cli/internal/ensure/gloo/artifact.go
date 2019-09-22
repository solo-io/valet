package gloo

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
)

type ArtifactDownloader interface {
	Download(ctx context.Context, remotePath, localPath string) error
}

var _ ArtifactDownloader = new(githubArtifactDownloader)

func NewGithubArtifactDownloader(client *github.Client, repoName, tag string) *githubArtifactDownloader {
	return &githubArtifactDownloader{
		client: client,
		repoName: repoName,
		tag: tag,
	}
}

type githubArtifactDownloader struct {
	client   *github.Client
	repoName string
	tag      string
}

func (d *githubArtifactDownloader) Download(ctx context.Context, remotePath, localPath string) error {
	release, err := d.getRelease(ctx, d.client)
	if err != nil {
		return err
	}
	asset, err := d.getAsset(ctx, release, remotePath)
	if err != nil {
		return err
	}
	return d.downloadAsset(ctx, asset, localPath)
}

func (d *githubArtifactDownloader) getRelease(ctx context.Context, client *github.Client) (*github.RepositoryRelease, error) {
	release, _, err := client.Repositories.GetReleaseByTag(ctx, "solo-io", d.repoName, d.tag)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not get release",
			zap.Error(err),
			zap.String("repo", d.repoName),
			zap.String("tag", d.tag))
		return nil, err
	}
	return release, nil
}

func (d *githubArtifactDownloader) getAsset(ctx context.Context, release *github.RepositoryRelease, remotePath string) (*github.ReleaseAsset, error) {
	for _, asset := range release.Assets {
		if asset.GetName() == remotePath {
			return &asset, nil
		}
	}
	contextutils.LoggerFrom(ctx).Errorw("Could not find asset",
		zap.String("repo", d.repoName),
		zap.String("tag", d.tag),
		zap.String("assetName", remotePath))
	return nil, errors.Errorf("Could not find asset")
}

func chmod(ctx context.Context, filepath string) error {
	err := os.Chmod(filepath, os.ModePerm)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not make artifact executable",
			zap.Error(err),
			zap.String("filepath", filepath))
	}
	return err
}

func (d* githubArtifactDownloader) downloadAsset(ctx context.Context, asset *github.ReleaseAsset, filepath string) error {
	contextutils.LoggerFrom(ctx).Infow("Downloading asset",
		zap.String("assetName", asset.GetName()),
		zap.String("localFile", filepath))
	rc, redirectUrl, err := d.client.Repositories.DownloadReleaseAsset(ctx, "solo-io", d.repoName, asset.GetID())
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
		return err
	}
	return chmod(ctx, filepath)
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

func NewUrlArtifactDownloader() *urlArtifactDownloader {
	return &urlArtifactDownloader{}
}

type urlArtifactDownloader struct {

}

var _ ArtifactDownloader = new(urlArtifactDownloader)

func (d *urlArtifactDownloader) Download(ctx context.Context, remotePath, localPath string) error {
	contextutils.LoggerFrom(ctx).Infow("Downloading glooctl", zap.String("remotePath", remotePath), zap.String("localPath", localPath))
	err := downloadFile(localPath, remotePath)
	if err != nil {
		return err
	}
	return chmod(ctx, localPath)
}
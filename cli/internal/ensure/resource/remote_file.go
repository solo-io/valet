package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/client"
)

type RemoteFile struct {
	RemotePath string `yaml:"remotePath"`
	LocalPath  string `yaml:"localPath"`
}

func (r *RemoteFile) Ensure(ctx context.Context) error {
	downloader := client.NewUrlArtifactDownloader()
	return downloader.Download(ctx, r.RemotePath, r.LocalPath)
}

func (r *RemoteFile) Teardown(ctx context.Context) error {
	return nil
}

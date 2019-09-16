package build

import (
	"fmt"
	"github.com/solo-io/go-utils/errors"
	"os"
	"os/exec"
)

const (
	ArtifactsDir = "_artifacts"
)

var (
	DefaultOs = []string{"linux"}

	getStorageLocation = func(product, version string) string {
		return fmt.Sprintf("gs://valet/artifacts/%s/%s/", product, version)
	}

	FailedToSyncArtifactsError = func(err error) error {
		return errors.Wrapf(err, "Failed to sync artifacts to bucket")
	}
)

func EnsureArtifactsDir() error {
	if err := os.RemoveAll(ArtifactsDir); !os.IsNotExist(err) {
		return err
	}
	return os.Mkdir(ArtifactsDir, os.ModePerm)
}

func SyncToGoogleStorage(product, version string) error {
	args := []string {
		"-m", "rsync", "-r",
		fmt.Sprintf("./%s/", ArtifactsDir),
		getStorageLocation(product, version),
	}
	cmd := exec.Command("gsutil", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(string(out))
		return FailedToSyncArtifactsError(err)
	}
	return nil
}
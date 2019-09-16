package build

import (
	"fmt"
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
)

func ensureArtifactsDir() error {
	if err := os.RemoveAll(ArtifactsDir); !os.IsNotExist(err) {
		return err
	}
	return os.Mkdir(ArtifactsDir, os.ModePerm)
}

func syncToGoogleStorage(product, version string) error {
	args := []string {
		"-m", "rsync", "-r",
		fmt.Sprintf("./%s/", ArtifactsDir),
		getStorageLocation(product, version),
	}
	cmd := exec.Command("gsutil", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(string(out))
		return err
	}
	return nil
}
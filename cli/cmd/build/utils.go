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

	getStorageDirectory = func(product, version string) string {
		return fmt.Sprintf("gs://valet/artifacts/%s/%s/", product, version)
	}

	getStorageLocation = func(product, version, filename string) string {
		return fmt.Sprintf("gs://valet/artifacts/%s/%s/%s", product, version, filename)
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
		getStorageDirectory(product, version),
	}
	cmd := exec.Command("gsutil", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(string(out))
		return err
	}
	return nil
}

func syncFileToGoogleStorage(product, version, filename string) error {
	fmt.Printf("Saving %s to google storage\n", filename)
	localFile := fmt.Sprintf("./%s/%s", ArtifactsDir, filename)
	remoteFile := getStorageLocation(product, version, filename)
	args := []string {"cp", localFile, remoteFile}
	cmd := exec.Command("gsutil", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(string(out))
		return err
	}
	return nil
}
package build

import (
	"os"
)

const (
	ArtifactsDir = "_artifacts"
)

var (
	DefaultOs = []string{"linux"}
)

func EnsureArtifactsDir() error {
	if _, err := os.Stat(ArtifactsDir); os.IsNotExist(err) {
		err = os.Mkdir(ArtifactsDir, os.ModePerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

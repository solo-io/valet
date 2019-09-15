package build

import (
	"fmt"
	"github.com/solo-io/valet/cli/cmd/build/artifacts"
	"github.com/solo-io/valet/cli/options"
	"os"
	"os/exec"
	"path/filepath"
)

func buildArtifacts(build artifacts.Build, opts options.Build) error {
	return buildGoArtifacts(build.Go, opts)
}

func buildGoArtifacts(goBuild artifacts.Go, opts options.Build) error {
	for _, binary := range goBuild.Binaries {
		if binary.Os == nil {
			binary.Os = DefaultOs
		}
		for _, osStr := range binary.Os {
			binaryName := fmt.Sprintf("%s-%s-%s", binary.Name, osStr, "amd64")
			parts := []string{
				"build",
				"-ldflags", fmt.Sprintf("-X %s=%s", goBuild.Version, opts.Version),
				"-gcflags", goBuild.GcFlags,
				"-o", filepath.Join(ArtifactsDir, binaryName),
				binary.Entrypoint,
			}
			cmd := exec.Command("go", parts...)
			cmd.Env = append(os.Environ(), goEnv(osStr)...)
			fmt.Printf("Building %s\n", binaryName)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf(string(output))
				return err
			}
		}
	}
	return nil
}

func goEnv(os string) []string {
	return []string{
		"CGO_ENABLED=0",
		"GOOS=" + os,
		"GOARCH=amd64",
	}
}

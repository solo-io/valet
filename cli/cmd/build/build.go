package build

import (
	"fmt"
	"github.com/solo-io/valet/cli/cmd/build/artifacts"
	"github.com/solo-io/valet/cli/options"
	"golang.org/x/sync/errgroup"
	"os"
	"os/exec"
	"path/filepath"
)

func buildArtifacts(build artifacts.Build, opts options.Build, productName string) error {
	return buildGoArtifacts(build.Go, opts, productName)
}

func buildGoArtifacts(goBuild artifacts.Go, opts options.Build, productName string) error {
	eg := errgroup.Group{}
	for _, binary := range goBuild.Binaries {
		b := binary
		eg.Go(func() error {
			return buildGoArtifact(goBuild, b, opts, productName)
		})
	}
	return eg.Wait()
}

func buildGoArtifact(goBuild artifacts.Go, binary artifacts.Binary, opts options.Build, productName string) error {
	if binary.Tests != nil {
		for _, test := range binary.Tests {
			testArgs := []string{
				"-r", test.Path,
			}
			cmd := exec.Command("ginkgo", testArgs...)
			fmt.Printf("Running tests: %s... ", test.Path)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Error: \n%s", string(output))
				return err
			}
			fmt.Printf("OK\n")
		}
	}
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
		if binary.Upload {
			if err := syncFileToGoogleStorage(productName, opts.Version, binaryName); err != nil {
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

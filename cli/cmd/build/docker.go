package build

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/solo-io/valet/cli/cmd/build/artifacts"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"os/exec"
	"path/filepath"
)

func docker(docker artifacts.Docker, opts options.Build) error {
	for _, registry := range docker.Registries {
		for _, container := range docker.Containers {
			retryFunc := func () error {
				return dockerContainer(registry, container, opts)
			}
			if err := retry.Do(retryFunc); err != nil {
				return err
			}
		}
	}
	return nil
}

func dockerContainer(registry string, container artifacts.Container, opts options.Build) error {
	dockerTag := fmt.Sprintf("%s:%s", filepath.Join(registry, container.Name), opts.Version)
	cmd := exec.Command("docker", "build", "-t", dockerTag, "-f", container.Dockerfile, "_artifacts")
	internal.Report("Building docker container %s", dockerTag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		internal.Report("Error: %s", string(output))
		return err
	}

	if opts.SkipDockerPush {
		return nil
	}

	cmd = exec.Command("docker", "push", dockerTag)
	internal.Report("Pushing docker container %s", dockerTag)
	output, err = cmd.CombinedOutput()
	if err != nil {
		internal.Report("Error: %s", string(output))
		return err
	}
	return nil
}

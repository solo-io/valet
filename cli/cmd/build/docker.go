package build

import (
	"fmt"
	"github.com/solo-io/valet/cli/cmd/build/artifacts"
	"github.com/solo-io/valet/cli/options"
	"os/exec"
	"path/filepath"
)

func docker(docker artifacts.Docker, opts options.Build) error {
	for _, registry := range docker.Registries {
		for _, container := range docker.Containers {
			if err := dockerContainer(registry, container, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func dockerContainer(registry string, container artifacts.Container, opts options.Build) error {
	dockerTag := fmt.Sprintf("%s:%s", filepath.Join(registry, container.Name), opts.Version)
	cmd := exec.Command("docker", "build", "-t", dockerTag, "-f", container.Dockerfile, "_artifacts")
	fmt.Printf("Building docker container %s\n", dockerTag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(string(output))
		return err
	}

	if opts.SkipDockerPush {
		return nil
	}

	cmd = exec.Command("docker", "push", dockerTag)
	fmt.Printf("Pushing docker container %s\n", dockerTag)
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(string(output))
		return err
	}
	return nil
}

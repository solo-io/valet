package build

import (
	"fmt"
	"github.com/solo-io/valet/cli/cmd/build/artifacts"
	"github.com/solo-io/valet/cli/options"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

func helm(helm artifacts.Helm, opts options.Build) error {
	for _, chart := range helm.Charts {
		if err := helmChart(chart, opts); err != nil {
			return err
		}
	}
	return nil
}

func helmChart(chart artifacts.Chart, opts options.Build) error {
	generate := exec.Command("go", "run", chart.Generator, opts.Version)
	fmt.Printf("Generating helm chart %s\n", chart.Name)
	output, err := generate.CombinedOutput()
	if err != nil {
		fmt.Printf(string(output))
		return err
	}
	pkg := exec.Command("helm", "package", "--destination", ArtifactsDir, chart.Directory)
	output, err = pkg.CombinedOutput()
	if err != nil {
		fmt.Printf(string(output))
		return err
	}

	for _, manifest := range chart.Manifests {
		cmd := []string{"helm", "template",
			chart.Directory,
			"--namespace", "gloo-system",
			"--set", "namespace.create=true",
		}
		if manifest.Values != "" {
			valuesPath := filepath.Join(chart.Directory, manifest.Values)
			cmd = append(cmd, "--values", valuesPath)
		}

		fmt.Printf("Generating manifest %s\n", manifest.Name)
		manifestCmd := exec.Command(cmd[0], cmd[1:]...)
		output, err = manifestCmd.CombinedOutput()
		if err != nil {
			fmt.Printf(string(output))
			return err
		}
		manifestPath := filepath.Join(ArtifactsDir, manifest.Name)
		err := ioutil.WriteFile(manifestPath, output, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

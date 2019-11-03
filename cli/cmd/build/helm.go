package build

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/solo-io/valet/cli/cmd/build/artifacts"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
)

func helm(helm artifacts.Helm, opts options.Build, productName string) error {
	for _, chart := range helm.Charts {
		if err := helmChart(chart, opts, productName); err != nil {
			return err
		}
	}
	return nil
}

func helmChart(chart artifacts.Chart, opts options.Build, productName string) error {
	generate := exec.Command("go", "run", chart.Generator, opts.Version)
	chartFilename := fmt.Sprintf("%s-%s.tgz", chart.Name, opts.Version)
	internal.Report("Generating helm chart %s", chartFilename)
	output, err := generate.CombinedOutput()
	if err != nil {
		internal.Report("Error: %s", string(output))
		return err
	}
	pkg := exec.Command("helm", "package", "--destination", ArtifactsDir, chart.Directory)
	output, err = pkg.CombinedOutput()
	if err != nil {
		internal.Report("Error: %s", string(output))
		return err
	}
	if chart.Upload {
		if err := syncFileToGoogleStorage(productName, opts.Version, chartFilename); err != nil {
			return err
		}
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

		internal.Report("Generating manifest %s", manifest.Name)
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
		if chart.Upload {
			if err := syncFileToGoogleStorage(productName, opts.Version, manifest.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

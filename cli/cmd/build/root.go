package build

import (
	"fmt"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/build/artifacts"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"time"
)

var (
	MustProvideFileError = errors.Errorf("Must provide a file option")
	MustProvideVersionError = errors.Errorf("Must provide a version option")
	CouldNotPrepareArtifactsDirError = func(err error) error {
		return errors.Wrapf(err, "Could not prepare artifacts directory")
	}
	CouldNotReadArtifactsFileError = func(err error) error {
		return errors.Wrapf(err, "Could not read artifacts file")
	}
	CouldNotBuildArtifactsError = func(err error) error {
		return errors.Wrapf(err, "Could not build artifacts")
	}
	CouldNotBuildContainersError = func(err error) error {
		return errors.Wrapf(err, "Could not build containers")
	}
	CouldNotCreateManifestsError = func(err error) error {
		return errors.Wrapf(err, "Could not create charts and manifests")
	}
)

func Build(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "build, package, and publish artifacts for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return build(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	cmd.PersistentFlags().StringVarP(&opts.Build.File, "file", "f", artifacts.DefaultArtifactsFile, "path to file containing artifacts spec")
	cmd.PersistentFlags().StringVarP(&opts.Build.Version, "version", "v", "", "artifacts version")
	cmd.PersistentFlags().BoolVar(&opts.Build.SkipDockerPush, "skip-docker-push", false, "skip pushing docker images")
	return cmd
}

func build(opts *options.Options) error {
	if opts.Build.File == "" {
		return MustProvideFileError
	}
	if opts.Build.Version == "" {
		return MustProvideVersionError
	}
	if err := EnsureArtifactsDir(); err != nil {
		return CouldNotPrepareArtifactsDirError(err)
	}
	artifactsCfg, err := artifacts.ReadArtifactsConfig(opts.Build.File)
	if err != nil {
		return CouldNotReadArtifactsFileError(err)
	}
	fmt.Printf("Artifacts version: %s\n", opts.Build.Version)
	fmt.Printf("Starting artifacts build [%s]\n", time.Now().Format(time.RFC3339))
	if err := buildArtifacts(artifactsCfg.Build, opts.Build); err != nil {
		return CouldNotBuildArtifactsError(err)
	}
	fmt.Printf("Finished artifacts build [%s]\n", time.Now().Format(time.RFC3339))
	if err := docker(artifactsCfg.Docker, opts.Build); err != nil {
		return CouldNotBuildContainersError(err)
	}
	fmt.Printf("Finished docker [%s]\n", time.Now().Format(time.RFC3339))
	if err := helm(artifactsCfg.Helm, opts.Build); err != nil {
		return CouldNotCreateManifestsError(err)
	}
	fmt.Printf("Finished charts and manifests [%s]\n", time.Now().Format(time.RFC3339))
	return nil
}
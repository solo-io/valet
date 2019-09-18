package build

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/build/artifacts"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

var (
	MustProvideFileError             = errors.Errorf("Must provide a file option")
	MustProvideVersionError          = errors.Errorf("Must provide a version option")
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
	FailedToSyncArtifactsError = func(err error) error {
		return errors.Wrapf(err, "Failed to sync artifacts to bucket")
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
	if err := ensureArtifactsDir(); err != nil {
		return CouldNotPrepareArtifactsDirError(err)
	}
	artifactsCfg, err := artifacts.ReadArtifactsConfig(opts.Build.File)
	if err != nil {
		return CouldNotReadArtifactsFileError(err)
	}
	internal.Report("Artifacts version: %s", opts.Build.Version)
	internal.Report("Starting artifacts build...")
	if err := buildArtifacts(artifactsCfg.Build, opts.Build, artifactsCfg.ProductName); err != nil {
		return CouldNotBuildArtifactsError(err)
	}
	internal.Report("Finished artifacts build")
	if err := docker(artifactsCfg.Docker, opts.Build); err != nil {
		return CouldNotBuildContainersError(err)
	}
	internal.Report("Finished docker")
	if err := helm(artifactsCfg.Helm, opts.Build, artifactsCfg.ProductName); err != nil {
		return CouldNotCreateManifestsError(err)
	}
	internal.Report("Finished charts and manifests")
	return nil
}

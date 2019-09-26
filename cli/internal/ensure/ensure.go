package ensure

import (
	"context"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/api"
	cert_manager "github.com/solo-io/valet/cli/internal/ensure/cert-manager"
	"github.com/solo-io/valet/cli/internal/ensure/cluster/gke"
	"github.com/solo-io/valet/cli/internal/ensure/cluster/minikube"
	"github.com/solo-io/valet/cli/internal/ensure/demo/petclinic"
	"github.com/solo-io/valet/cli/internal/ensure/gloo"
	"github.com/solo-io/valet/cli/internal/ensure/resources"
	"github.com/solo-io/valet/cli/internal/ensure/workflow"
)

var (
	NoClusterDefinedError = errors.Errorf("No cluster defined")
	NoGlooDefinedError    = errors.Errorf("No gloo defined")
)

type Ensurer interface {
	Ensure(ctx context.Context, valet *api.Valet, cfg *api.EnsureConfig) error
}

var _ Ensurer = new(ensurer)

func NewEnsurer() Ensurer {
	return &ensurer{}
}

type ensurer struct{}

func (e *ensurer) Ensure(ctx context.Context, valet *api.Valet, cfg *api.EnsureConfig) error {
	if cfg.Cluster != nil {
		var clusterErr error
		if cfg.Cluster.GKE != nil {
			clusterErr = gke.EnsureGke(ctx, cfg.Cluster.GKE)
		} else if cfg.Cluster.Minikube != nil {
			clusterErr = minikube.EnsureMinikube(ctx, cfg.Cluster.Minikube)
		} else {
			return NoClusterDefinedError
		}

		if clusterErr != nil {
			return clusterErr
		}
	}

	if cfg.CertManager != nil {
		if err := cert_manager.EnsureCertManager(ctx, cfg.CertManager); err != nil {
			return err
		}
	}

	var glooManager gloo.GlooManager
	if cfg.Gloo != nil {
		glooManager = gloo.NewGlooManager(valet, cfg.Gloo)
		err := gloo.EnsureGloo(ctx, glooManager, cfg.Gloo)
		if err != nil {
			return err
		}
	}

	for _, work := range cfg.Workflows {
		if glooManager == nil {
			return NoGlooDefinedError
		}
		workflowRunner := workflow.NewWorkflowRunner(glooManager)
		if err := workflowRunner.Run(ctx, work); err != nil {
			return err
		}
	}

	if cfg.Demos != nil {
		if cfg.Demos.Petclinic != nil {
			err := petclinic.EnsurePetclinic(ctx, cfg.Demos.Petclinic)
			if err != nil {
				return err
			}
		}
	}

	if cfg.Resources != nil {
		err := resources.EnsureResources(ctx, cfg.Resources)
		if err != nil {
			return err
		}
	}

	return nil
}



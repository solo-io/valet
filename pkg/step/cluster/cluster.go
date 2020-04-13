package cluster

import (
	"fmt"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/render"

	errors "github.com/rotisserie/eris"
)

type ClusterResource interface {
	Ensure(ctx *api.WorkflowContext, values render.Values) error
	Teardown(ctx *api.WorkflowContext, values render.Values) error
	SetContext(ctx *api.WorkflowContext, values render.Values) error
}

var (
	_ api.Step = new(EnsureCluster)

	NoClusterDefinedError = errors.Errorf("no cluster defined")
)

type clusterStep struct {
	Minikube *Minikube `json:"minikube"`
	GKE      *GKE      `json:"gke"`
	EKS      *EKS      `json:"eks"`
}

type EnsureCluster clusterStep

func (e *EnsureCluster) GetDescription(ctx *api.WorkflowContext, values render.Values) (string, error) {
	if e.Minikube != nil {
		return fmt.Sprintf("Ensuring minikube cluster"), nil
	} else if e.GKE != nil {
		return fmt.Sprintf("Ensuring GKE cluster"), nil
	} else if e.EKS != nil {
		return fmt.Sprintf("Ensuring EKS cluster"), nil
	}
	return "", NoClusterDefinedError
}

func (e *EnsureCluster) Run(ctx *api.WorkflowContext, values render.Values) error {
	if e.Minikube != nil {
		return e.Minikube.Ensure(ctx, values)
	} else if e.GKE != nil {
		return e.GKE.Ensure(ctx, values)
	} else if e.EKS != nil {
		return e.EKS.Ensure(ctx, values)
	}
	return NoClusterDefinedError
}

func (e *EnsureCluster) GetDocs(ctx *api.WorkflowContext, values render.Values, flags render.Flags) (string, error) {
	panic("implement me")
}

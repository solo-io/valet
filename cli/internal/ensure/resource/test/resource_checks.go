package resource

import (
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/cluster"
	"github.com/solo-io/valet/cli/internal/ensure/resource/workflow"
)

var (
	_ resource.Resource = new(cluster.Cluster)
	_ resource.Resource = new(cluster.GKE)
	_ resource.Resource = new(cluster.Minikube)

	_ resource.Resource = new(application.Application)
	_ resource.Resource = new(application.Resource)
	_ resource.Resource = new(application.Secret)
	_ resource.Resource = new(application.Patch)

	_ resource.Resource = new(workflow.Workflow)
	_ resource.Resource = new(workflow.Condition)
	_ resource.Resource = new(workflow.Curl)
	_ resource.Resource = new(workflow.DnsEntry)
	_ resource.Resource = new(workflow.Step)
)

package application

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ns = "ns"
)

var (
	_ resource.Resource = new(Namespace)
	_ Renderable = new(Namespace)
)

type Namespace struct {
	Name        string            `yaml:"name" valet:"key=Namespace"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}

func (n *Namespace) Ensure(ctx context.Context, inputs render.InputParams, command cmd.Factory) error {
	cmd.Stdout().Println("Ensuring namespace %s", n.Name)
	err := command.Kubectl().Create(ns).WithName(n.Name).DryRunAndApply(ctx, command)
	if err != nil {
		return err
	}
	for k, v := range n.Labels {
		labelString := fmt.Sprintf("%s=%s", k, v)
		if err := command.Kubectl().With("label", "ns", n.Name, labelString, "--overwrite").Cmd().Run(ctx); err != nil {
			return err
		}
	}
	for k, v := range n.Annotations {
		annotationString := fmt.Sprintf("%s=%s", k, v)
		if err := command.Kubectl().With("annotate", "ns", n.Name, annotationString, "--overwrite").Cmd().Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (n *Namespace) Teardown(ctx context.Context, inputs render.InputParams, command cmd.Factory) error {
	cmd.Stdout().Println("Tearing down namespace %s", n.Name)
	return command.Kubectl().Delete(ns).WithName(n.Name).Cmd().Run(ctx)
}

func (n *Namespace) Render(ctx context.Context, inputs render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error) {
	if err := inputs.Values.RenderFields(n); err != nil {
		return nil, err
	}
	namespace := corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        n.Name,
			Labels:      n.Labels,
			Annotations: n.Annotations,
		},
	}
	r, err := kuberesource.ConvertToUnstructured(&namespace)
	if err != nil {
		return nil, err
	}
	return kuberesource.UnstructuredResources{r}, nil
}
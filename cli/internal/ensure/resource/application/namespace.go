package application

import (
	"context"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	_ Renderable = new(Namespace)
)

type Namespace struct {
	Name        string            `yaml:"name" valet:"key=Namespace"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}

func (n *Namespace) Render(ctx context.Context, inputs render.InputParams) (kuberesource.UnstructuredResources, error) {
	if err := inputs.RenderFields(n); err != nil {
		return nil, err
	}
	cmd.Stdout().Println("Rendering namespace %s", n.Name)
	namespace := corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
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

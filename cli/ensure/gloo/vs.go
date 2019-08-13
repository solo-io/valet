package gloo

import (
	"context"
	"github.com/solo-io/valet/cli/options"
)

type UiVirtualServiceCreator interface {
	Create(ctx context.Context, gloo options.Gloo) error
}

var _ UiVirtualServiceCreator = new(kubectlUiVirtualServiceCreator)

type kubectlUiVirtualServiceCreator struct {

}

const (
	GlooUiVirtualService = ""
)

func (k *kubectlUiVirtualServiceCreator) Create(ctx context.Context, gloo options.Gloo) error {

}








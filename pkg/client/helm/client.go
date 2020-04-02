package helm

import "github.com/solo-io/go-utils/installutils/helminstall"

//go:generate mockgen -destination ./mocks/helm_client_mock.go github.com/solo-io/valet/pkg/client/helm Client

type Client interface {
	Install(config *helminstall.InstallerConfig) error
}

func NewClient() *helmClient {
	return &helmClient{}
}

type helmClient struct {
}

func (h *helmClient) Install(config *helminstall.InstallerConfig) error {
	inst := helminstall.MustInstaller()
	return inst.Install(config)
}

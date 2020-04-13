package helm

import (
	errors "github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helminstall"
	"helm.sh/helm/v3/pkg/release"
)

//go:generate mockgen -destination ./mocks/helm_client_mock.go github.com/solo-io/valet/pkg/client/helm Client

type Client interface {
	Install(config *helminstall.InstallerConfig) error
	GetRelease(releaseName, releaseNamespace string) (*release.Release, error)
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

func (h *helmClient) GetRelease(releaseName, releaseNamespace string)  (*release.Release, error) {
	client := helminstall.DefaultHelmClient()
	releaseLister, err := client.ReleaseList("", "", releaseNamespace)
	if err != nil {
		return nil, err
	}
	releases, err := releaseLister.Run()
	if err != nil {
		return nil, err
	}
	for _, release := range releases {
		if release.Name == releaseName {
			return release, nil
		}
	}
	return nil, errors.Errorf("Release %s not found in namespace %s", releaseName, releaseNamespace)
}
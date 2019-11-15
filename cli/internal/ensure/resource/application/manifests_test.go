package application_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var _ = Describe("Manifests", func() {

	const (
		registryName              = "test-registry"
		registryPath              = "test"
		manifestRegistryPath      = "files/manifests/petclinic.yaml"
		otherManifestRegistryPath = "files/manifests/static-upstream.yaml"
	)

	var (
		ctx          = context.TODO()
		testRegistry = render.DirectoryRegistry{
			Path: registryPath,
		}
	)

	Context("render", func() {

		It("should load multiple manifests from a registry", func() {
			manifests := &application.Manifests{
				RegistryName: registryName,
				Paths:        []string{
					manifestRegistryPath,
					otherManifestRegistryPath,
				},
			}
			input := render.InputParams{}
			input.SetRegistry(registryName, &testRegistry)
			resources, err := manifests.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(7))
		})

	})
})

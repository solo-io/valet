package application_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var _ = Describe("Refs", func() {

	const (
		appPath = "test/apps/example-default-app.yaml"
		appRegistryPath = "apps/example-registry-app.yaml"

		registryName = "test-registry"
		registryPath = "test"
	)

	var (
		ctx = context.TODO()
		testRegistry = render.LocalRegistry{
			WorkingDirectory: registryPath,
		}
	)

	Context("render", func() {

		It("should correctly load an app from a registry", func() {
			ref := &application.Ref{
				RegistryName: registryName,
				Path: appRegistryPath,
			}
			input := render.InputParams{}
			input.SetRegistry(registryName, &testRegistry)
			resources, err := ref.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(6))
		})

		It("should correctly load an app from the default registry", func() {
			ref := &application.Ref{
				Path: appPath,
			}
			input := render.InputParams{}
			resources, err := ref.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(6))
		})
	})
})

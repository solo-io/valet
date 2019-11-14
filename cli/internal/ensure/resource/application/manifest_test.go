package application_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var _ = Describe("Manifest", func() {

	const (
		manifestPath = "test/files/manifests/petclinic.yaml"
		manifestUrl  = "https://raw.githubusercontent.com/sololabs/demos/b523571c66057a5591bce22ad896729f1fee662b/petclinic_demo/petclinic.yaml"

		registryName         = "test-registry"
		registryPath         = "test"
		manifestRegistryPath = "files/manifests/petclinic.yaml"
	)

	var (
		ctx          = context.TODO()
		emptyInput   = render.InputParams{}
		testRegistry = render.DirectoryRegistry{
			Path: registryPath,
		}
	)

	Context("render", func() {
		It("should load a manifest exactly from URL", func() {
			manifest := &application.Manifest{
				Path: manifestUrl,
			}
			resources, err := manifest.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(2))
		})

		It("should return an error for a fake path", func() {
			manifest := &application.Manifest{
				Path: "path/to/my/fake/manifest.yaml",
			}
			_, err := manifest.Render(ctx, emptyInput)
			Expect(err).NotTo(BeNil())
		})

		It("should load a manifest exactly from URL provided with the Path value", func() {
			manifest := &application.Manifest{}
			input := render.InputParams{
				Values: render.Values{
					render.PathKey: manifestUrl,
				},
			}
			resources, err := manifest.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(2))
		})

		It("should load a manifest from relative path", func() {
			manifest := &application.Manifest{
				Path: manifestPath,
			}
			resources, err := manifest.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(6))
		})

		It("should load a manifest from a relative path provided with the Path value", func() {
			manifest := &application.Manifest{}
			input := render.InputParams{
				Values: render.Values{
					render.PathKey: manifestPath,
				},
			}
			resources, err := manifest.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(6))
		})

		It("should load a manifest from a different registry", func() {
			manifest := &application.Manifest{
				RegistryName: registryName,
				Path:         manifestRegistryPath,
			}
			input := render.InputParams{}
			input.SetRegistry(registryName, &testRegistry)
			resources, err := manifest.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(6))
		})

		It("should load a manifest URL from a different registry", func() {
			// URLs should always behave the same, regardless of registry
			manifest := &application.Manifest{
				RegistryName: registryName,
				Path:         manifestUrl,
			}
			input := render.InputParams{}
			input.SetRegistry(registryName, &testRegistry)
			resources, err := manifest.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(2))
		})
	})
})

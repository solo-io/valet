package application_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var _ = Describe("template", func() {

	const (
		templatePath = "test/files/templates/cert.yaml"

		registryName         = "test-registry"
		registryPath         = "test"
		templateRegistryPath = "files/templates/cert.yaml"

		valueDomain    = "test.com"
		valueNamespace = "test-namespace"
	)

	var (
		ctx          = context.TODO()
		emptyInput   = render.InputParams{}
		testRegistry = render.DirectoryRegistry{
			Path: registryPath,
		}
		values = render.Values{
			render.DomainKey:    valueDomain,
			render.NamespaceKey: valueNamespace,
		}
		input = render.InputParams{Values: values}
	)

	Context("render", func() {

		It("returns error if required values not provided", func() {
			template := &application.Template{
				Path: templatePath,
			}
			_, err := template.Render(ctx, emptyInput)
			Expect(err).NotTo(BeNil())
		})

		It("returns error for bad template path", func() {
			template := &application.Template{
				Path: "path/to/my/fake/template.yaml",
			}
			_, err := template.Render(ctx, emptyInput)
			Expect(err).NotTo(BeNil())
		})

		It("should load a template", func() {
			template := &application.Template{
				Path: templatePath,
			}
			resources, err := template.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(1))
			actual := resources[0]
			Expect(actual.GetName()).To(Equal(valueDomain))
			Expect(actual.GetNamespace()).To(Equal(valueNamespace))
		})

		It("should load a template from a non-default registry", func() {
			template := &application.Template{
				Path:         templateRegistryPath,
				RegistryName: registryName,
			}
			input.SetRegistry(registryName, &testRegistry)
			resources, err := template.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(1))
			actual := resources[0]
			Expect(actual.GetName()).To(Equal(valueDomain))
			Expect(actual.GetNamespace()).To(Equal(valueNamespace))
		})

	})
})

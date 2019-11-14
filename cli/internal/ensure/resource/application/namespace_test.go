package application_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var _ = Describe("Namespace", func() {

	const (
		namespace = "test-namespace"
	)

	var (
		ctx        = context.TODO()
		emptyInput = render.InputParams{}
		labels     = map[string]string{
			"foo": "bar",
		}
		annotations = map[string]string{
			"baz": "bat",
		}
	)

	Context("render", func() {
		It("handles all fields", func() {
			ns := application.Namespace{
				Name:        namespace,
				Labels:      labels,
				Annotations: annotations,
			}
			resources, err := ns.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(1))
			actual := resources[0]
			Expect(actual.GetLabels()).To(Equal(labels))
			Expect(actual.GetAnnotations()).To(Equal(annotations))
			Expect(actual.GetName()).To(Equal(namespace))
		})

		It("renders name from Namespace value", func() {
			ns := application.Namespace{}
			input := render.InputParams{
				Values: render.Values{
					render.NamespaceKey: namespace,
				},
			}
			resources, err := ns.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(1))
			actual := resources[0]
			Expect(actual.GetLabels()).To(BeNil())
			Expect(actual.GetAnnotations()).To(BeNil())
			Expect(actual.GetName()).To(Equal(namespace))
		})
	})
})

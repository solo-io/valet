package resource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var _ = Describe("Manifest", func() {

	Context("default-ns", func() {
		var (
			appRef = application.Ref{
				Flags: []string{"foo1"},
				Values: render.Values{
					"foo2": "foo3",
				},
				Path: "foo4",
			}
			templatedAppRef = application.Ref{
				Path: "{{ .RandomValue }}",
			}
		)

		It("should do nothing when app ref fully provided", func() {
			a := appRef
			err := emptyValues.RenderFields(&a)
			Expect(err).To(BeNil())
			Expect(a).To(Equal(appRef))
		})

		It("should use values for namespace", func() {
			a := templatedAppRef
			err := values.RenderFields(&a)
			Expect(err).To(BeNil())
			Expect(a.Path).To(Equal(randomValue))
		})
	})
})

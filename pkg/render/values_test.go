package render_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

var _ = Describe("Values", func() {

	const (
		namespace   = "test-namespace"
		version     = "test-version"
		timeout     = "240s"
		randomValue = "foo"
		hostedZone  = "hosted-zone"
		domain      = "test-domain"
	)

	var (
		emptyValues = render.Values{}
		values      = render.Values{
			render.NamespaceKey:  namespace,
			render.VersionKey:    version,
			"Timeout":            timeout,
			"RandomValue":        randomValue,
			render.DomainKey:     domain,
			render.HostedZoneKey: hostedZone,
		}
		runner = cmd.DefaultCommandRunner()
	)

	Context("merge values", func() {
		It("doesn't modify the input", func() {
			input := emptyValues
			output := input.MergeValues(values)
			Expect(output).Should(Equal(values))
			Expect(input).Should(Equal(render.Values{}))
		})

		It("doesn't override values already supplied", func() {
			input := values
			otherValues := input.DeepCopy()
			otherValues[render.NamespaceKey] = "other-namespace"
			output := input.MergeValues(otherValues)
			Expect(output).Should(Equal(values))
		})

	})

	It("works for values referencing other values", func() {
		vals := render.Values{
			"Namespace":          "gloo-system",
			"UpstreamName":       "template:{{ .Namespace }}-apiserver-ui-8080",
			"UpstreamNamespace":  "key:Namespace",
			"VirtualServiceName": "glooui",
			"Domain":             "glooui.testing.valet.corp.solo.io",
		}
		Expect(vals.GetValue("Namespace", runner)).To(Equal("gloo-system"))
		Expect(vals.GetValue("UpstreamName", runner)).To(Equal("gloo-system-apiserver-ui-8080"))
		Expect(vals.GetValue("UpstreamNamespace", runner)).To(Equal("gloo-system"))
		Expect(vals.GetValue("VirtualServiceName", runner)).To(Equal("glooui"))
		Expect(vals.GetValue("Domain", runner)).To(Equal("glooui.testing.valet.corp.solo.io"))
	})
})

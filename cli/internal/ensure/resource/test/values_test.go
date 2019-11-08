package resource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"github.com/solo-io/valet/cli/internal/ensure/resource/workflow"
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
	)

	Context("merge values", func() {
		It("doesn't modify the input", func() {
			input := render.InputParams{
				Values: emptyValues,
			}
			output := input.MergeValues(values)
			Expect(output.Values).Should(Equal(values))
			Expect(input.Values).Should(Equal(render.Values{}))
		})

		It("doesn't override values already supplied", func() {
			input := render.InputParams{
				Values: values,
			}
			otherValues := input.DeepCopy().Values
			otherValues[render.NamespaceKey] = "other-namespace"
			output := input.MergeValues(otherValues)
			Expect(output.Values).Should(Equal(values))
		})

	})

	Context("conditions", func() {
		var (
			emptyCondition = workflow.Condition{}
			condition      = workflow.Condition{
				Timeout:   "foo1",
				Type:      "foo2",
				Namespace: "foo3",
				Name:      "foo4",
				Value:     "foo5",
				Jsonpath:  "foo6",
			}
		)

		It("should do nothing when condition fully provided", func() {
			c := condition
			err := emptyValues.RenderFields(&c)
			Expect(err).To(BeNil())
			Expect(c).To(Equal(condition))
		})

		It("should render templated values for conditions", func() {
			c := workflow.Condition{
				Timeout: "{{ .Timeout }}",
				Name:    "{{ .RandomValue }}",
			}
			err := values.RenderFields(&c)
			Expect(err).To(BeNil())
			Expect(c.Timeout).To(Equal(timeout))
			Expect(c.Name).To(Equal(randomValue))
		})

		It("should render default timeout for conditions", func() {
			c := emptyCondition
			err := emptyValues.RenderFields(&c)
			Expect(err).To(BeNil())
			Expect(c.Timeout).To(Equal(workflow.DefaultTimeout))
		})
	})

	Context("helm charts", func() {
		var (
			emptyHelmChart = application.HelmChart{}
			helmChart      = application.HelmChart{
				Namespace: "foo1",
				Version:   "foo2",
				ChartName: "foo3",
				RepoUrl:   "foo4",
				RepoName:  "foo5",
			}
		)

		It("should do nothing when helm chart fully provided", func() {
			h := helmChart
			err := emptyValues.RenderFields(&h)
			Expect(err).To(BeNil())
			Expect(h).To(Equal(helmChart))
		})

		It("should use values for helm charts", func() {
			h := emptyHelmChart
			err := values.RenderFields(&h)
			Expect(err).To(BeNil())
			Expect(h.Version).To(Equal(version))
			Expect(h.Namespace).To(Equal(namespace))
		})
	})

	Context("secrets", func() {
		var (
			emptySecret = application.Secret{}
			secret      = application.Secret{
				Namespace: "foo1",
				Name:      "foo2",
				Entries: map[string]application.SecretValue{
					"foo3": {
						File: "foo4",
					},
					"foo5": {
						EnvVar: "FOO6",
					},
					"foo7": {
						GcloudKmsEncryptedFile: &application.GcloudKmsEncryptedFile{
							Key:            "foo8",
							Keyring:        "foo9",
							GcloudProject:  "foo10",
							CiphertextFile: "foo11",
						},
					},
				},
			}
		)

		It("should do nothing when secret fully provided", func() {
			s := secret
			err := emptyValues.RenderFields(&s)
			Expect(err).To(BeNil())
			Expect(s).To(Equal(secret))
		})

		It("should use values for namespace", func() {
			s := emptySecret
			err := values.RenderFields(&s)
			Expect(err).To(BeNil())
			Expect(s.Namespace).To(Equal(namespace))
		})
	})

	Context("dns entries", func() {
		var (
			emptyDnsEntry = workflow.DnsEntry{}
			dnsEntry      = workflow.DnsEntry{
				HostedZone: "foo1",
				Domain:     "foo2",
				Service: workflow.ServiceRef{
					Name:      "foo3",
					Namespace: "foo4",
				},
			}
		)

		It("should do nothing when dns entry fully provided", func() {
			d := dnsEntry
			err := emptyValues.RenderFields(&d)
			Expect(err).To(BeNil())
			Expect(d).To(Equal(dnsEntry))
		})

		It("should use values for namespace", func() {
			d := emptyDnsEntry
			err := values.RenderFields(&d)
			Expect(err).To(BeNil())
			Expect(d.Domain).To(Equal(domain))
			Expect(d.HostedZone).To(Equal(hostedZone))
		})
	})

	Context("app ref", func() {
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

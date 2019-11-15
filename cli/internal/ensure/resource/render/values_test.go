package render

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ = Describe("values", func() {

	const (
		testKey       = "TestKey"
		testValue     = "TestValue"
		templateKey   = "Template"
		templateValue = "TemplateValue"
		testTemplate  = "{{ .Template }}"
	)

	var (
		// ctx    = context.TODO()
		runner = cmd.DefaultCommandRunner()
		values Values
	)

	BeforeEach(func() {
		values = Values{}
	})

	Context("tags", func() {
		type testStruct struct {
			One   string `valet:"template,key=TestKey,default=one"`
			Two   string `valet:"template,key=TestKey"`
			Three string `valet:"template,default=three"`
			Four  string `valet:"key=TestKey,default=four"`
			Five  string `valet:"key=TestKey"`
			Six   string `valet:"default=six"`
			Seven string `valet:"template"`
		}

		var (
			test *testStruct
		)

		BeforeEach(func() {
			test = &testStruct{}
		})
		Context("default", func() {

			It("will use the default value if no value exists", func() {
				err := values.RenderFields(test, runner)
				Expect(err).NotTo(HaveOccurred())
				Expect(test.One).To(Equal("one"))
				Expect(test.Three).To(Equal("three"))
				Expect(test.Four).To(Equal("four"))
				Expect(test.Six).To(Equal("six"))
			})

			It("will not use the default value if a value already exists", func() {
				replacement := "hello"
				test := &testStruct{
					One:   replacement,
					Three: replacement,
					Four:  replacement,
					Six:   replacement,
				}
				err := values.RenderFields(test, runner)
				Expect(err).NotTo(HaveOccurred())
				Expect(test.One).To(Equal(replacement))
				Expect(test.Three).To(Equal(replacement))
				Expect(test.Four).To(Equal(replacement))
				Expect(test.Six).To(Equal(replacement))
			})

		})

		Context("key", func() {
			BeforeEach(func() {
				values[testKey] = testValue
			})
			It("Will use the key if no default value exists unless template exists", func() {
				err := values.RenderFields(test, runner)
				Expect(err).NotTo(HaveOccurred())
				Expect(test.Five).To(Equal(testValue))
			})

			It("Will use the key and override the default value unless template exists", func() {
				err := values.RenderFields(test, runner)
				Expect(err).NotTo(HaveOccurred())
				Expect(test.Four).To(Equal(testValue))
			})
		})

		Context("template", func() {
			BeforeEach(func() {
				values[testKey] = testValue
				values[templateKey] = templateValue
			})
			It("will template if no value exists already", func() {
				test = &testStruct{
					Two:   testTemplate,
					Seven: testTemplate,
				}
				err := values.RenderFields(test, runner)
				Expect(err).NotTo(HaveOccurred())
				Expect(test.Two).To(Equal(templateValue))
				Expect(test.Seven).To(Equal(templateValue))
			})

			It("will template when a default value exists", func() {
				test = &testStruct{
					One:   testTemplate,
					Three: testTemplate,
					Seven: testTemplate,
				}
				err := values.RenderFields(test, runner)
				Expect(err).NotTo(HaveOccurred())
				Expect(test.One).To(Equal(templateValue))
				Expect(test.Three).To(Equal(templateValue))
				Expect(test.Seven).To(Equal(templateValue))
			})
		})
	})

	Context("get value", func() {

	})
})

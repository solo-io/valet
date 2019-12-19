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

	Context("nested struct", func() {
		type innerTestStruct struct {
			One string `valet:"template,key=TestKey,default=one"`
		}
		type testStructPtr struct {
			One string `valet:"template,key=TestKey,default=one"`
			Two *innerTestStruct
		}

		type testStruct struct {
			One string `valet:"template,key=TestKey,default=one"`
			Two innerTestStruct
		}
		var (
			testPtr *testStructPtr
			test *testStruct
		)

		It("will render child with pointer recursively", func() {
			testPtr = &testStructPtr{
				Two: &innerTestStruct{},
			}
			values[testKey] = testValue
			err := values.RenderFields(testPtr, runner)
			Expect(err).NotTo(HaveOccurred())
			Expect(testPtr.One).To(Equal(testValue))
			Expect(testPtr.Two.One).To(Equal(testValue))
		})

		It("will render child struct recursively", func() {
			test = &testStruct{
				Two: innerTestStruct{},
			}
			values[testKey] = testValue
			err := values.RenderFields(test, runner)
			Expect(err).NotTo(HaveOccurred())
			Expect(test.One).To(Equal(testValue))
			Expect(test.Two.One).To(Equal(testValue))
		})

		It("will not render nil child struct", func() {
			testPtr = &testStructPtr{
				Two: nil,
			}
			values[testKey] = testValue
			err := values.RenderFields(testPtr, runner)
			Expect(err).NotTo(HaveOccurred())
			Expect(testPtr.One).To(Equal(testValue))
			Expect(testPtr.Two).To(BeNil())
		})

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

		Context("key", func() {
			It("Will use the key if it can be found", func() {
				values[testKey] = testValue
				err := values.RenderFields(test, runner)
				Expect(err).NotTo(HaveOccurred())
				Expect(test.One).To(Equal(testValue))
				Expect(test.Two).To(Equal(testValue))
				Expect(test.Four).To(Equal(testValue))
				Expect(test.Five).To(Equal(testValue))
			})

			It("Will not use the key if it canot be found", func() {
				err := values.RenderFields(test, runner)
				Expect(err).NotTo(HaveOccurred())
				Expect(test.Two).To(Equal(""))
				Expect(test.Five).To(Equal(""))
			})
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

			It("will template when a default value exists, or key has been found", func() {
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
})

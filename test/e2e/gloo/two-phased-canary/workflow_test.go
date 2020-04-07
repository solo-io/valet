package two_phased_canary_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	two_phased_canary "github.com/solo-io/valet/test/e2e/gloo/two-phased-canary"
	"testing"
)

func TestTwoPhasedCanary(t *testing.T) {
	RegisterFailHandler(Fail)
	testutils.RegisterPreFailHandler(
		func() {
			testutils.PrintTrimmedStack()
		})
	testutils.RegisterCommonFailHandlers()
	RunSpecs(t, "Two Phased Canary Test Suite")
}

var _ = Describe("Two Phased Canary", func() {
	testWorkflow := two_phased_canary.GetTestWorkflow()

	BeforeSuite(func() {
		testWorkflow.Setup(".")
	})

	It("works", func() {
		testWorkflow.Run(".")
	})
})

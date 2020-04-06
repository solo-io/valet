package rate_limiting_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	rate_limiting "github.com/solo-io/valet/test/e2e/gloo/rate-limiting"
	"testing"
)

func TestRateLimit(t *testing.T) {
	RegisterFailHandler(Fail)
	testutils.RegisterPreFailHandler(
		func() {
			testutils.PrintTrimmedStack()
		})
	testutils.RegisterCommonFailHandlers()
	RunSpecs(t, "Rate Limit Suite")
}

var _ = Describe("Rate limit", func() {
	testWorkflow := rate_limiting.GetTestWorkflow()

	BeforeSuite(func() {
		testWorkflow.Setup(".")
	})

	It("works", func() {
		testWorkflow.Run(".")
	})
})

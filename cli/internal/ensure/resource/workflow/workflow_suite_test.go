package workflow_test

import (
	"github.com/solo-io/go-utils/testutils"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

func TestWorkflow(t *testing.T) {
	RegisterFailHandler(Fail)
	testutils.RegisterPreFailHandler(
		func() {
			testutils.PrintTrimmedStack()
		})
	testutils.RegisterCommonFailHandlers()
	T = t
	RunSpecs(t, "Workflow Suite")
}
package render_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

func TestRender(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "Render Suite")
}

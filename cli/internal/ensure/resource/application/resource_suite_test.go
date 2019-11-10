package application_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

func TestResources(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "Resource Suite")
}

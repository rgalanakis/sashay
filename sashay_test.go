package sashay_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rgalanakis/sashay"
)

func TestEchoSwagger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sashay Suite")
}

var _ = Describe("Sashay", func() {

	It("tests", func() {
		Expect(sashay.RunTest()).To(Equal(1))
	})
})

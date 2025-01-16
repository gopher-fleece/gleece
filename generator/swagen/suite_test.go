package swagen

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSwagenModule(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Swagen Module Suite")
}

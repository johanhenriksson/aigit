package aigit

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAigit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aigit Suite")
}

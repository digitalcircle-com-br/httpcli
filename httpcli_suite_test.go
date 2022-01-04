package httpcli_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHttpcli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Httpcli Suite")
}


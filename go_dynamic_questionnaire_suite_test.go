package go_dynamic_questionnaire_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGoDynamicQuestionnaire(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoDynamicQuestionnaire Suite")
}

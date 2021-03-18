package controllers

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestBannerLoggerLabels(t *testing.T) {
	g := NewGomegaWithT(t)

	var logger = &BannerLogger{
		currentStage: 2,
		parent: &BannerLogger{
			currentStage: 4,
		},
	}

	g.Expect(logger.label()).To(Equal("4.2"))
}

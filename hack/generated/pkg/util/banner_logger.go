package controllers

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type BannerLogger struct {
	currentStage int
	parent       *BannerLogger
}

// NewBannerLogger creates a new BannerLogger with an expected number of stages
func NewBannerLogger() *BannerLogger {
	return &BannerLogger{}
}

func (b *BannerLogger) Header(content string) {
	b.WriteBanner("=", fmt.Sprintf("%s %s", b.label(), content))
}

func (b *BannerLogger) Subheader(content string) {
	b.WriteBanner("-", fmt.Sprintf("%s %s", b.label(), content))
}

func (b *BannerLogger) WriteBanner(line string, content string) {
	lineLength := utf8.RuneCountInString(content)
	fmt.Println(strings.Repeat(line, lineLength))
	fmt.Println(content)
	fmt.Println(strings.Repeat(line, lineLength))
}

// label returns an indexed identifier for the current stage
func (b *BannerLogger) label() string {
	if b.parent == nil {
		return fmt.Sprint(b.currentStage)
	}

	return fmt.Sprintf("%s.%d", b.parent.label(), b.currentStage)
}

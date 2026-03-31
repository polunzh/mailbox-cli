package tui

import (
	"fmt"

	"github.com/zhenqiang/mailbox-cli/internal/model"
)

// DetailView renders the full content of a single message.
type DetailView struct {
	detail *model.MessageDetail
}

// NewDetailView creates a detail view for the given message.
func NewDetailView(detail *model.MessageDetail) *DetailView {
	return &DetailView{detail: detail}
}

// RenderContent returns the text content of the message for display.
func (dv *DetailView) RenderContent() string {
	if dv.detail == nil {
		return ""
	}
	d := dv.detail
	return fmt.Sprintf("%s\nFrom: %s\nSubject: %s\n\n%s",
		StyleDim.Render(d.ReceivedAt),
		StyleBold.Render(d.From),
		StyleBold.Render(d.Subject),
		d.TextBody,
	)
}

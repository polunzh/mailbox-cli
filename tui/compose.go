package tui

import (
	"os"

	"github.com/zhenqiang/mailbox-cli/internal/model"
)

// ComposeView manages composing or replying to a message.
type ComposeView struct {
	replyTo *model.MessageDetail
	to      []string
	subject string
	body    string
}

// NewComposeView creates a compose view. Pass orig != nil for a reply.
func NewComposeView(orig *model.MessageDetail) *ComposeView {
	cv := &ComposeView{}
	if orig != nil {
		cv.replyTo = orig
		cv.to = []string{orig.From}
		cv.subject = "Re: " + orig.Subject
	}
	return cv
}

// Draft returns the current draft constructed from the compose state.
func (cv *ComposeView) Draft() model.Draft {
	var inReplyTo *model.MessageLocator
	if cv.replyTo != nil {
		loc := cv.replyTo.Locator
		inReplyTo = &loc
	}
	return model.Draft{
		To:        cv.to,
		Subject:   cv.subject,
		Body:      cv.body,
		InReplyTo: inReplyTo,
	}
}

// ResolveEditor returns the editor to use: $EDITOR or "vi" as fallback.
func ResolveEditor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "vi"
}

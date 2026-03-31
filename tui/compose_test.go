package tui_test

import (
	"os"
	"testing"

	"github.com/zhenqiang/mailbox-cli/internal/model"
	"github.com/zhenqiang/mailbox-cli/tui"
)

func TestDetailViewContent(t *testing.T) {
	detail := &model.MessageDetail{
		Locator:  model.MessageLocator{ID: "1"},
		From:     "sender@example.com",
		Subject:  "Test Subject",
		TextBody: "Hello world",
	}
	dv := tui.NewDetailView(detail)
	content := dv.RenderContent()
	if content == "" {
		t.Fatal("expected non-empty detail view content")
	}
}

func TestComposeEditorFallback(t *testing.T) {
	os.Unsetenv("EDITOR")
	editor := tui.ResolveEditor()
	if editor != "vi" {
		t.Fatalf("expected vi fallback, got %q", editor)
	}
}

func TestComposeEditorOverride(t *testing.T) {
	os.Setenv("EDITOR", "nano")
	defer os.Unsetenv("EDITOR")
	editor := tui.ResolveEditor()
	if editor != "nano" {
		t.Fatalf("expected nano, got %q", editor)
	}
}

func TestComposeDraftInit(t *testing.T) {
	cv := tui.NewComposeView(nil)
	if cv == nil {
		t.Fatal("expected non-nil compose view")
	}
}

func TestComposeReplyDraftInit(t *testing.T) {
	orig := &model.MessageDetail{
		Locator: model.MessageLocator{ID: "1"},
		From:    "reply-to@example.com",
		Subject: "Original",
	}
	cv := tui.NewComposeView(orig)
	draft := cv.Draft()
	if len(draft.To) == 0 || draft.To[0] != "reply-to@example.com" {
		t.Fatalf("expected reply-to address, got %v", draft.To)
	}
	if draft.Subject != "Re: Original" {
		t.Fatalf("expected Re: subject, got %q", draft.Subject)
	}
}

package tui

import (
	"strings"
	"testing"

	"github.com/polunzh/mailbox-cli/internal/model"
)

func renderedLineCount(s string) int {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}

func TestListViewFitsTerminalHeight(t *testing.T) {
	app := NewApp([]model.Account{
		{ID: "gmail:a@example.com", Provider: "gmail", Email: "a@example.com"},
	}, nil)
	app.width = 120
	app.height = 24
	for i := 0; i < 40; i++ {
		app.messages = append(app.messages, model.Message{
			Locator:    model.MessageLocator{AccountID: "gmail:a@example.com", Provider: "gmail", ID: "1"},
			From:       "sender@example.com",
			Subject:    "Hello",
			ReceivedAt: "2026-04-02T10:00:00Z",
		})
	}
	app.applyFilter()

	view := app.View()
	if got := renderedLineCount(view.Content); got > app.height {
		t.Fatalf("list view rendered %d lines for terminal height %d", got, app.height)
	}
}

func TestDetailViewFitsTerminalHeight(t *testing.T) {
	app := NewApp([]model.Account{
		{ID: "gmail:a@example.com", Provider: "gmail", Email: "a@example.com"},
	}, nil)
	app.width = 80
	app.height = 24
	app.state = viewDetail
	app.selectedMsg = &model.MessageDetail{
		Locator:  model.MessageLocator{AccountID: "gmail:a@example.com", Provider: "gmail", ID: "1"},
		From:     "sender@example.com",
		Subject:  "Hello",
		TextBody: "line1\nline2\nline3",
	}

	view := app.View()
	if got := renderedLineCount(view.Content); got > app.height {
		t.Fatalf("detail view rendered %d lines for terminal height %d", got, app.height)
	}
}

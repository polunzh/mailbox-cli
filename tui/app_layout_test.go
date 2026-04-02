package tui

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/polunzh/mailbox-cli/internal/model"
	"github.com/polunzh/mailbox-cli/internal/provider"
)

type fakeMailProvider struct {
	detail    *model.MessageDetail
	detailErr error
	getCalls  int
}

func (f *fakeMailProvider) Authenticate() (*provider.AuthResult, error) { return nil, nil }
func (f *fakeMailProvider) ListMessages(opts model.ListOptions) ([]model.Message, error) {
	return nil, nil
}
func (f *fakeMailProvider) GetMessage(loc model.MessageLocator) (*model.MessageDetail, error) {
	f.getCalls++
	if f.detailErr != nil {
		return nil, f.detailErr
	}
	return f.detail, nil
}
func (f *fakeMailProvider) SendMessage(d model.Draft) (*model.MessageLocator, error) { return nil, nil }

func renderedLineCount(s string) int {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
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

func TestLongListShowsScrollbar(t *testing.T) {
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

	view := stripANSI(app.View().Content)
	if !strings.Contains(view, "█") {
		t.Fatal("expected long list view to render a scrollbar thumb")
	}
}

func TestLoadPreviewUsesCacheBeforeRefreshing(t *testing.T) {
	loc := model.MessageLocator{AccountID: "gmail:a@example.com", Provider: "gmail", ID: "1"}
	cached := &model.MessageDetail{Locator: loc, Subject: "cached", TextBody: "cached body"}
	fresh := &model.MessageDetail{Locator: loc, Subject: "fresh", TextBody: "fresh body"}
	p := &fakeMailProvider{detail: fresh}

	app := NewApp([]model.Account{{ID: "gmail:a@example.com", Provider: "gmail", Email: "a@example.com"}}, p)
	app.width = 120
	app.height = 24
	app.setCachedDetail(loc.ID, cached)

	cmd := app.loadPreview(loc)
	if app.selectedMsg != cached {
		t.Fatal("expected cached detail to be shown immediately")
	}
	if !app.previewLoading {
		t.Fatal("expected previewLoading while refresh is in progress")
	}
	if cmd == nil {
		t.Fatal("expected background refresh command when cache exists")
	}

	msg := cmd()
	updated, ok := msg.(detailLoaded)
	if !ok {
		t.Fatalf("expected detailLoaded, got %T", msg)
	}
	modelAfter, _ := app.Update(updated)
	next := modelAfter.(*App)
	if next.selectedMsg == nil || next.selectedMsg.Subject != "fresh" {
		t.Fatal("expected fresh detail to replace cached detail after refresh")
	}
	if p.getCalls != 1 {
		t.Fatalf("expected one refresh call, got %d", p.getCalls)
	}
}

func TestLoadPreviewKeepsCacheWhenRefreshFails(t *testing.T) {
	loc := model.MessageLocator{AccountID: "gmail:a@example.com", Provider: "gmail", ID: "1"}
	cached := &model.MessageDetail{Locator: loc, Subject: "cached", TextBody: "cached body"}
	p := &fakeMailProvider{detailErr: errors.New("network down")}

	app := NewApp([]model.Account{{ID: "gmail:a@example.com", Provider: "gmail", Email: "a@example.com"}}, p)
	app.width = 120
	app.height = 24
	app.setCachedDetail(loc.ID, cached)

	cmd := app.loadPreview(loc)
	msg := cmd()
	modelAfter, _ := app.Update(msg)
	next := modelAfter.(*App)
	if next.selectedMsg != cached {
		t.Fatal("expected cached detail to remain visible when refresh fails")
	}
	if next.statusMsg != "已经是最新的 Tip" {
		t.Fatalf("expected latest-tip status, got %q", next.statusMsg)
	}
}

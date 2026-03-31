package tui_test

import (
	"testing"

	"github.com/zhenqiang/mailbox-cli/internal/model"
	"github.com/zhenqiang/mailbox-cli/tui"
)

func TestListViewNavigateDown(t *testing.T) {
	msgs := []model.Message{
		{Locator: model.MessageLocator{ID: "1"}, Subject: "First"},
		{Locator: model.MessageLocator{ID: "2"}, Subject: "Second"},
	}
	lv := tui.NewListView(msgs)
	if lv.SelectedIndex() != 0 {
		t.Fatalf("expected initial index 0, got %d", lv.SelectedIndex())
	}
	lv.MoveDown()
	if lv.SelectedIndex() != 1 {
		t.Fatalf("expected index 1, got %d", lv.SelectedIndex())
	}
}

func TestListViewNavigateUp(t *testing.T) {
	msgs := []model.Message{
		{Locator: model.MessageLocator{ID: "1"}, Subject: "First"},
		{Locator: model.MessageLocator{ID: "2"}, Subject: "Second"},
	}
	lv := tui.NewListView(msgs)
	lv.MoveDown()
	lv.MoveUp()
	if lv.SelectedIndex() != 0 {
		t.Fatalf("expected index 0 after up, got %d", lv.SelectedIndex())
	}
}

func TestListViewBoundaryDown(t *testing.T) {
	msgs := []model.Message{
		{Locator: model.MessageLocator{ID: "1"}, Subject: "Only"},
	}
	lv := tui.NewListView(msgs)
	lv.MoveDown() // should not go past last
	if lv.SelectedIndex() != 0 {
		t.Fatalf("expected index to stay 0, got %d", lv.SelectedIndex())
	}
}

func TestListViewBoundaryUp(t *testing.T) {
	msgs := []model.Message{
		{Locator: model.MessageLocator{ID: "1"}, Subject: "Only"},
	}
	lv := tui.NewListView(msgs)
	lv.MoveUp() // should not go negative
	if lv.SelectedIndex() != 0 {
		t.Fatalf("expected index to stay 0, got %d", lv.SelectedIndex())
	}
}

func TestListViewSelectedMessage(t *testing.T) {
	msgs := []model.Message{
		{Locator: model.MessageLocator{ID: "1"}, Subject: "First"},
		{Locator: model.MessageLocator{ID: "2"}, Subject: "Second"},
	}
	lv := tui.NewListView(msgs)
	lv.MoveDown()
	sel := lv.Selected()
	if sel == nil {
		t.Fatal("expected selected message")
	}
	if sel.Locator.ID != "2" {
		t.Fatalf("unexpected selected ID: %q", sel.Locator.ID)
	}
}

func TestListViewUnreadFilter(t *testing.T) {
	msgs := []model.Message{
		{Locator: model.MessageLocator{ID: "1"}, Subject: "Read", Unread: false},
		{Locator: model.MessageLocator{ID: "2"}, Subject: "Unread", Unread: true},
	}
	lv := tui.NewListView(msgs)
	lv.SetUnreadFilter(true)
	filtered := lv.FilteredMessages()
	if len(filtered) != 1 {
		t.Fatalf("expected 1 unread message, got %d", len(filtered))
	}
	if filtered[0].Locator.ID != "2" {
		t.Fatalf("unexpected message after filter: %q", filtered[0].Locator.ID)
	}
}

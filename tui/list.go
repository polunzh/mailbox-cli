package tui

import "github.com/zhenqiang/mailbox-cli/internal/model"

// ListView manages message list state.
type ListView struct {
	messages    []model.Message
	cursor      int
	unreadOnly  bool
}

// NewListView creates a list view with the given messages.
func NewListView(messages []model.Message) *ListView {
	return &ListView{messages: messages}
}

// SelectedIndex returns the current cursor position within FilteredMessages.
func (lv *ListView) SelectedIndex() int {
	return lv.cursor
}

// MoveDown moves the cursor down one row (clamped to last item).
func (lv *ListView) MoveDown() {
	filtered := lv.FilteredMessages()
	if lv.cursor < len(filtered)-1 {
		lv.cursor++
	}
}

// MoveUp moves the cursor up one row (clamped to 0).
func (lv *ListView) MoveUp() {
	if lv.cursor > 0 {
		lv.cursor--
	}
}

// Selected returns the currently highlighted message, or nil if the list is empty.
func (lv *ListView) Selected() *model.Message {
	filtered := lv.FilteredMessages()
	if len(filtered) == 0 {
		return nil
	}
	if lv.cursor >= len(filtered) {
		return nil
	}
	m := filtered[lv.cursor]
	return &m
}

// IsUnreadFilter reports whether the unread-only filter is active.
func (lv *ListView) IsUnreadFilter() bool {
	return lv.unreadOnly
}

// SetUnreadFilter enables or disables the unread-only filter.
// Resets the cursor to 0 when the filter changes.
func (lv *ListView) SetUnreadFilter(unreadOnly bool) {
	if lv.unreadOnly != unreadOnly {
		lv.unreadOnly = unreadOnly
		lv.cursor = 0
	}
}

// FilteredMessages returns the visible messages based on the current filter.
func (lv *ListView) FilteredMessages() []model.Message {
	if !lv.unreadOnly {
		return lv.messages
	}
	var out []model.Message
	for _, m := range lv.messages {
		if m.Unread {
			out = append(out, m)
		}
	}
	return out
}

// Render returns a string representation of the list for the TUI.
func (lv *ListView) Render(width int) string {
	filtered := lv.FilteredMessages()
	if len(filtered) == 0 {
		return StyleDim.Render("No messages.") + "\n"
	}
	var lines []byte
	for i, m := range filtered {
		cursor := "  "
		if i == lv.cursor {
			cursor = "> "
		}
		unread := " "
		if m.Unread {
			unread = StyleTitle.Render("●")
		}
		from := truncate(m.From, 24)
		subject := m.Subject
		if subject == "" {
			subject = StyleDim.Render("(no subject)")
		}
		date := StyleDim.Render(truncate(m.ReceivedAt, 16))
		var line string
		if i == lv.cursor {
			line = StyleBold.Render(cursor+unread+" "+from) + "  " + subject + "  " + date + "\n"
		} else {
			line = StyleDim.Render(cursor) + unread + " " + from + "  " + StyleDim.Render(subject) + "  " + date + "\n"
		}
		lines = append(lines, []byte(line)...)
	}
	return string(lines)
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

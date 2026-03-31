package renderer_test

import (
	"strings"
	"testing"

	"github.com/zhenqiang/mailbox-cli/internal/renderer"
)

func TestHTMLToText(t *testing.T) {
	html := "<p>Hello <b>world</b></p>"
	got := renderer.HTMLToText(html)
	if !strings.Contains(got, "Hello") || !strings.Contains(got, "world") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestHTMLToTextStripsTagsOnly(t *testing.T) {
	html := "<a href='x'>link text</a>"
	got := renderer.HTMLToText(html)
	if strings.Contains(got, "<a") {
		t.Fatalf("HTML tags not stripped: %q", got)
	}
}

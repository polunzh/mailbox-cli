package release

import (
	"strings"
	"testing"
)

func TestRenderHomebrewFormula(t *testing.T) {
	formula, err := RenderHomebrewFormula(HomebrewFormulaData{
		Repo:            "polunzh/mailbox-cli",
		Version:         "v0.1.0",
		DarwinAMD64SHA:  "amd64sha",
		DarwinARM64SHA:  "arm64sha",
	})
	if err != nil {
		t.Fatalf("RenderHomebrewFormula returned error: %v", err)
	}

	assertContains(t, formula, "class Mailbox < Formula")
	assertContains(t, formula, "homepage \"https://github.com/polunzh/mailbox-cli\"")
	assertContains(t, formula, "version \"0.1.0\"")
	assertContains(t, formula, "on_macos do")
	assertContains(t, formula, "on_arm do")
	assertContains(t, formula, "on_intel do")
	assertContains(t, formula, "url \"https://github.com/polunzh/mailbox-cli/releases/download/v0.1.0/mailbox_0.1.0_darwin-arm64.tar.gz\"")
	assertContains(t, formula, "sha256 \"arm64sha\"")
	assertContains(t, formula, "url \"https://github.com/polunzh/mailbox-cli/releases/download/v0.1.0/mailbox_0.1.0_darwin-amd64.tar.gz\"")
	assertContains(t, formula, "sha256 \"amd64sha\"")
	assertContains(t, formula, "bin.install \"mailbox\"")
	assertContains(t, formula, "system \"#{bin}/mailbox\", \"--help\"")
}

func TestRenderHomebrewFormulaRequiresCompleteData(t *testing.T) {
	_, err := RenderHomebrewFormula(HomebrewFormulaData{
		Repo:    "polunzh/mailbox-cli",
		Version: "v0.1.0",
	})
	if err == nil {
		t.Fatal("RenderHomebrewFormula should reject incomplete data")
	}
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected formula to contain %q\nformula:\n%s", needle, haystack)
	}
}

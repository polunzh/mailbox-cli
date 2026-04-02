package cmd_test

import (
	"testing"

	"github.com/polunzh/mailbox-cli/cmd"
)

func TestRootCommandStructure(t *testing.T) {
	// Ensure root command is properly initialized
	if cmd.RootCmd == nil {
		t.Fatal("RootCmd should not be nil")
	}

	// Verify command name
	if cmd.RootCmd.Name() != "mailbox" {
		t.Errorf("expected command name 'mailbox', got %q", cmd.RootCmd.Name())
	}

	// Verify persistent flags exist
	if cmd.RootCmd.PersistentFlags().Lookup("account") == nil {
		t.Error("expected 'account' persistent flag")
	}
	if cmd.RootCmd.PersistentFlags().Lookup("json") == nil {
		t.Error("expected 'json' persistent flag")
	}
}

package cmd

import (
	"github.com/polunzh/mailbox-cli/tui"
	"github.com/spf13/cobra"
)

var (
	accountFlag string
	jsonFlag    bool
)

// RootCmd is the root command for the CLI
var RootCmd = &cobra.Command{
	Use:   "mailbox",
	Short: "A terminal email client",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountStore, err := loadAccountStore()
		if err != nil {
			return err
		}
		credStore, err := loadCredentialStore()
		if err != nil {
			return err
		}
		return tui.Launch(accountStore, credStore, newRegistry())
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&accountFlag, "account", "", "Account ID or email")
	RootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output JSON")
}

func Execute() error {
	return RootCmd.Execute()
}

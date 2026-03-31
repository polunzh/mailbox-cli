package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zhenqiang/mailbox-cli/tui"
)

var (
	accountFlag string
	jsonFlag    bool
)

var rootCmd = &cobra.Command{
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
	rootCmd.PersistentFlags().StringVar(&accountFlag, "account", "", "Account ID or email")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output JSON")
}

func Execute() error {
	return rootCmd.Execute()
}

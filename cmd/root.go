package cmd

import (
	"github.com/spf13/cobra"
)

var (
	accountFlag string
	jsonFlag    bool
)

var rootCmd = &cobra.Command{
	Use:   "mailbox",
	Short: "A terminal email client",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Placeholder: will launch TUI in Task 15
		return cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&accountFlag, "account", "", "Account ID or email")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output JSON")
}

func Execute() error {
	return rootCmd.Execute()
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var authUseCmd = &cobra.Command{
	Use:   "use",
	Short: "Set the default account",
	RunE: func(cmd *cobra.Command, args []string) error {
		if accountFlag == "" {
			return fmt.Errorf("--account is required")
		}
		accountStore, err := loadAccountStore()
		if err != nil {
			return err
		}
		acct, err := resolveAccount(accountStore, accountFlag)
		if err != nil {
			return err
		}
		if err := accountStore.SetDefault(acct.ID); err != nil {
			return fmt.Errorf("set default: %w", err)
		}
		if _, err := fmt.Fprintf(os.Stdout, "Default account set to %s\n", acct.ID); err != nil {
			return fmt.Errorf("write success output: %w", err)
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authUseCmd)
}

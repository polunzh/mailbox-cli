package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove an authenticated account",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountStore, err := loadAccountStore()
		if err != nil {
			return err
		}
		credStore, err := loadCredentialStore()
		if err != nil {
			return err
		}

		acct, err := resolveAccount(accountStore, accountFlag)
		if err != nil {
			return err
		}

		if err := credStore.Delete(acct.CredKey); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: could not remove credential: %v\n", err)
		}
		if err := accountStore.Remove(acct.ID); err != nil {
			return fmt.Errorf("remove account: %w", err)
		}
		if _, err := fmt.Fprintf(os.Stdout, "Logged out %s\n", acct.Email); err != nil {
			return fmt.Errorf("write success output: %w", err)
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLogoutCmd)
}

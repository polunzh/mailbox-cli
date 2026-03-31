package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authenticated accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountStore, err := loadAccountStore()
		if err != nil {
			return err
		}

		accts, err := accountStore.List()
		if err != nil {
			return err
		}
		defaultID := ""
		if def, err := accountStore.GetDefault(); err == nil {
			defaultID = def.ID
		}

		if jsonFlag {
			type accountJSON struct {
				ID          string `json:"id"`
				Provider    string `json:"provider"`
				Email       string `json:"email"`
				DisplayName string `json:"displayName"`
			}
			type payload struct {
				Accounts         []accountJSON `json:"accounts"`
				DefaultAccountID string        `json:"defaultAccountId"`
			}
			p := payload{DefaultAccountID: defaultID, Accounts: []accountJSON{}}
			for _, a := range accts {
				p.Accounts = append(p.Accounts, accountJSON{
					ID:          a.ID,
					Provider:    a.Provider,
					Email:       a.Email,
					DisplayName: a.DisplayName,
				})
			}
			return json.NewEncoder(os.Stdout).Encode(p)
		}

		if len(accts) == 0 {
			cmd.Println("No authenticated accounts.")
			return nil
		}
		for _, a := range accts {
			line := a.ID
			if a.ID == defaultID {
				line += " (default)"
			}
			cmd.Println(line)
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}

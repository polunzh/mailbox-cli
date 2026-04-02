package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhenqiang/mailbox-cli/internal/model"
)

var (
	loginProvider string
	loginEmail    string
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate a new account",
	RunE: func(cmd *cobra.Command, args []string) error {
		if loginProvider == "" {
			return fmt.Errorf("--provider is required (gmail or qq)")
		}

		accountStore, err := loadAccountStore()
		if err != nil {
			return err
		}
		credStore, err := loadCredentialStore()
		if err != nil {
			return err
		}
		reg := newRegistry()

		var acct model.Account
		var cred string

		switch loginProvider {
		case "gmail":
			// Gmail: run OAuth flow; email resolved from token.
			oauthSrv, err := startOAuthCallback()
			if err != nil {
				return err
			}
			authURL := gmailAuthURL(oauthSrv.Port())
			if _, err := fmt.Fprintf(os.Stderr, "Opening browser for Gmail OAuth...\n%s\n", authURL); err != nil {
				return fmt.Errorf("write auth prompt: %w", err)
			}
			if err := openBrowser(authURL); err != nil {
				if _, writeErr := fmt.Fprintln(os.Stderr, "Could not open browser automatically. Please visit the URL above."); writeErr != nil {
					return fmt.Errorf("write browser fallback: %w", writeErr)
				}
			}
			code, err := oauthSrv.WaitForCode(oauthTimeout)
			if err != nil {
				return fmt.Errorf("oauth: %w", err)
			}
			email, token, err := exchangeGmailCode(oauthSrv.Port(), code)
			if err != nil {
				return fmt.Errorf("gmail token exchange: %w", err)
			}
			acct = model.Account{
				ID:       model.MakeAccountID("gmail", email),
				Provider: "gmail",
				Email:    email,
				CredKey:  "mailbox-cli/" + email,
			}
			cred = token
		case "qq":
			if loginEmail == "" {
				return fmt.Errorf("--email is required for QQ provider")
			}
			if _, err := fmt.Fprint(os.Stderr, "Enter QQ app password (授权码): "); err != nil {
				return fmt.Errorf("write password prompt: %w", err)
			}
			password, err := readPassword()
			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}
			acct = model.Account{
				ID:       model.MakeAccountID("qq", loginEmail),
				Provider: "qq",
				Email:    loginEmail,
				CredKey:  "mailbox-cli/" + loginEmail,
			}
			cred = password
		default:
			return fmt.Errorf("unknown provider %q; supported: gmail, qq", loginProvider)
		}

		// Verify credential by authenticating.
		p, err := reg.Build(acct.Provider, acct, cred)
		if err != nil {
			return fmt.Errorf("build provider: %w", err)
		}
		if _, err := p.Authenticate(); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		if err := credStore.Set(acct.CredKey, cred); err != nil {
			return fmt.Errorf("store credential: %w", err)
		}
		if err := accountStore.Add(acct); err != nil {
			return fmt.Errorf("save account: %w", err)
		}

		// If this is the first account, set it as default automatically.
		if existing, _ := accountStore.GetDefault(); existing.ID == "" {
			_ = accountStore.SetDefault(acct.ID)
		}

		if _, err := fmt.Fprintf(os.Stdout, "Logged in as %s (%s)\n", acct.Email, acct.Provider); err != nil {
			return fmt.Errorf("write success output: %w", err)
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authLoginCmd.Flags().StringVar(&loginProvider, "provider", "", "Provider: gmail or qq")
	authLoginCmd.Flags().StringVar(&loginEmail, "email", "", "Email address (required for QQ)")
}

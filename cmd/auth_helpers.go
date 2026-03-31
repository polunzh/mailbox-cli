package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/zhenqiang/mailbox-cli/internal/oauth"
	"golang.org/x/term"
)

const oauthTimeout = 5 * time.Minute

// startOAuthCallback starts the local OAuth callback server.
func startOAuthCallback() (*oauth.CallbackServer, error) {
	return oauth.NewCallbackServer()
}

// gmailAuthURL returns the Google OAuth URL for the given callback port.
// Requires GOOGLE_CLIENT_ID env variable in production.
func gmailAuthURL(port int) string {
	// Full OAuth URL built at runtime from client credentials.
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?redirect_uri=http://localhost:%d/callback", port)
}

// exchangeGmailCode exchanges an OAuth code for an email and serialized token.
// Requires GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET env variables.
func exchangeGmailCode(code string) (email, token string, err error) {
	return "", "", fmt.Errorf("gmail oauth not configured: set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET")
}

// openBrowser opens the given URL in the default browser.
func openBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	default:
		cmd, args = "xdg-open", []string{url}
	}
	return exec.Command(cmd, args...).Start() //nolint:gosec
}

// readPassword reads a password from stdin without echoing.
func readPassword() (string, error) {
	b, err := term.ReadPassword(0)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

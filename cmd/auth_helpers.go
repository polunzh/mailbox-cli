package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/zhenqiang/mailbox-cli/internal/oauth"
	gooauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	goterm "golang.org/x/term"
)

const oauthTimeout = 5 * time.Minute

var gmailScopes = []string{
	"https://www.googleapis.com/auth/gmail.readonly",
	"https://www.googleapis.com/auth/gmail.send",
	"https://www.googleapis.com/auth/userinfo.email",
}

// startOAuthCallback starts the local OAuth callback server.
func startOAuthCallback() (*oauth.CallbackServer, error) {
	return oauth.NewCallbackServer()
}

// gmailOAuthConfig builds the OAuth2 config from environment variables.
func gmailOAuthConfig(redirectURL string) (*gooauth2.Config, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}
	return &gooauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       gmailScopes,
		Endpoint:     google.Endpoint,
	}, nil
}

// gmailAuthURL returns the Google OAuth consent URL for the given callback port.
func gmailAuthURL(port int) string {
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)
	cfg, err := gmailOAuthConfig(redirectURL)
	if err != nil {
		// Fall back to a partial URL that will show a clear error in the browser.
		return "https://accounts.google.com/o/oauth2/v2/auth?error=missing_client_credentials"
	}
	return cfg.AuthCodeURL("state", gooauth2.AccessTypeOffline)
}

// exchangeGmailCode exchanges an OAuth code for an email and serialized token JSON.
func exchangeGmailCode(port int, code string) (email, token string, err error) {
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)
	cfg, err := gmailOAuthConfig(redirectURL)
	if err != nil {
		return "", "", err
	}
	tok, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		return "", "", fmt.Errorf("token exchange: %w", err)
	}

	// Fetch email from userinfo endpoint.
	client := cfg.Client(context.Background(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", fmt.Errorf("fetch userinfo: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, _ := io.ReadAll(resp.Body)
	var info struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &info); err != nil || info.Email == "" {
		return "", "", fmt.Errorf("parse userinfo: %w", err)
	}

	tokenJSON, err := json.Marshal(tok)
	if err != nil {
		return "", "", fmt.Errorf("serialize token: %w", err)
	}
	return info.Email, string(tokenJSON), nil
}

// openBrowser opens the given URL in the default browser.
func openBrowser(rawURL string) error {
	// Validate URL before opening.
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{rawURL}
	default:
		cmd, args = "xdg-open", []string{rawURL}
	}
	return exec.Command(cmd, args...).Start() //nolint:gosec
}

// readPassword reads a password from stdin without echoing.
func readPassword() (string, error) {
	b, err := goterm.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println() // newline after hidden input
	return strings.TrimSpace(string(b)), nil
}

// refreshGmailToken uses a stored token JSON to get a fresh HTTP client.
func refreshGmailToken(tokenJSON string) (*http.Client, error) {
	var tok gooauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &tok); err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	cfg, err := gmailOAuthConfig("")
	if err != nil {
		return nil, err
	}
	return cfg.Client(context.Background(), &tok), nil
}

package provider

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/polunzh/mailbox-cli/internal/model"
)

// --- decodeBase64 tests ---

func TestDecodeBase64_Valid(t *testing.T) {
	// "Hello World" base64url encoded (no padding)
	input := "SGVsbG8gV29ybGQ"
	result := decodeBase64(input)
	if result != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", result)
	}
}

func TestDecodeBase64_WithPadding(t *testing.T) {
	// "Hello" base64url encoded
	input := "SGVsbG8"
	result := decodeBase64(input)
	if result != "Hello" {
		t.Errorf("expected 'Hello', got %q", result)
	}
}

func TestDecodeBase64_InvalidReturnsRaw(t *testing.T) {
	// Invalid base64 should return raw input
	input := "!!!invalid!!!"
	result := decodeBase64(input)
	if result != input {
		t.Errorf("expected raw input %q, got %q", input, result)
	}
}

func TestDecodeBase64_Empty(t *testing.T) {
	result := decodeBase64("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

// --- parseAddress tests ---

func TestParseAddress_WithDisplayName(t *testing.T) {
	input := "John Doe <john@example.com>"
	result := parseAddress(input)
	if result != "john@example.com" {
		t.Errorf("expected 'john@example.com', got %q", result)
	}
}

func TestParseAddress_JustEmail(t *testing.T) {
	input := "john@example.com"
	result := parseAddress(input)
	if result != "john@example.com" {
		t.Errorf("expected 'john@example.com', got %q", result)
	}
}

func TestParseAddress_WithWhitespace(t *testing.T) {
	input := "  john@example.com  "
	result := parseAddress(input)
	if result != "john@example.com" {
		t.Errorf("expected 'john@example.com', got %q", result)
	}
}

func TestParseAddress_Empty(t *testing.T) {
	result := parseAddress("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestParseAddress_MissingClosingBracket(t *testing.T) {
	input := "John Doe <john@example.com"
	result := parseAddress(input)
	// Should return trimmed input since no closing bracket
	if result != input {
		t.Errorf("expected %q, got %q", input, result)
	}
}

// --- parseAddressList tests ---

func TestParseAddressList_Single(t *testing.T) {
	input := "john@example.com"
	result := parseAddressList(input)
	if len(result) != 1 || result[0] != "john@example.com" {
		t.Errorf("expected ['john@example.com'], got %v", result)
	}
}

func TestParseAddressList_Multiple(t *testing.T) {
	input := "john@example.com, Jane Doe <jane@example.com>, bob@test.org"
	result := parseAddressList(input)
	expected := []string{"john@example.com", "jane@example.com", "bob@test.org"}
	if len(result) != 3 {
		t.Fatalf("expected 3 addresses, got %d: %v", len(result), result)
	}
	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("address %d: expected %q, got %q", i, exp, result[i])
		}
	}
}

func TestParseAddressList_Empty(t *testing.T) {
	result := parseAddressList("")
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestParseAddressList_WithWhitespace(t *testing.T) {
	input := "  john@example.com  ,  jane@example.com  "
	result := parseAddressList(input)
	if len(result) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(result))
	}
	if result[0] != "john@example.com" || result[1] != "jane@example.com" {
		t.Errorf("unexpected result: %v", result)
	}
}

// --- buildRFC822 tests ---

func TestBuildRFC822_Basic(t *testing.T) {
	draft := model.Draft{
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Hello World",
	}
	result := buildRFC822("sender@example.com", draft)

	expectedParts := []string{
		"From: sender@example.com",
		"To: recipient@example.com",
		"Subject: Test Subject",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"\r\n",
		"Hello World",
	}

	for _, part := range expectedParts {
		if !bytes.Contains([]byte(result), []byte(part)) {
			t.Errorf("expected result to contain %q, got:\n%s", part, result)
		}
	}
}

func TestBuildRFC822_MultipleRecipients(t *testing.T) {
	draft := model.Draft{
		To:      []string{"a@example.com", "b@example.com", "c@example.com"},
		Subject: "Multi",
		Body:    "Body text",
	}
	result := buildRFC822("sender@test.org", draft)

	if !bytes.Contains([]byte(result), []byte("To: a@example.com, b@example.com, c@example.com")) {
		t.Errorf("expected multiple recipients in To header, got:\n%s", result)
	}
}

func TestBuildRFC822_EmptyBody(t *testing.T) {
	draft := model.Draft{
		To:      []string{"r@example.com"},
		Subject: "Empty",
		Body:    "",
	}
	result := buildRFC822("s@example.com", draft)

	// Should still have proper headers ending with empty line
	if !bytes.Contains([]byte(result), []byte("\r\n\r\n")) {
		t.Error("expected CRLF CRLF between headers and body")
	}
}

// --- Constructor tests ---

func TestNewGmailProvider(t *testing.T) {
	acct := model.Account{ID: "gmail:test@example.com", Provider: "gmail", Email: "test@example.com"}
	p := NewGmailProvider(acct, "fake-token")
	if p == nil {
		t.Fatal("NewGmailProvider returned nil")
	}
	if p.account.ID != acct.ID {
		t.Errorf("expected account ID %q, got %q", acct.ID, p.account.ID)
	}
	if p.token != "fake-token" {
		t.Errorf("expected token 'fake-token', got %q", p.token)
	}
}

func TestNewGmailProviderWithHTTPClient(t *testing.T) {
	acct := model.Account{ID: "gmail:test@example.com", Provider: "gmail", Email: "test@example.com"}
	client := &http.Client{}
	p := NewGmailProviderWithHTTPClient(acct, client, "https://gmail.example.com")
	if p == nil {
		t.Fatal("NewGmailProviderWithHTTPClient returned nil")
	}
	if p.client != client {
		t.Error("expected client to be set")
	}
}

func TestNewGmailProviderWithBaseURL_TrailingSlash(t *testing.T) {
	acct := model.Account{ID: "gmail:test@example.com", Provider: "gmail", Email: "test@example.com"}
	p := NewGmailProviderWithBaseURL(acct, "token", "https://gmail.example.com/")
	// The URL should have trailing slash removed
	if p.baseURL != "https://gmail.example.com" {
		t.Errorf("expected baseURL without trailing slash, got %q", p.baseURL)
	}
}

func TestGmailProvider_Authenticate(t *testing.T) {
	acct := model.Account{ID: "gmail:test@example.com", Provider: "gmail", Email: "test@example.com"}
	p := NewGmailProvider(acct, "fake-token")

	result, err := p.Authenticate()
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	if result.Credential != "fake-token" {
		t.Errorf("expected credential 'fake-token', got %q", result.Credential)
	}
}

func TestNewQQProvider(t *testing.T) {
	acct := model.Account{ID: "qq:test@qq.com", Provider: "qq", Email: "test@qq.com"}
	p := NewQQProvider(acct, "password")
	if p == nil {
		t.Fatal("NewQQProvider returned nil")
	}
	if p.account.ID != acct.ID {
		t.Errorf("expected account ID %q, got %q", acct.ID, p.account.ID)
	}
}



package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/polunzh/mailbox-cli/internal/account"
)

// --- WriteJSONError tests ---

func TestWriteJSONError_Basic(t *testing.T) {
	var buf bytes.Buffer
	WriteJSONError(&buf, ErrCodeInvalidArguments, "test message")

	var result struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if result.Error.Code != "invalid_arguments" {
		t.Errorf("expected code=invalid_arguments, got %q", result.Error.Code)
	}
	if result.Error.Message != "test message" {
		t.Errorf("expected message='test message', got %q", result.Error.Message)
	}
}

// --- MapErrorCode tests ---

func TestMapErrorCode_NoDefaultAccount(t *testing.T) {
	code := MapErrorCode(ErrNoDefaultAccount)
	if code != ErrCodeNoDefaultAccount {
		t.Errorf("expected %q, got %q", ErrCodeNoDefaultAccount, code)
	}
}

func TestMapErrorCode_AccountNotFound(t *testing.T) {
	code := MapErrorCode(account.ErrNotFound)
	if code != ErrCodeAccountNotFound {
		t.Errorf("expected %q, got %q", ErrCodeAccountNotFound, code)
	}
}

func TestMapErrorCode_AmbiguousAccount(t *testing.T) {
	code := MapErrorCode(account.ErrAmbiguous)
	if code != ErrCodeAmbiguousAccount {
		t.Errorf("expected %q, got %q", ErrCodeAmbiguousAccount, code)
	}
}

func TestMapErrorCode_UnknownError(t *testing.T) {
	code := MapErrorCode(errors.New("random error"))
	if code != ErrCodeInvalidArguments {
		t.Errorf("expected %q for unknown error, got %q", ErrCodeInvalidArguments, code)
	}
}

// --- MapDefaultAccountError tests ---

func TestMapDefaultAccountError_NotFoundGivesNoDefault(t *testing.T) {
	code := MapDefaultAccountError(account.ErrNotFound)
	if code != ErrCodeNoDefaultAccount {
		t.Errorf("expected %q, got %q", ErrCodeNoDefaultAccount, code)
	}
}

func TestMapDefaultAccountError_DelegatesToMapErrorCode(t *testing.T) {
	code := MapDefaultAccountError(account.ErrAmbiguous)
	if code != ErrCodeAmbiguousAccount {
		t.Errorf("expected %q, got %q", ErrCodeAmbiguousAccount, code)
	}
}

// --- ErrorCode constants test ---

func TestAllErrorCodesAreDefined(t *testing.T) {
	codes := []struct {
		code  ErrorCode
		name  string
		value string
	}{
		{ErrCodeNoAuthenticatedAccounts, "ErrCodeNoAuthenticatedAccounts", "no_authenticated_accounts"},
		{ErrCodeNoDefaultAccount, "ErrCodeNoDefaultAccount", "no_default_account"},
		{ErrCodeAccountNotFound, "ErrCodeAccountNotFound", "account_not_found"},
		{ErrCodeAmbiguousAccount, "ErrCodeAmbiguousAccount", "ambiguous_account"},
		{ErrCodeMessageNotFound, "ErrCodeMessageNotFound", "message_not_found"},
		{ErrCodeAuthExpired, "ErrCodeAuthExpired", "auth_expired"},
		{ErrCodeNetworkError, "ErrCodeNetworkError", "network_error"},
		{ErrCodeInvalidArguments, "ErrCodeInvalidArguments", "invalid_arguments"},
	}

	for _, c := range codes {
		if c.code == "" {
			t.Errorf("%s should not be empty", c.name)
		}
		if string(c.code) != c.value {
			t.Errorf("%s expected=%q, got=%q", c.name, c.value, c.code)
		}
	}
}

// --- Sentinel errors test ---

func TestSentinelErrors(t *testing.T) {
	if ErrNoDefaultAccount == nil {
		t.Error("ErrNoDefaultAccount should not be nil")
	}
	if ErrAccountNotFound == nil {
		t.Error("ErrAccountNotFound should not be nil")
	}
	if ErrAmbiguousAccount == nil {
		t.Error("ErrAmbiguousAccount should not be nil")
	}
}

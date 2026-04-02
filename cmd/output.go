package cmd

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/polunzh/mailbox-cli/internal/account"
)

// ErrorCode is a machine-readable error code for --json mode.
type ErrorCode string

const (
	ErrCodeNoAuthenticatedAccounts ErrorCode = "no_authenticated_accounts"
	ErrCodeNoDefaultAccount        ErrorCode = "no_default_account"
	ErrCodeAccountNotFound         ErrorCode = "account_not_found"
	ErrCodeAmbiguousAccount        ErrorCode = "ambiguous_account"
	ErrCodeMessageNotFound         ErrorCode = "message_not_found"
	ErrCodeAuthExpired             ErrorCode = "auth_expired"
	ErrCodeNetworkError            ErrorCode = "network_error"
	ErrCodeInvalidArguments        ErrorCode = "invalid_arguments"
)

// Sentinel errors for mapping.
var (
	ErrNoDefaultAccount = errors.New("no default account configured")
	ErrAccountNotFound  = account.ErrNotFound
	ErrAmbiguousAccount = account.ErrAmbiguous
)

// errorPayload is the JSON structure for error output.
type errorPayload struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// WriteJSONError writes a JSON error payload to w.
func WriteJSONError(w io.Writer, code ErrorCode, message string) {
	var p errorPayload
	p.Error.Code = string(code)
	p.Error.Message = message
	_ = json.NewEncoder(w).Encode(p)
}

// MapErrorCode maps a Go error to the closest ErrorCode.
func MapErrorCode(err error) ErrorCode {
	if errors.Is(err, ErrNoDefaultAccount) {
		return ErrCodeNoDefaultAccount
	}
	if errors.Is(err, account.ErrAmbiguous) {
		return ErrCodeAmbiguousAccount
	}
	if errors.Is(err, account.ErrNotFound) {
		return ErrCodeAccountNotFound
	}
	return ErrCodeInvalidArguments
}

// MapDefaultAccountError returns ErrCodeNoDefaultAccount when the error arose
// from GetDefault, otherwise delegates to MapErrorCode.
func MapDefaultAccountError(err error) ErrorCode {
	if errors.Is(err, account.ErrNotFound) {
		return ErrCodeNoDefaultAccount
	}
	return MapErrorCode(err)
}

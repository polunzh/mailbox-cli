package cmd_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/polunzh/mailbox-cli/cmd"
)

func TestErrorPayloadShape(t *testing.T) {
	var buf bytes.Buffer
	cmd.WriteJSONError(&buf, cmd.ErrCodeInvalidArguments, "bad input")
	var out struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(&buf).Decode(&out); err != nil {
		t.Fatalf("decode error payload: %v", err)
	}
	if out.Error.Code != string(cmd.ErrCodeInvalidArguments) {
		t.Fatalf("unexpected code: %q", out.Error.Code)
	}
	if out.Error.Message != "bad input" {
		t.Fatalf("unexpected message: %q", out.Error.Message)
	}
}

func TestAllRequiredErrorCodesExist(t *testing.T) {
	required := []cmd.ErrorCode{
		cmd.ErrCodeNoAuthenticatedAccounts,
		cmd.ErrCodeNoDefaultAccount,
		cmd.ErrCodeAccountNotFound,
		cmd.ErrCodeAmbiguousAccount,
		cmd.ErrCodeMessageNotFound,
		cmd.ErrCodeAuthExpired,
		cmd.ErrCodeNetworkError,
		cmd.ErrCodeInvalidArguments,
	}
	for _, code := range required {
		if code == "" {
			t.Fatal("empty error code constant")
		}
	}
}

func TestErrorCodeMapping(t *testing.T) {
	// account store errors map to specific codes
	cases := []struct {
		err  error
		want cmd.ErrorCode
	}{
		{cmd.ErrNoDefaultAccount, cmd.ErrCodeNoDefaultAccount},
		{cmd.ErrAccountNotFound, cmd.ErrCodeAccountNotFound},
		{cmd.ErrAmbiguousAccount, cmd.ErrCodeAmbiguousAccount},
	}
	for _, c := range cases {
		got := cmd.MapErrorCode(c.err)
		if got != c.want {
			t.Fatalf("MapErrorCode(%v) = %q, want %q", c.err, got, c.want)
		}
	}
}

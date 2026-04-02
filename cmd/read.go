package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/polunzh/mailbox-cli/internal/model"
	"github.com/spf13/cobra"
)

// WriteReadJSON writes a { "message": ... } payload.
func WriteReadJSON(w io.Writer, detail *model.MessageDetail) error {
	type detailJSON struct {
		Locator           model.MessageLocator `json:"locator"`
		From              string               `json:"from"`
		To                []string             `json:"to"`
		CC                []string             `json:"cc"`
		Subject           string               `json:"subject"`
		ReceivedAt        string               `json:"receivedAt"`
		Unread            bool                 `json:"unread"`
		TextBody          string               `json:"textBody"`
		HTMLBodyAvailable bool                 `json:"htmlBodyAvailable"`
	}
	type payload struct {
		Message detailJSON `json:"message"`
	}
	p := payload{
		Message: detailJSON{
			Locator:           detail.Locator,
			From:              detail.From,
			To:                detail.To,
			CC:                detail.CC,
			Subject:           detail.Subject,
			ReceivedAt:        detail.ReceivedAt,
			Unread:            detail.Unread,
			TextBody:          detail.TextBody,
			HTMLBodyAvailable: detail.HtmlBodyAvailable,
		},
	}
	return json.NewEncoder(w).Encode(p)
}

// ValidateReadFlags checks incompatible flag combinations.
func validateReadFlags(jsonMode, browserFlag bool) error {
	if jsonMode && browserFlag {
		return errors.New("--browser is not allowed in --json mode")
	}
	return nil
}

var (
	readLocator string
	readBrowser bool
)

var readCmd = &cobra.Command{
	Use:   "read [id]",
	Short: "Read a message",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		positionalID := ""
		if len(args) > 0 {
			positionalID = args[0]
		}
		sel, err := ParseSelector(positionalID, readLocator)
		if err != nil {
			if jsonFlag {
				WriteJSONError(os.Stdout, ErrCodeInvalidArguments, err.Error())
				return nil
			}
			return err
		}
		if err := validateReadFlags(jsonFlag, readBrowser); err != nil {
			if jsonFlag {
				WriteJSONError(os.Stdout, ErrCodeInvalidArguments, err.Error())
				return nil
			}
			return err
		}

		accountStore, err := loadAccountStore()
		if err != nil {
			return err
		}
		acct, err := resolveAccount(accountStore, accountFlag)
		if err != nil {
			if jsonFlag {
				WriteJSONError(os.Stdout, MapErrorCode(err), err.Error())
				return nil
			}
			return err
		}
		loc, err := sel.ToLocator(acct)
		if err != nil {
			return err
		}
		credStore, err := loadCredentialStore()
		if err != nil {
			return err
		}
		p, err := buildProvider(newRegistry(), acct, credStore)
		if err != nil {
			return err
		}
		detail, err := p.GetMessage(loc)
		if err != nil {
			if jsonFlag {
				WriteJSONError(os.Stdout, ErrCodeMessageNotFound, err.Error())
				return nil
			}
			return err
		}
		if jsonFlag {
			return WriteReadJSON(os.Stdout, detail)
		}
		if _, err := fmt.Fprintf(os.Stdout, "From: %s\nSubject: %s\n\n%s\n", detail.From, detail.Subject, detail.TextBody); err != nil {
			return fmt.Errorf("write message output: %w", err)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(readCmd)
	readCmd.Flags().StringVar(&readLocator, "locator", "", "Message locator JSON")
	readCmd.Flags().BoolVar(&readBrowser, "browser", false, "Open HTML body in browser")
}

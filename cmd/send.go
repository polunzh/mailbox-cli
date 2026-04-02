package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/polunzh/mailbox-cli/internal/model"
	"github.com/spf13/cobra"
)

// WriteSendJSON writes { "sent": true, "locator": ..., "inReplyTo": ... }.
// inReplyTo is omitted when nil.
func WriteSendJSON(w io.Writer, loc *model.MessageLocator, inReplyTo *model.MessageLocator) error {
	type payload struct {
		Sent      bool                  `json:"sent"`
		Locator   model.MessageLocator  `json:"locator"`
		InReplyTo *model.MessageLocator `json:"inReplyTo,omitempty"`
	}
	return json.NewEncoder(w).Encode(payload{
		Sent:      true,
		Locator:   *loc,
		InReplyTo: inReplyTo,
	})
}

var (
	sendTo      []string
	sendSubject string
	sendBody    string
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a new message",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := ResolveBody(sendBody, "", jsonFlag)
		if err != nil {
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
		credStore, err := loadCredentialStore()
		if err != nil {
			return err
		}
		p, err := buildProvider(newRegistry(), acct, credStore)
		if err != nil {
			return err
		}

		draft := model.Draft{
			To:      sendTo,
			Subject: sendSubject,
			Body:    body,
		}
		loc, err := p.SendMessage(draft)
		if err != nil {
			if jsonFlag {
				WriteJSONError(os.Stdout, ErrCodeNetworkError, err.Error())
				return nil
			}
			return err
		}
		if jsonFlag {
			return WriteSendJSON(os.Stdout, loc, nil)
		}
		if _, err := fmt.Fprintf(os.Stdout, "Sent. Message ID: %s\n", loc.ID); err != nil {
			return fmt.Errorf("write send output: %w", err)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(sendCmd)
	sendCmd.Flags().StringArrayVar(&sendTo, "to", nil, "Recipient addresses")
	sendCmd.Flags().StringVar(&sendSubject, "subject", "", "Message subject")
	sendCmd.Flags().StringVar(&sendBody, "body", "", "Message body")
}

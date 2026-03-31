package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhenqiang/mailbox-cli/internal/model"
	"github.com/zhenqiang/mailbox-cli/internal/provider"
)

// ProviderAuthResult is re-exported for use in test files.
type ProviderAuthResult = provider.AuthResult

// ErrMessageNotFoundSentinel is returned when a message is not found.
var ErrMessageNotFoundSentinel = errors.New("message not found")

// WriteListJSON writes a { "messages": [...] } payload.
func WriteListJSON(w io.Writer, msgs []model.Message) error {
	type summaryJSON struct {
		Locator    model.MessageLocator `json:"locator"`
		From       string               `json:"from"`
		Subject    string               `json:"subject"`
		ReceivedAt string               `json:"receivedAt"`
		Unread     bool                 `json:"unread"`
	}
	type payload struct {
		Messages []summaryJSON `json:"messages"`
	}
	p := payload{Messages: []summaryJSON{}}
	for _, m := range msgs {
		p.Messages = append(p.Messages, summaryJSON{
			Locator:    m.Locator,
			From:       m.From,
			Subject:    m.Subject,
			ReceivedAt: m.ReceivedAt,
			Unread:     m.Unread,
		})
	}
	return json.NewEncoder(w).Encode(p)
}

var (
	listLimit  int
	listUnread bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List messages",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountStore, err := loadAccountStore()
		if err != nil {
			if jsonFlag {
				WriteJSONError(os.Stdout, ErrCodeNoDefaultAccount, err.Error())
				return nil
			}
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
		msgs, err := p.ListMessages(model.ListOptions{Limit: listLimit, Unread: listUnread})
		if err != nil {
			if jsonFlag {
				WriteJSONError(os.Stdout, ErrCodeNetworkError, err.Error())
				return nil
			}
			return err
		}
		if jsonFlag {
			return WriteListJSON(os.Stdout, msgs)
		}
		for _, m := range msgs {
			unread := " "
			if m.Unread {
				unread = "*"
			}
			fmt.Fprintf(os.Stdout, "%s [%s] %s — %s\n", unread, m.Locator.ID, m.From, m.Subject)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().IntVar(&listLimit, "limit", 20, "Maximum number of messages")
	listCmd.Flags().BoolVar(&listUnread, "unread", false, "Show only unread messages")
}

package cmd

import (
	"fmt"
	"os"

	"github.com/polunzh/mailbox-cli/internal/model"
	"github.com/spf13/cobra"
)

var (
	replyLocator string
	replyBody    string
	replySubject string
)

var replyCmd = &cobra.Command{
	Use:   "reply [id]",
	Short: "Reply to a message",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		positionalID := ""
		if len(args) > 0 {
			positionalID = args[0]
		}
		sel, err := ParseSelector(positionalID, replyLocator)
		if err != nil {
			if jsonFlag {
				WriteJSONError(os.Stdout, ErrCodeInvalidArguments, err.Error())
				return nil
			}
			return err
		}

		body, err := ResolveBody(replyBody, "", jsonFlag)
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
		origLoc, err := sel.ToLocator(acct)
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

		// Fetch original to get reply-to address.
		orig, err := p.GetMessage(origLoc)
		if err != nil {
			if jsonFlag {
				WriteJSONError(os.Stdout, ErrCodeMessageNotFound, err.Error())
				return nil
			}
			return err
		}

		subject := replySubject
		if subject == "" {
			subject = "Re: " + orig.Subject
		}
		draft := model.Draft{
			To:        []string{orig.From},
			Subject:   subject,
			Body:      body,
			InReplyTo: &origLoc,
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
			return WriteSendJSON(os.Stdout, loc, &origLoc)
		}
		if _, err := fmt.Fprintf(os.Stdout, "Reply sent. Message ID: %s\n", loc.ID); err != nil {
			return fmt.Errorf("write reply output: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(replyCmd)
	replyCmd.Flags().StringVar(&replyLocator, "locator", "", "Original message locator JSON")
	replyCmd.Flags().StringVar(&replyBody, "body", "", "Reply body")
	replyCmd.Flags().StringVar(&replySubject, "subject", "", "Override reply subject")
}

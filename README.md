[中文](./README.zh-CN.md) | **English**

# mailbox-cli

A terminal email client for developers. Supports an interactive TUI for daily use and CLI subcommands for scripting and automation.

## Features

- Full-screen TUI (Bubble Tea) with list, detail, and compose views
- CLI subcommands: `auth`, `list`, `read`, `send`, `reply`
- Multi-account support (Gmail + QQ Mail)
- `--json` mode for scripting with structured output and error codes
- Credentials stored in system keychain with file fallback (0600)

## Installation

```bash
go install github.com/zhenqiang/mailbox-cli@latest
```

Or build from source:

```bash
git clone https://github.com/zhenqiang/mailbox-cli
cd mailbox-cli
go build -o mailbox .
```

## Usage

### TUI

```bash
mailbox
```

### CLI

```bash
# Authenticate
mailbox auth login --provider gmail
mailbox auth login --provider qq --email you@qq.com
mailbox auth status
mailbox auth use --account gmail:you@gmail.com

# List messages
mailbox list
mailbox list --unread --limit 10

# Read a message
mailbox read <id>
mailbox read --locator '{"accountId":"gmail:you@gmail.com","provider":"gmail","id":"<id>"}'

# Send
mailbox send --to recipient@example.com --subject "Hello" --body "Hi there"

# Reply
mailbox reply <id> --body "Thanks"
```

### JSON mode

All commands support `--json` for machine-readable output:

```bash
mailbox --json list
mailbox --json read --locator '...'
mailbox --json send --to a@b.com --subject "Hi" --body "Hello"
```

Error shape: `{ "error": { "code": "...", "message": "..." } }`

## Supported Providers

| Provider | Auth | Protocol |
|----------|------|----------|
| Gmail | OAuth2 | Gmail API |
| QQ Mail | App password | IMAP / SMTP |

## Global Flags

| Flag | Description |
|------|-------------|
| `--account` | Account ID (`provider:email`) or email address |
| `--json` | Output JSON instead of human-readable text |

## License

MIT

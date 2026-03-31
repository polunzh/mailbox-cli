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

## Setup

### Gmail

1. Go to [Google Cloud Console](https://console.cloud.google.com/) and create a project.
2. Enable the **Gmail API** under APIs & Services.
3. Create an **OAuth 2.0 Client ID** (Application type: Desktop app).
4. Download the credentials and note the Client ID and Client Secret.
5. Set environment variables before running `auth login`:

```bash
export GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
export GOOGLE_CLIENT_SECRET=your-client-secret
mailbox auth login --provider gmail
```

To persist across sessions, add the exports to your shell profile (`~/.zshrc`, `~/.bashrc`, etc.).

### QQ Mail

Enable IMAP/SMTP access and generate an app password (授权码) in QQ Mail settings → Account → POP3/IMAP/SMTP. Then:

```bash
mailbox auth login --provider qq --email you@qq.com
# Enter the app password when prompted
```

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

[中文](./README.zh-CN.md) | **English**

# mailbox-cli

[![CI](https://github.com/polunzh/mailbox-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/polunzh/mailbox-cli/actions/workflows/ci.yml)
[![Codecov](./.github/badges/codecov.svg)](https://app.codecov.io/gh/polunzh/mailbox-cli)

A terminal email client for developers. Supports an interactive TUI for daily use and CLI subcommands for scripting and automation.

## Features

- **Modern TUI**: Split-pane layout (wide screens) or single-pane (narrow), vim-style keybindings
- **Responsive**: Automatically adapts to terminal size; list + detail side-by-side or full-screen
- **Persistent Toolbar**: Bottom shortcut bar showing context-aware keybindings
- **Visual Feedback**: Blue background highlight for selected row, unread messages marked with ●
- **Pagination**: Auto-load more messages when scrolling near bottom
- **CLI subcommands**: `auth`, `list`, `read`, `send`, `reply` for scripting
- **Multi-account**: Gmail (OAuth2) + QQ Mail (IMAP/SMTP)
- **JSON mode**: Structured output for automation with error codes
- **Secure**: Credentials stored in system keychain with file fallback (0600)

## Installation

On macOS, Homebrew is the recommended installation method:

```bash
brew install polunzh/tap/mailbox
```

You can also install the latest version with Go:

```bash
go install github.com/polunzh/mailbox-cli@latest
```

Or build from source:

```bash
git clone https://github.com/polunzh/mailbox-cli
cd mailbox-cli
go build -o mailbox .
```

Prebuilt archives for tagged versions are published on [GitHub Releases](https://github.com/polunzh/mailbox-cli/releases). The Homebrew formula is published to [`polunzh/tap`](https://github.com/polunzh/homebrew-tap).

To cut a release from your local checkout:

```bash
make release VERSION=v0.1.0
```

## Usage

### TUI

```bash
mailbox
```

**Keyboard Shortcuts:**

| Key | Action |
|-----|--------|
| `j/k` | Navigate down/up |
| `g/G` | First/last message |
| `Enter` | Open message |
| `h/←/Esc` | Go back (single-pane mode) |
| `n` | New message hint |
| `r` | Refresh list / Reply hint |
| `u` | Toggle unread filter |
| `?` | Toggle help panel |
| `q` | Quit |

**Features:**
- Split-pane on wide terminals (≥100 cols): list on left, detail on right
- Single-pane on narrow terminals: list or detail full-screen
- Persistent toolbar at bottom showing available shortcuts
- Blue background highlight for selected message
- Auto-pagination: scroll to bottom to load more messages
- Unread messages marked with ● and bold text

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

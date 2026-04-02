**中文** | [English](./README.md)

# mailbox-cli

[![CI](https://github.com/polunzh/mailbox-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/polunzh/mailbox-cli/actions/workflows/ci.yml)
[![Codecov](./.github/badges/codecov.svg)](https://app.codecov.io/gh/polunzh/mailbox-cli)

面向开发者的终端邮件客户端。支持全屏 TUI 日常使用，也支持 CLI 子命令用于脚本和自动化。

## 功能

- **现代化 TUI**：分栏布局（宽屏）或单栏（窄屏），Vim 风格快捷键
- **响应式设计**：根据终端尺寸自动适配，列表和详情并排或全屏
- **常驻工具栏**：底部快捷键提示栏，根据当前视图动态显示
- **视觉反馈**：选中行蓝色高亮，未读邮件用 ● 标记
- **分页加载**：滚动到底部附近自动加载更多邮件
- **CLI 子命令**：`auth`、`list`、`read`、`send`、`reply` 供脚本使用
- **多账号**：Gmail（OAuth2）+ QQ 邮箱（IMAP/SMTP）
- **JSON 模式**：结构化输出，适合自动化
- **安全**：凭证存储在系统钥匙串，fallback 到本地文件（权限 0600）

## 安装

在 macOS 上，推荐使用 Homebrew 安装：

```bash
brew install polunzh/tap/mailbox
```

也可以用 Go 安装最新版：

```bash
go install github.com/polunzh/mailbox-cli@latest
```

或从源码构建：

```bash
git clone https://github.com/polunzh/mailbox-cli
cd mailbox-cli
go build -o mailbox .
```

带版本号的预编译压缩包会发布在 [GitHub Releases](https://github.com/polunzh/mailbox-cli/releases)。Homebrew formula 会发布到 [`polunzh/tap`](https://github.com/polunzh/homebrew-tap)。

在本地发版可直接执行：

```bash
make release VERSION=v0.1.0
```

## 使用

### TUI

```bash
mailbox
```

**快捷键：**

| 按键 | 功能 |
|------|------|
| `j/k` | 上/下导航 |
| `g/G` | 第一封/最后一封邮件 |
| `Enter` | 打开邮件 |
| `h/←/Esc` | 返回（单栏模式） |
| `n` | 新邮件提示 |
| `r` | 刷新列表 / 回复提示 |
| `u` | 切换未读过滤 |
| `?` | 显示/隐藏帮助面板 |
| `q` | 退出 |

**特性：**
- 宽终端（≥100 列）分栏显示：左侧列表，右侧详情
- 窄终端单栏显示：列表或详情全屏
- 底部常驻工具栏显示当前可用快捷键
- 选中消息蓝色背景高亮
- 自动分页：滚动到底部加载更多邮件
- 未读邮件用 ● 标记并加粗显示

### CLI

```bash
# 认证
mailbox auth login --provider gmail
mailbox auth login --provider qq --email you@qq.com
mailbox auth status
mailbox auth use --account gmail:you@gmail.com

# 列出邮件
mailbox list
mailbox list --unread --limit 10

# 阅读邮件
mailbox read <id>
mailbox read --locator '{"accountId":"gmail:you@gmail.com","provider":"gmail","id":"<id>"}'

# 发送邮件
mailbox send --to recipient@example.com --subject "你好" --body "正文内容"

# 回复邮件
mailbox reply <id> --body "收到，谢谢"
```

### JSON 模式

所有命令支持 `--json`，输出机器可读格式：

```bash
mailbox --json list
mailbox --json read --locator '...'
mailbox --json send --to a@b.com --subject "Hi" --body "Hello"
```

错误格式：`{ "error": { "code": "...", "message": "..." } }`

## 配置

### Gmail

1. 前往 [Google Cloud Console](https://console.cloud.google.com/) 创建项目。
2. 在 APIs & Services 中启用 **Gmail API**。
3. 创建 **OAuth 2.0 客户端 ID**（应用类型选 Desktop app）。
4. 记录 Client ID 和 Client Secret。
5. 运行 `auth login` 前设置环境变量：

```bash
export GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
export GOOGLE_CLIENT_SECRET=your-client-secret
mailbox auth login --provider gmail
```

建议将 export 写入 `~/.zshrc` 或 `~/.bashrc` 以持久生效。

### QQ 邮箱

在 QQ 邮箱设置 → 账户 → POP3/IMAP/SMTP 中开启服务并生成授权码，然后：

```bash
mailbox auth login --provider qq --email you@qq.com
# 按提示输入授权码
```

## 支持的邮件提供商

| 提供商 | 认证方式 | 协议 |
|--------|----------|------|
| Gmail | OAuth2 | Gmail API |
| QQ 邮箱 | 授权码 | IMAP / SMTP |

## 全局参数

| 参数 | 说明 |
|------|------|
| `--account` | 账号 ID（`provider:email`）或邮箱地址 |
| `--json` | 输出 JSON 而非人类可读文本 |

## 许可证

MIT

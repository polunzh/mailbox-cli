**中文** | [English](./README.md)

# mailbox-cli

面向开发者的终端邮件客户端。支持全屏 TUI 日常使用，也支持 CLI 子命令用于脚本和自动化。

## 功能

- 全屏 TUI（Bubble Tea），包含列表、详情、写信视图
- CLI 子命令：`auth`、`list`、`read`、`send`、`reply`
- 多账号支持（Gmail + QQ 邮箱）
- `--json` 模式，输出结构化数据和错误码，适合脚本使用
- 凭证存储在系统钥匙串，fallback 到本地文件（权限 0600）

## 安装

```bash
go install github.com/zhenqiang/mailbox-cli@latest
```

或从源码构建：

```bash
git clone https://github.com/zhenqiang/mailbox-cli
cd mailbox-cli
go build -o mailbox .
```

## 使用

### TUI

```bash
mailbox
```

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

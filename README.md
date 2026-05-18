# HermesPage

轻量级 HTML 报告托管平台。将 Hermes agent 生成的 HTML 报告托管到 Web，提供统一的浏览入口和 MCP 集成。

## 功能

- **报告托管** — 上传 HTML 报告，通过 Web 直接访问
- **用户系统** — 多用户支持，管理员创建账号，个人 API Token
- **权限控制** — 报告支持 public/private，用户只看到自己的 + 公开的
- **自动提取** — 从 HTML 自动解析标题和标签
- **分类管理** — 按目录自动分类，标签/用户筛选
- **搜索** — 实时搜索标题和标签
- **MCP 集成** — 提供 MCP server，Hermes/Claude 可直接调用发布报告
- **一键部署** — Docker Compose，`restart: always` 开机自启

## 快速开始

### 本地开发

```bash
# 确保 Go 1.26+ 已安装
go version

# 启动 server（首次启动会进入 setup 引导）
HERMES_JWT_SECRET=dev-secret go run . serve

# 浏览器打开
open http://localhost:5487

# 或者预设管理员账号启动
HERMES_JWT_SECRET=dev-secret HERMES_ADMIN_USER=admin HERMES_ADMIN_PASS=admin go run . serve

# 灌入 mock 数据（含 admin/admin + demo/demo 两个账号）
go run scripts/mockdata.go
```

### Docker 部署（推荐）

```bash
# 配置环境变量
cp .env.example .env
# 编辑 .env 设置 HERMES_JWT_SECRET

# 一键启动（开机自启）
docker compose up -d

# 查看日志
docker compose logs -f
```

### Nginx 反向代理

参考 `deploy/nginx.conf`，将域名指向本机 5487 端口。

## 首次使用

1. 启动服务后打开浏览器
2. 如未设置 `HERMES_ADMIN_USER`/`HERMES_ADMIN_PASS`，会进入引导页创建管理员
3. 登录后进入设置页（点击右上角用户名 → 设置）查看 API Token 和 MCP 配置
4. 管理员可在设置页创建其他用户

## MCP 配置

在设置页可直接复制 MCP 配置。格式如下：

```json
{
  "mcpServers": {
    "hermespage": {
      "command": "/path/to/hermespage",
      "args": ["mcp"],
      "env": {
        "HERMES_SERVER_URL": "https://hermes.example.com",
        "HERMES_TOKEN": "tok_your_personal_token"
      }
    }
  }
}
```

### MCP Tools

| Tool | 说明 |
|------|------|
| `publish_report` | 发布 HTML 报告（传内容或文件路径，支持 visibility 参数） |
| `list_reports` | 列出/搜索已发布报告 |
| `delete_report` | 删除指定报告 |
| `get_report_info` | 查看单篇报告详情 |

## 自动提取

上传 HTML 时，server 会自动提取元信息。在 HTML 中添加以下 meta 标签即可：

```html
<head>
  <title>报告标题</title>
  <meta name="hermes-title" content="自定义标题（优先于 title 标签）">
  <meta name="hermes-tags" content="标签1,标签2,标签3">
</head>
```

## API

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/setup/status` | 检查是否需要初始设置 |
| POST | `/api/setup` | 首次创建管理员 |
| POST | `/api/auth/login` | 登录获取 JWT |

### 需要认证（JWT 或用户 Token）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/auth/me` | 当前用户信息 |
| POST | `/api/auth/reset-token` | 重置自己的 API Token |
| GET | `/api/list` | 列出报告（未登录只看 public） |
| POST | `/api/upload` | 上传报告 |
| DELETE | `/api/delete/{id}` | 删除报告 |
| PUT | `/api/report/{id}/visibility` | 切换报告公开/私有 |
| GET | `/api/report/{id}` | 报告详情 |

### 管理员接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/users` | 列出所有用户 |
| POST | `/api/users` | 创建用户 |
| DELETE | `/api/users/{id}` | 删除用户 |
| POST | `/api/users/{id}/reset-token` | 重置用户 Token |

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `HERMES_PORT` | 5487 | 服务端口 |
| `HERMES_DATA_DIR` | ./reports | 报告存储目录 |
| `HERMES_WEB_DIR` | ./web | 前端静态文件目录 |
| `HERMES_JWT_SECRET` | (随机) | JWT 签名密钥，建议生产环境固定 |
| `HERMES_ADMIN_USER` | - | 预设管理员用户名（可选） |
| `HERMES_ADMIN_PASS` | - | 预设管理员密码（可选） |

## 项目结构

```
├── main.go                  # 入口（serve/mcp 子命令）
├── internal/
│   ├── config/              # 配置
│   ├── auth/                # 用户存储、JWT、Token
│   ├── storage/             # 文件存储 + metadata
│   ├── handler/             # REST API + 中间件
│   └── mcpserver/           # MCP server
├── web/                     # 前端 SPA
│   ├── index.html           # 首页
│   ├── login.html           # 登录页
│   ├── setup.html           # 初始引导页
│   ├── settings.html        # 设置页（Token + MCP + 用户管理）
│   ├── app.js               # 前端逻辑
│   └── style.css            # 样式
├── scripts/mockdata.go      # Mock 数据生成脚本
├── deploy/nginx.conf        # Nginx 反向代理配置样例
├── Dockerfile               # 多阶段构建
└── docker-compose.yml       # 一键部署
```

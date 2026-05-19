# HermesPage 技术规格 (SPEC)

## 1. 系统架构

```
┌──────────────┐     HTTP API      ┌──────────────────┐     Static Files
│ Hermes Agent │ ──────────────────▶│   Go Server      │◀──────────────── Browser
│ (MCP Client) │                    │   :5487          │
└──────────────┘                    ├──────────────────┤
                                    │ /api/*    → API  │
       ┌──────────┐                 │ /reports/* → 报告 │
       │ MCP Srv  │─── HTTP ───────▶│ /*        → SPA  │
       │  (Go)    │                 └────────┬─────────┘
       └──────────┘                          │
                                             ▼
                                    ┌──────────────────┐
                                    │   File System    │
                                    │ reports/         │
                                    │   users.json     │
                                    │   metadata.json  │
                                    │   {category}/    │
                                    │     {file}.html  │
                                    └──────────────────┘
```

## 2. 认证体系

### 2.1 认证方式

| 场景 | 认证方式 |
|------|---------|
| 前端浏览 | JWT（登录后存 localStorage，请求头 Bearer） |
| MCP/API 上传 | 用户个人 Token（Bearer tok_xxx） |
| 管理操作 | JWT 或 Token（需 admin role） |

### 2.2 Token 识别逻辑

`Authorization: Bearer xxx` 中的 token：
- 以 `tok_` 开头 → 按用户 Token 查找对应用户
- 其他 → 视为 JWT，解析获取用户信息

### 2.3 JWT

- 有效期：7 天
- 密钥：`HERMES_JWT_SECRET` 环境变量（不设则随机生成，重启后旧 token 失效）
- Claims: `user_id`, `username`, `role`, `exp`

## 3. 数据模型

### 3.1 用户 `reports/users.json`

```json
{
  "users": [
    {
      "id": "u_abc12345",
      "username": "admin",
      "password_hash": "$2a$10$...",
      "role": "admin",
      "token": "tok_xxxxxxxxxxxxxxxx",
      "created_at": "2026-05-18T10:00:00Z"
    }
  ]
}
```

- `id`：`u_` + 8位随机字符串
- `password_hash`：bcrypt 哈希
- `role`：`admin` 或 `user`
- `token`：`tok_` + 32位随机字符串

### 3.2 报告 `reports/metadata.json`

```json
{
  "reports": [
    {
      "id": "a1b2c3d4",
      "filename": "analysis-2026-05-18.html",
      "title": "竞品分析报告",
      "category": "analysis",
      "tags": ["竞品", "周报"],
      "size": 45200,
      "created_at": "2026-05-18T10:30:00Z",
      "url": "/reports/analysis/analysis-2026-05-18.html",
      "owner": "u_abc12345",
      "visibility": "public"
    }
  ]
}
```

### 3.3 字段规则
- `id`：8 位随机十六进制字符串
- `owner`：上传者的用户 ID
- `visibility`：`public` | `private`
- `title`：优先级：上传参数 > `<meta name="hermes-title">` > `<title>` > 文件名
- `category`：字母数字 + 连字符，全小写，对应磁盘子目录

## 4. REST API

### 4.1 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/setup/status` | 返回 `{needs_setup: bool}` |
| POST | `/api/setup` | 首次创建管理员（仅 setup 模式可用） |
| POST | `/api/auth/login` | 登录，返回 JWT + 用户信息 |

### 4.2 认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/auth/me` | 当前用户信息（含 token） |
| POST | `/api/auth/reset-token` | 重置自己的 API Token |
| GET | `/api/list` | 列出报告（optionalAuth，未登录只返回 public） |
| POST | `/api/upload` | 上传报告（multipart/form-data） |
| DELETE | `/api/delete/{id}` | 删除报告（本人或 admin） |
| PUT | `/api/report/{id}/visibility` | 切换 public/private |
| GET | `/api/report/{id}` | 报告详情（optionalAuth） |

### 4.3 管理员接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/users` | 列出所有用户 |
| POST | `/api/users` | 创建用户 `{username, password, role}` |
| DELETE | `/api/users/{id}` | 删除用户 |
| POST | `/api/users/{id}/reset-token` | 重置用户 Token |

### 4.4 报告文件访问

`GET /reports/{category}/{filename}`：
- public 报告：任何人可访问
- private 报告：需认证（Header 或 URL `?token=` 参数）

### 4.5 上传接口详情

`POST /api/upload`（multipart/form-data）：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | HTML 文件（仅 .html/.htm） |
| title | string | 否 | 报告标题 |
| tags | string | 否 | 逗号分隔的标签 |
| category | string | 否 | 分类名（默认 uncategorized） |
| visibility | string | 否 | public/private（默认 private） |

### 4.6 列表接口响应

```json
{
  "reports": [...],
  "categories": ["analysis", "daily"],
  "total": 12,
  "owners": {"u_abc": "admin", "u_def": "demo"}
}
```

## 5. 前端页面

| 页面 | 路径 | 说明 |
|------|------|------|
| 首页 | `/` (index.html) | 报告列表，分类/标签/用户筛选 |
| 登录 | `/login.html` | 用户名 + 密码登录 |
| 引导 | `/setup.html` | 首次创建管理员 |
| 设置 | `/settings.html` | Token 管理、MCP 配置、用户管理（admin） |

### 5.1 首页功能
- 分组卡片视图（今天/昨天/本周/更早）
- 筛选栏：分类 + 标签 + 用户（admin only）
- 卡片显示：标题、分类 badge、标签、时间、大小、来源用户
- 可见性图标：点击切换 public/private（标题右侧）
- 用户下拉菜单（设置 + 退出登录）

### 5.2 设置页功能
- 我的 API Token（查看/复制/重新生成）
- MCP 配置（JSON 配置 + tools 说明）
- 用户管理（admin：创建/删除用户，重置 token）

## 6. MCP Server

Go 实现，使用 `github.com/mark3labs/mcp-go` SDK。支持两种传输模式：

### 6.1 Streamable HTTP（推荐）

作为 web server 的 `/mcp` 路由提供服务，MCP 客户端通过 HTTP 直接连接，无需本地安装二进制。

配置：
```json
{
  "mcpServers": {
    "hermespage": {
      "url": "https://page.example.com/mcp",
      "headers": {
        "Authorization": "Bearer tok_xxx"
      }
    }
  }
}
```

认证：从 HTTP Authorization header 提取用户 Token，MCP tools 以该用户身份操作。

### 6.2 stdio（本地模式）

通过 `hermespage mcp` 子命令启动，stdio 传输，供本地 MCP 客户端使用。

环境变量：
- `HERMES_SERVER_URL`：Go server 地址（默认 `http://localhost:5487`）
- `HERMES_TOKEN`：用户个人 API Token

### 6.3 Tools

| Tool | 参数 | 说明 |
|------|------|------|
| `publish_report` | html_content/file_path, title, tags, category, visibility | 发布报告 |
| `list_reports` | category, search | 列出/搜索报告 |
| `delete_report` | id | 删除报告 |
| `get_report_info` | id | 报告详情 |

## 7. 安全

- 所有写操作需认证（JWT 或用户 Token）
- 密码使用 bcrypt 哈希存储
- private 报告仅 owner 和 admin 可访问
- 上传文件仅接受 `.html`/`.htm`
- 文件名清理：去除路径穿越字符
- 单文件大小限制：10MB

## 8. 部署

### 8.1 Docker

```bash
docker compose up -d
```

### 8.2 Nginx 反向代理

参考 `deploy/nginx.conf`。

### 8.3 网络架构

```
Internet → Nginx (HTTPS :443) → Docker: hermespage (:5487)
                                 ├── /api/*     API
                                 ├── /mcp       MCP Streamable HTTP
                                 ├── /reports/* 报告文件（权限控制）
                                 └── /*         前端 SPA
```

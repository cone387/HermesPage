# HermesPage

## 项目概述

HermesPage 是一个轻量级 HTML 报告托管平台。Hermes agent 在本地生成 HTML 报告，HermesPage 负责托管这些报告并提供外网 Web 访问。

## 核心需求

1. **托管 HTML 报告**：将 Hermes 生成的 HTML 文件托管为可访问的网页
2. **统一入口首页**：分组卡片视图，按日期分组，支持搜索和分类筛选
3. **自动分类**：报告按目录自动归类，从 HTML 自动提取标题和标签
4. **MCP 集成**：提供 MCP server 供 Hermes 直接调用发布报告
5. **外网访问**：部署到公网服务器

## 技术栈

- **后端**：Go（标准库为主，单二进制部署）
- **前端**：纯 HTML + CSS + vanilla JS（无框架、无构建工具）
- **MCP**：Go（使用 mcp-go 社区 SDK，编译为同一二进制的子命令）
- **存储**：文件系统 + JSON 元数据（无数据库）

## 项目结构

```
HermesPage/
├── cmd/
│   ├── serve/       # 子命令：启动 web server
│   └── mcp/         # 子命令：启动 MCP server (stdio)
├── internal/
│   ├── config/      # 配置
│   ├── storage/     # 文件存储 + metadata 管理
│   ├── handler/     # HTTP API handlers
│   └── mcpserver/   # MCP server 实现
├── web/             # 前端 SPA（静态文件）
├── reports/         # 报告数据目录（不进 git）
├── main.go          # 入口，子命令分发
├── go.mod
├── CLAUDE.md        # 本文件：项目上下文和规则
├── SPEC.md          # 详细技术规格
├── Makefile         # 构建和运行命令
└── .gitignore
```

## 开发规则

### Go 代码
- 使用标准库，不引入 web 框架（net/http 足够）
- 代码风格遵循 `gofmt`
- 错误处理：返回有意义的 HTTP 状态码和 JSON 错误信息
- 配置通过环境变量，不用配置文件

### 前端代码
- 纯 vanilla JS，不用任何框架或构建工具
- CSS 使用变量便于主题切换
- 保持单文件可读（不拆太碎）

### MCP 代码
- 使用 `mcp-go` 社区 SDK（github.com/mark3labs/mcp-go）
- 通过 HTTP 调用本地 Go server API，不直接操作文件系统
- stdio 传输，供 Claude/Hermes 作为 MCP server 调用

### 通用规则
- 不使用数据库，元数据存 JSON 文件
- API 认证使用 Bearer Token（简单 API Key）
- 报告文件一旦上传不做修改，只能删除
- 分类（category）= reports/ 下的子目录名，仅一层
- ID 使用 8 位随机字符串

## 构建和运行

```bash
# 开发 - 启动 web server
go run . serve

# 开发 - 启动 MCP server (stdio)
go run . mcp

# 构建单二进制
go build -o hermespage .

# 使用
hermespage serve          # 启动 web server
hermespage mcp            # 启动 MCP server (stdio, 供 Hermes 调用)

# 部署
scp hermespage user@server:/opt/hermespage/
scp -r web/ user@server:/opt/hermespage/
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| HERMES_PORT | 5487 | 服务端口 |
| HERMES_DATA_DIR | ./reports | 报告存储目录 |
| HERMES_WEB_DIR | ./web | 前端静态文件目录 |
| HERMES_JWT_SECRET | (随机) | JWT 签名密钥 |
| HERMES_ADMIN_USER | - | 预设管理员用户名（可选） |
| HERMES_ADMIN_PASS | - | 预设管理员密码（可选） |

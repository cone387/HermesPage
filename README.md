# HermesPage

轻量级 HTML 报告托管平台。将 Hermes agent 生成的 HTML 报告托管到 Web，提供统一的浏览入口和 MCP 集成。

## 功能

- **报告托管** — 上传 HTML 报告，通过 Web 直接访问
- **自动提取** — 从 HTML 自动解析标题和标签
- **分类管理** — 按目录自动分类，标签按钮筛选
- **搜索** — 实时搜索标题和标签
- **MCP 集成** — 提供 MCP server，Hermes/Claude 可直接调用发布报告
- **一键部署** — Docker Compose，`restart: always` 开机自启

## 快速开始

### 本地开发

```bash
# 确保 Go 1.22+ 已安装
go version

# 启动 server
HERMES_API_KEY=dev-key go run . serve

# 浏览器打开
open http://localhost:8080

# 上传测试报告
curl -X POST -H "Authorization: Bearer dev-key" \
  -F "file=@your-report.html" \
  -F "category=daily" \
  http://localhost:8080/api/upload

# 灌入 mock 数据预览效果
HERMES_API_KEY=dev-key bash scripts/mock-data.sh
```

### Docker 部署（推荐）

```bash
# 设置 API Key
export HERMES_API_KEY=your-secret-key

# 一键启动（开机自启）
docker compose up -d

# 查看日志
docker compose logs -f
```

## MCP 配置

将以下配置添加到 Hermes/Claude 的 MCP settings：

```json
{
  "mcpServers": {
    "hermespage": {
      "command": "/path/to/hermespage",
      "args": ["mcp"],
      "env": {
        "HERMES_SERVER_URL": "http://your-server:8080",
        "HERMES_API_KEY": "your-secret-key"
      }
    }
  }
}
```

### MCP Tools

| Tool | 说明 |
|------|------|
| `publish_report` | 发布 HTML 报告（传内容或文件路径） |
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

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/list` | 列出报告 | 无 |
| POST | `/api/upload` | 上传报告 | Bearer Token |
| DELETE | `/api/delete/{id}` | 删除报告 | Bearer Token |
| GET | `/api/report/{id}` | 报告详情 | 无 |
| GET | `/reports/{category}/{file}` | 直接访问报告 | 无 |

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `HERMES_PORT` | 8080 | 服务端口 |
| `HERMES_DATA_DIR` | ./reports | 报告存储目录 |
| `HERMES_API_KEY` | (必填) | 上传/删除认证 token |
| `HERMES_WEB_DIR` | ./web | 前端静态文件目录 |

## 项目结构

```
├── main.go                  # 入口（serve/mcp 子命令）
├── internal/
│   ├── config/              # 配置
│   ├── storage/             # 文件存储 + metadata
│   ├── handler/             # REST API
│   └── mcpserver/           # MCP server
├── web/                     # 前端 SPA
├── scripts/mock-data.sh     # Mock 数据脚本
├── Dockerfile               # 多阶段构建
└── docker-compose.yml       # 一键部署
```

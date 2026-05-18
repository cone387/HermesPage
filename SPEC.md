# HermesPage 技术规格 (SPEC)

## 1. 系统架构

```
┌──────────────┐     HTTP API      ┌──────────────────┐     Static Files
│ Hermes Agent │ ──────────────────▶│   Go Server      │◀──────────────── Browser
│ (MCP Client) │                    │   :8080          │
└──────────────┘                    ├──────────────────┤
                                    │ /api/*    → API  │
       ┌──────────┐                 │ /reports/* → 报告 │
       │ MCP Srv  │─── HTTP ───────▶│ /*        → SPA  │
       │ (Python) │                 └────────┬─────────┘
       └──────────┘                          │
                                             ▼
                                    ┌──────────────────┐
                                    │   File System    │
                                    │ reports/         │
                                    │   metadata.json  │
                                    │   {category}/    │
                                    │     {file}.html  │
                                    └──────────────────┘
```

## 2. REST API 规格

### 2.1 列出报告 `GET /api/list`

**无需认证**

Query 参数（均可选）：
- `category` - 按分类筛选
- `tag` - 按标签筛选（可多次出现）
- `search` - 搜索标题

响应 `200 OK`：
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
      "url": "/reports/analysis/analysis-2026-05-18.html"
    }
  ],
  "categories": ["analysis", "daily", "monitoring"],
  "total": 42
}
```

### 2.2 上传报告 `POST /api/upload`

**需要认证**：`Authorization: Bearer {API_KEY}`

Content-Type: `multipart/form-data`

表单字段：
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | HTML 文件 |
| title | string | 否 | 报告标题（不传则自动提取） |
| tags | string | 否 | 逗号分隔的标签列表 |
| category | string | 否 | 分类名（默认 "uncategorized"） |

响应 `201 Created`：
```json
{
  "id": "a1b2c3d4",
  "url": "/reports/analysis/analysis-2026-05-18.html",
  "title": "竞品分析报告",
  "category": "analysis",
  "tags": ["竞品", "周报"],
  "size": 45200,
  "created_at": "2026-05-18T10:30:00Z"
}
```

错误响应：
- `401 Unauthorized` - API Key 无效或缺失
- `400 Bad Request` - 文件缺失或非 HTML

### 2.3 删除报告 `DELETE /api/delete/{id}`

**需要认证**：`Authorization: Bearer {API_KEY}`

响应 `200 OK`：
```json
{
  "message": "deleted",
  "id": "a1b2c3d4"
}
```

错误响应：
- `401 Unauthorized`
- `404 Not Found` - ID 不存在

## 3. 元数据模型

`reports/metadata.json` 结构：
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
      "created_at": "2026-05-18T10:30:00Z"
    }
  ]
}
```

### 字段规则
- `id`：8 位随机字母数字字符串，上传时生成
- `filename`：原始上传文件名（如有冲突则追加 ID 后缀）
- `title`：优先级：上传参数 > `<meta name="hermes-title">` > `<title>` > 文件名去扩展名
- `category`：字母数字 + 连字符，全小写，对应磁盘上的子目录
- `tags`：字符串数组。来源：上传参数 > `<meta name="hermes-tags" content="tag1,tag2">`
- `size`：文件字节数
- `created_at`：ISO 8601 UTC 时间

## 4. 自动提取逻辑

上传 HTML 时，server 解析文件提取元信息：

```
1. 读取 HTML 内容
2. 正则匹配 <meta name="hermes-title" content="...">  → title
3. 正则匹配 <meta name="hermes-tags" content="...">   → tags (逗号分割)
4. 正则匹配 <title>...</title>                         → fallback title
5. 如果都没有 → 用文件名（去扩展名，下划线/连字符换空格）
```

## 5. 前端 SPA 规格

### 5.1 页面布局

```
┌─────────────────────────────────────────────────────┐
│ Header                                              │
│   Logo: "HermesPage"                                │
│   搜索框 (实时过滤)                                   │
│   分类下拉 (All / daily / analysis / ...)            │
└─────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────┐
│ Content                                             │
│                                                     │
│ ─── 今天 (2026-05-18) ──────────────────────────── │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐            │
│ │ Title... │ │ Title... │ │ Title... │            │
│ │ [daily]  │ │[analysis]│ │ [daily]  │            │
│ │ #tag1    │ │ #tag2    │ │ #tag3    │            │
│ │ 10:30    │ │ 09:15    │ │ 08:00    │            │
│ │ 45KB     │ │ 32KB     │ │ 28KB     │            │
│ └──────────┘ └──────────┘ └──────────┘            │
│                                                     │
│ ─── 昨天 (2026-05-17) ──────────────────────────── │
│ ┌──────────┐ ┌──────────┐                          │
│ │ ...      │ │ ...      │                          │
│ └──────────┘ └──────────┘                          │
│                                                     │
│ ─── 更早 ───────────────────────────────────────── │
│ ...                                                 │
└─────────────────────────────────────────────────────┘
```

### 5.2 卡片内容
- 标题（截断到两行）
- 分类 badge（带颜色）
- 标签（小圆角标签）
- 时间（今天显示 HH:MM，其他显示日期）
- 文件大小（KB/MB）

### 5.3 交互
- 点击卡片 → `window.open(report.url, '_blank')`
- 搜索框 → 输入时实时过滤（debounce 300ms）
- 分类下拉 → 选择后立即过滤
- 空状态 → 显示 "暂无报告" 提示

### 5.4 分组逻辑（本地时间）
- 今天: created_at 是今天
- 昨天: created_at 是昨天
- 本周: created_at 在本周内（不含今天和昨天）
- 更早: 其他

## 6. MCP Server 规格

Go 实现，使用 `github.com/mark3labs/mcp-go` SDK。通过 stdio 传输与 Hermes/Claude 通信。
MCP server 通过 HTTP 调用本地 Go server 的 REST API 完成操作。

### 6.1 Tools 总览

| Tool | 说明 | 典型使用场景 |
|------|------|-------------|
| `publish_report` | 发布 HTML 报告 | Hermes 生成报告后直接发布 |
| `list_reports` | 列出已发布报告 | 查看有哪些报告、搜索特定报告 |
| `delete_report` | 删除报告 | 清理过期或错误的报告 |
| `get_report_info` | 获取单个报告详情 | 查看某篇报告的元信息和访问链接 |

### 6.2 Tool: publish_report

发布 HTML 报告到 HermesPage，返回访问 URL。

参数：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| html_content | string | 二选一 | HTML 内容字符串 |
| file_path | string | 二选一 | 本地 HTML 文件路径 |
| title | string | 否 | 报告标题（不传则自动提取） |
| tags | string | 否 | 逗号分隔的标签 |
| category | string | 否 | 分类名（默认 uncategorized） |

返回示例：
```
已发布: "竞品分析报告"
URL: http://your-server:8080/reports/analysis/report.html
ID: a1b2c3d4
分类: analysis | 标签: 竞品, 周报
```

### 6.3 Tool: list_reports

列出所有已发布的报告，支持筛选。

参数：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| category | string | 否 | 按分类筛选 |
| search | string | 否 | 搜索标题关键词 |

返回示例：
```
共 3 篇报告:

1. [a1b2c3d4] 竞品分析报告
   分类: analysis | 标签: 竞品, 周报 | 2026-05-18 10:30
   URL: /reports/analysis/report.html

2. [e5f6g7h8] 日报 05-17
   分类: daily | 标签: 日报 | 2026-05-17 18:00
   URL: /reports/daily/daily-0517.html
```

### 6.4 Tool: delete_report

删除指定报告（文件 + 元数据）。

参数：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 报告 ID |

返回示例：
```
已删除: "竞品分析报告" (a1b2c3d4)
```

### 6.5 Tool: get_report_info

获取单个报告的详细信息。

参数：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 报告 ID |

返回示例：
```
标题: 竞品分析报告
ID: a1b2c3d4
分类: analysis
标签: 竞品, 周报
大小: 45.2 KB
创建时间: 2026-05-18 10:30:00
URL: http://your-server:8080/reports/analysis/report.html
```

### 6.6 MCP 配置（环境变量）
- `HERMES_SERVER_URL`：Go server 地址（默认 `http://localhost:8080`）
- `HERMES_API_KEY`：认证 token

### 6.7 二进制使用方式

```bash
hermespage serve    # 启动 web server
hermespage mcp      # 启动 MCP server (stdio)
```

Hermes/Claude 的 MCP 配置示例：
```json
{
  "mcpServers": {
    "hermespage": {
      "command": "/path/to/hermespage",
      "args": ["mcp"],
      "env": {
        "HERMES_SERVER_URL": "http://localhost:8080",
        "HERMES_API_KEY": "your-api-key"
      }
    }
  }
}
```

## 7. 安全

- 上传/删除接口需 Bearer Token 认证
- 报告浏览和列表接口无需认证（公开访问）
- 上传文件仅接受 `.html`/`.htm` 扩展名
- 文件名清理：去除路径穿越字符（`..`、`/`、`\`）
- 单文件上传大小限制：10MB

## 8. 部署架构

### 8.1 Docker 一键部署

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o hermespage .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/hermespage .
COPY web/ ./web/
VOLUME /app/reports
EXPOSE 8080
CMD ["./hermespage", "serve"]
```

docker-compose.yml：
```yaml
services:
  hermespage:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./reports:/app/reports
    environment:
      - HERMES_API_KEY=your-secret-key
    restart: always   # 开机自启动
```

部署命令：
```bash
docker compose up -d   # 一键启动，自动拉起
```

### 8.2 网络架构

```
Internet → Nginx/Caddy (HTTPS, :443) → Docker: hermespage (:8080)
                                        ├── /api/*     API
                                        ├── /reports/* 报告文件
                                        └── /*         前端 SPA
```

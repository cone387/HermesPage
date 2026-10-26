# 设计：HermesPage 改名 Acta

日期：2026-10-26
状态：已批准

## 背景与目标

"HermesPage" 把平台绑死在 Hermes agent 上，平台实际可托管任意来源的 HTML 报告。新名 **Acta**（拉丁语"纪事/公报"，源自古罗马每日公告 Acta Diurna；学刊传统如 Acta Mathematica）——短、独特、含义精准，经检索无强撞名。

目标：彻底改名（品牌、代码标识符、环境变量、构建、文档、仓库），不做运行时兼容。

## 标识符映射

| 旧 | 新 |
|----|----|
| 品牌 HermesPage | Acta |
| Go 模块 `github.com/hermespage/hermespage` | `github.com/cone387/acta`（顺带修正模块路径与仓库不一致的旧债） |
| 二进制 `hermespage` | `acta` |
| `HERMES_PORT` / `HERMES_DATA_DIR` / `HERMES_WEB_DIR` / `HERMES_JWT_SECRET` / `HERMES_ADMIN_USER` / `HERMES_ADMIN_PASS` / `HERMES_SERVER_URL` / `HERMES_TOKEN` | `ACTA_PORT` / `ACTA_DATA_DIR` / `ACTA_WEB_DIR` / `ACTA_JWT_SECRET` / `ACTA_ADMIN_USER` / `ACTA_ADMIN_PASS` / `ACTA_SERVER_URL` / `ACTA_TOKEN` |
| MCP server 名 `HermesPage`（3 处）及 settings.html MCP 配置键 `hermespage` | `Acta` / `acta` |
| HTML meta 前缀 `hermes-title` / `hermes-tags`（storage.go 正则） | `acta-title` / `acta-tags`，干净切换，不双读 |
| nginx 示例域名 `hermes.example.com` | `acta.example.com` |
| GitHub 仓库 `cone387/HermesPage` | `cone387/acta`（执行前向用户最终确认；GitHub 自动重定向旧 URL） |

**保持不变**（数据/接口层，改动会破坏存量与部署）：默认端口 5487、数据目录 `./reports`、API 路径 `/api/*`、`/mcp`、`/reports/*`、token 前缀 `tok_`、metadata.json 格式、报告文件不可变规则。

## 存量数据决策（用户拍板）

干净切换，不做兼容。存量报告**无需迁移**：标题/标签在上传时已提取进 metadata.json，HTML 内的 `hermes-*` meta 此后不再被读取（提取仅发生在上传时）。一次性外部更新项：

1. 正式服务器 `.env`：`HERMES_*` 全部改为 `ACTA_*`
2. Hermes agent 侧生成模板/提示词：meta 前缀改 `acta-*`（只影响新上传报告的自动提取）

## 改动范围（约 20 个文件）

- **代码**：go.mod、全部 import 路径（handler/storage/auth/mcpserver/main）、config.go、config_test.go、mcpserver.go、storage.go
- **前端**：web/ 4 个 HTML（title、logo、settings 的 MCP 配置键）
- **构建部署**：Dockerfile、docker-compose.yml、deploy/nginx.conf、Makefile
- **文档**：README.md、CLAUDE.md、SPEC.md 全量替换（含环境变量表、部署命令、示例）
- **顺手修正**：Makefile 引用已废弃的 `HERMES_API_KEY` 与 8080 端口（认证现为用户 token 体系、默认端口 5487），改名时一并对齐，否则改完仍是错的
- **仓库**：GitHub 重命名 + 本地 remote 更新（单独确认后执行）

## 验证

1. `go build ./...`、`go vet`、`go test ./...` 全过（测试同步改用 `ACTA_*`）
2. 本地 E2E：`ACTA_*` 环境变量启动 serve，MCP initialize → `publish_report`，确认返回 URL、server 名 "Acta"、`acta-tags` meta 自动提取均正常
3. `grep -ri hermes` 仅允许出现在历史注释/迁移说明中（目标：零残留）

## 部署迁移（一次性，用户执行）

`.env` 环境变量前缀替换 → `docker compose up -d --build`。数据卷与 metadata 无变化。

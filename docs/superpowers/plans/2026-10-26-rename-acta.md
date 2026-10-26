# HermesPage → Acta 改名实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将项目从 HermesPage 彻底改名为 Acta（品牌、Go 模块路径、二进制、环境变量前缀 ACTA_*、HTML meta 前缀、MCP server 名、构建部署文件、文档、GitHub 仓库），数据层与 API 路径保持不变。

**Architecture:** 纯机械改名，无逻辑变更。规格见 `docs/superpowers/specs/2026-10-26-project-rename-acta-design.md`。不变项：端口 5487、`/api/*`、`/mcp`、`/reports/*`、`tok_` 前缀、metadata.json 格式、`./reports` 数据目录。meta 前缀干净切换（不双读 hermes-*）。

**Tech Stack:** Go 标准库 + mcp-go；纯 HTML/JS 前端；Docker Compose 部署；bash（Git Bash on Windows）。

**注意：** 所有 `sed` 在 Git Bash 下执行。执行 Task 9（仓库改名）前必须再次向用户确认。

---

### Task 1: Go 模块路径 `github.com/cone387/acta`

**Files:**
- Modify: `go.mod`
- Modify: 所有引用 `github.com/hermespage/hermespage` 的 .go 文件（main.go、internal/handler/handler.go、internal/handler/auth_handler.go、internal/handler/user_handler.go 等，以 grep 结果为准）

- [ ] **Step 1: 改模块名并批量替换 import**

```bash
go mod edit -module github.com/cone387/acta
grep -rl 'github.com/hermespage/hermespage' --include='*.go' . | xargs sed -i 's|github.com/hermespage/hermespage|github.com/cone387/acta|g'
```

- [ ] **Step 2: 验证构建与测试**

Run: `go build ./... && go test ./...`
Expected: 全部编译通过，config/mcpserver 测试 PASS

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "refactor: rename Go module to github.com/cone387/acta"
```

---

### Task 2: 环境变量前缀 `HERMES_*` → `ACTA_*`（TDD）

**Files:**
- Modify: `internal/config/config_test.go`
- Modify: `internal/config/config.go`

- [ ] **Step 1: 先改测试（红）**

将 `internal/config/config_test.go` 全量替换为：

```go
package config

import "testing"

func TestServerURLFromEnv(t *testing.T) {
	t.Setenv("ACTA_SERVER_URL", "https://acta.example.com")
	t.Setenv("ACTA_PORT", "5487")

	cfg := Load()
	if cfg.ServerURL != "https://acta.example.com" {
		t.Errorf("ServerURL = %q, want %q", cfg.ServerURL, "https://acta.example.com")
	}
}

func TestServerURLTrimsTrailingSlash(t *testing.T) {
	t.Setenv("ACTA_SERVER_URL", "https://acta.example.com/")
	t.Setenv("ACTA_PORT", "5487")

	cfg := Load()
	if cfg.ServerURL != "https://acta.example.com" {
		t.Errorf("ServerURL = %q, want %q (trailing slash trimmed)", cfg.ServerURL, "https://acta.example.com")
	}
}

func TestServerURLDefault(t *testing.T) {
	t.Setenv("ACTA_SERVER_URL", "")
	t.Setenv("ACTA_PORT", "9000")

	cfg := Load()
	if cfg.ServerURL != "http://localhost:9000" {
		t.Errorf("ServerURL = %q, want %q", cfg.ServerURL, "http://localhost:9000")
	}
}

func TestAllEnvPrefixes(t *testing.T) {
	cfg := Load()
	if cfg.Port != "5487" {
		t.Errorf("Port = %q, want 5487", cfg.Port)
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `go test ./internal/config/`
Expected: FAIL（config.go 仍读 HERMES_*，ServerURL 返回默认值而非期望值）

- [ ] **Step 3: 改实现（绿）**

将 `internal/config/config.go` 全量替换为：

```go
package config

import (
	"os"
	"strings"
)

type Config struct {
	Port       string
	DataDir    string
	WebDir     string
	JWTSecret  string
	AdminUser  string
	AdminPass  string
	ServerURL  string
}

func Load() *Config {
	port := getEnv("ACTA_PORT", "5487")
	serverURL := strings.TrimRight(os.Getenv("ACTA_SERVER_URL"), "/")
	if serverURL == "" {
		serverURL = "http://localhost:" + port
	}
	return &Config{
		Port:       port,
		DataDir:    getEnv("ACTA_DATA_DIR", "./reports"),
		WebDir:     getEnv("ACTA_WEB_DIR", "./web"),
		JWTSecret:  os.Getenv("ACTA_JWT_SECRET"),
		AdminUser:  os.Getenv("ACTA_ADMIN_USER"),
		AdminPass:  os.Getenv("ACTA_ADMIN_PASS"),
		ServerURL:  serverURL,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 4: 运行确认通过**

Run: `go test ./internal/config/ -v`
Expected: 4 个测试全部 PASS

- [ ] **Step 5: 同步 main.go 里的 stdio token 变量**

`main.go` 的 `runMCP()` 中 `token := os.Getenv("HERMES_TOKEN")` 改为 `token := os.Getenv("ACTA_TOKEN")`。

Run: `go build ./... && go test ./...`
Expected: 全部通过

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "refactor: rename env var prefix HERMES_* to ACTA_*"
```

---

### Task 3: HTML meta 前缀 `hermes-*` → `acta-*`（TDD，干净切换）

**Files:**
- Create: `internal/storage/storage_test.go`
- Modify: `internal/storage/storage.go:99-100`

- [ ] **Step 1: 写失败测试**

创建 `internal/storage/storage_test.go`：

```go
package storage

import (
	"reflect"
	"testing"
)

func TestExtractTitleActaMeta(t *testing.T) {
	html := []byte(`<html><head><meta name="acta-title" content="自定义标题"></head><body>x</body></html>`)
	if got := ExtractTitle(html, "fallback.html"); got != "自定义标题" {
		t.Errorf("ExtractTitle = %q, want %q", got, "自定义标题")
	}
}

func TestExtractTitleFallsBackToTitleTag(t *testing.T) {
	html := []byte(`<html><head><title>标签页标题</title></head></html>`)
	if got := ExtractTitle(html, "fallback.html"); got != "标签页标题" {
		t.Errorf("ExtractTitle = %q, want %q", got, "标签页标题")
	}
}

func TestExtractTitleIgnoresLegacyHermesMeta(t *testing.T) {
	html := []byte(`<html><head><meta name="hermes-title" content="旧标题"></head></html>`)
	if got := ExtractTitle(html, "my-report.html"); got != "my report" {
		t.Errorf("ExtractTitle = %q, want %q (hermes meta ignored after clean switch)", got, "my report")
	}
}

func TestExtractTagsActaMeta(t *testing.T) {
	html := []byte(`<html><head><meta name="acta-tags" content="bug, 线上 , 修复"></head></html>`)
	want := []string{"bug", "线上", "修复"}
	if got := ExtractTags(html); !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractTags = %v, want %v", got, want)
	}
}

func TestExtractTagsEmpty(t *testing.T) {
	if got := ExtractTags([]byte(`<html></html>`)); len(got) != 0 {
		t.Errorf("ExtractTags = %v, want empty", got)
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `go test ./internal/storage/`
Expected: `TestExtractTitleActaMeta` 和 `TestExtractTagsActaMeta` FAIL（正则仍匹配 hermes-*）

- [ ] **Step 3: 改正则**

`internal/storage/storage.go` 第 99-100 行：

```go
	reTitleMeta = regexp.MustCompile(`(?i)<meta\s+name=["']acta-title["']\s+content=["']([^"']+)["']`)
	reTagsMeta  = regexp.MustCompile(`(?i)<meta\s+name=["']acta-tags["']\s+content=["']([^"']+)["']`)
```

- [ ] **Step 4: 运行确认通过**

Run: `go test ./internal/storage/ -v`
Expected: 5 个测试全部 PASS（含 hermes meta 被忽略、回落到文件名 "my report"）

- [ ] **Step 5: Commit**

```bash
git add internal/storage/ && git commit -m "refactor: switch HTML meta prefix to acta-title/acta-tags"
```

---

### Task 4: MCP server 品牌名

**Files:**
- Modify: `internal/mcpserver/mcpserver.go`（第 25、42、62 行附近共 3 处 `"HermesPage"`，第 109 行工具描述）

- [ ] **Step 1: 替换（直接编辑文件）**

- 3 处 `server.NewMCPServer(` 之后的 `"HermesPage",`（约第 25、42、62 行）→ `"Acta",`
- `mcp.WithDescription("发布 HTML 报告到 HermesPage，返回访问 URL")` → `mcp.WithDescription("发布 HTML 报告到 Acta，返回访问 URL")`

替换后自查：`grep -n "Acta\|HermesPage" internal/mcpserver/mcpserver.go`，应只有 3 处 `"Acta",` 和 1 处描述文本，无 HermesPage 残留。

- [ ] **Step 2: 验证**

Run: `go build ./... && go test ./...`
Expected: 编译通过、测试 PASS

- [ ] **Step 3: Commit**

```bash
git add internal/mcpserver/mcpserver.go && git commit -m "refactor: rename MCP server to Acta"
```

---

### Task 5: 前端品牌

**Files:**
- Modify: `web/index.html`（第 6 行 title、第 12 行 logo h1）
- Modify: `web/login.html`（第 6 行 title、第 69 行 h2）
- Modify: `web/settings.html`（第 6 行 title、第 191 行 logo h1、第 337 行 MCP 配置键 `"hermespage"` → `"acta"`）
- Modify: `web/setup.html`（第 6 行 title、第 75 行 h2）

- [ ] **Step 1: 批量替换并核对**

```bash
sed -i 's/HermesPage/Acta/g; s/"hermespage"/"acta"/g' web/index.html web/login.html web/settings.html web/setup.html
grep -rn -i "hermes" web/  # Expected: 无输出
```

- [ ] **Step 2: 浏览器冒烟**

Run: `make dev`（或 `ACTA_JWT_SECRET=dev-secret go run . serve`），浏览器打开 http://localhost:5487
Expected: 首页 logo 显示 Acta；登录页/设置页标题为 Acta；设置页 MCP 配置 JSON 中键名为 `acta`

- [ ] **Step 3: Commit**

```bash
git add web/ && git commit -m "refactor: rebrand frontend to Acta"
```

---

### Task 6: 构建与部署文件

**Files:**
- Modify: `Dockerfile`（第 4、12、18 行）
- Modify: `docker-compose.yml`（全量，含服务名）
- Modify: `deploy/nginx.conf`（4 处 hermes.example.com）
- Modify: `Makefile`（全量，修正废弃的 HERMES_API_KEY/8080）
- Modify: `scripts/mockdata.go`（第 119、122、174 行）

- [ ] **Step 1: Dockerfile**

```bash
sed -i 's/hermespage/acta/g' Dockerfile
```

- [ ] **Step 2: docker-compose.yml 全量替换**

```yaml
services:
  acta:
    build: .
    ports:
      - "5487:5487"
    volumes:
      - ./reports:/app/reports
    environment:
      - ACTA_PORT=5487
      - ACTA_DATA_DIR=/app/reports
      - ACTA_WEB_DIR=/app/web
      - ACTA_JWT_SECRET=${ACTA_JWT_SECRET:-change-me-in-production}
      - ACTA_ADMIN_USER=${ACTA_ADMIN_USER:-}
      - ACTA_ADMIN_PASS=${ACTA_ADMIN_PASS:-}
      - ACTA_SERVER_URL=${ACTA_SERVER_URL:-}
    restart: always
```

- [ ] **Step 3: nginx.conf**

```bash
sed -i 's/hermes\.example\.com/acta.example.com/g' deploy/nginx.conf
```

- [ ] **Step 4: Makefile 全量替换**

```makefile
.PHONY: build run dev mcp clean docker-up docker-down test-upload

BINARY=acta
GOFLAGS=-v

build:
	go build $(GOFLAGS) -o $(BINARY) .

run: build
	ACTA_JWT_SECRET=dev-secret ./$(BINARY) serve

dev:
	ACTA_JWT_SECRET=dev-secret go run . serve

mcp:
	ACTA_SERVER_URL=http://localhost:5487 go run . mcp

clean:
	rm -f $(BINARY) $(BINARY).exe

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

# usage: make test-upload ACTA_TOKEN=tok_xxx
test-upload:
	@echo '<html><head><title>Test Report</title><meta name="acta-tags" content="test,demo"></head><body><h1>Hello Acta</h1></body></html>' > /tmp/test-report.html
	curl -X POST -H "Authorization: Bearer $(ACTA_TOKEN)" -F "file=@/tmp/test-report.html" -F "category=test" http://localhost:5487/api/upload
	@echo ""
```

- [ ] **Step 5: scripts/mockdata.go 与 scripts/mock-data.sh**

```bash
sed -i 's/hermes-tags/acta-tags/g; s/HermesPage/Acta/g; s/localhost:8080/localhost:5487/g' scripts/mockdata.go
sed -i 's/hermes-tags/acta-tags/g' scripts/mock-data.sh
```

- [ ] **Step 5b: main.go 品牌字符串**

```bash
sed -i 's/Usage: hermespage <command>/Usage: acta <command>/; s/HermesPage server starting/Acta server starting/' main.go
```

- [ ] **Step 6: 验证构建**

Run: `go build ./... && go vet ./...`
Expected: 通过

- [ ] **Step 7: Commit**

```bash
git add -A && git commit -m "refactor: rename build/deploy artifacts to acta, fix stale Makefile env vars"
```

---

### Task 7: 文档全量替换

**Files:**
- Modify: `README.md`、`CLAUDE.md`、`SPEC.md`

- [ ] **Step 1: 批量替换（保留 "Hermes agent" 作为客户端名称的语义性提及）**

```bash
sed -i 's/HermesPage/Acta/g; s/hermespage/acta/g; s/HERMES_/ACTA_/g; s/hermes-title/acta-title/g; s/hermes-tags/acta-tags/g; s/hermes\.example\.com/acta.example.com/g' README.md CLAUDE.md SPEC.md
```

- [ ] **Step 2: 人工复查 diff**

Run: `git diff README.md CLAUDE.md SPEC.md | head -200`
检查点：环境变量表 8 个变量名、`go run . serve`/`acta mcp` 命令、scp 部署路径 `/opt/acta/`、"Hermes agent" 等指代客户端 agent 的语句应保持原样（如语境变成任意 agent 也可顺手润色，但不大改）。

- [ ] **Step 3: Commit**

```bash
git add README.md CLAUDE.md SPEC.md && git commit -m "docs: rebrand documentation to Acta"
```

---

### Task 8: 残留扫描与端到端验证

**Files:** 无新增，纯验证

- [ ] **Step 1: 残留扫描**

```bash
grep -rin "hermes" --include="*.go" --include="*.sh" --include="*.html" --include="*.js" --include="*.css" --include="*.yml" --include="*.md" --include="*.conf" --include="Dockerfile" --include="Makefile" --include="go.mod" . | grep -v "docs/superpowers"
```

Expected: 无输出（`docs/superpowers/` 下的设计文档保留历史记录，属预期）

- [ ] **Step 2: 全量构建与测试**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: 全部通过

- [ ] **Step 3: MCP 端到端验证（复用 2026-10-26 URL bug 修复时的流程）**

```bash
go build -o /tmp/acta-test.exe . && mkdir -p /tmp/acta-e2e
ACTA_PORT=15487 ACTA_DATA_DIR=/tmp/acta-e2e ACTA_SERVER_URL=https://acta.example.com ACTA_ADMIN_USER=admin ACTA_ADMIN_PASS=pass123 /tmp/acta-test.exe serve &
# 从启动日志取 token（tok_开头）
```

```bash
# initialize（响应中 serverInfo.name 应为 "Acta"，响应头取 Mcp-Session-Id）
curl -s -D /tmp/h.txt -X POST http://localhost:15487/mcp \
  -H "Content-Type: application/json" -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"e2e","version":"1.0"}}}'

# initialized 通知 + publish_report（HTML 带 acta-tags meta，验证新前缀提取）
curl -s -X POST http://localhost:15487/mcp -H "Content-Type: application/json" -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer <TOKEN>" -H "Mcp-Session-Id: <SID>" \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}'

curl -s -X POST http://localhost:15487/mcp -H "Content-Type: application/json" -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer <TOKEN>" -H "Mcp-Session-Id: <SID>" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"publish_report","arguments":{"html_content":"<html><head><meta name=\"acta-tags\" content=\"e2e,acta\"></head><body>t</body></html>","category":"troubleshooting"}}}'
```

Expected:
- initialize 响应含 `"serverInfo":{"name":"Acta"`
- publish_report 返回 `URL: https://acta.example.com/reports/troubleshooting/...`
- 用 `list_reports` 调用或查 `/api/list`（带 token）确认标签为 `e2e,acta`（acta-tags 提取生效）

- [ ] **Step 4: 清理**

```bash
kill %1 2>/dev/null; rm -f /tmp/acta-test.exe /tmp/h.txt; rm -rf /tmp/acta-e2e
```

- [ ] **Step 5: Commit（仅当 Step 1 扫描后有收尾修改；无修改则跳过本步）**

```bash
git add -A && git commit -m "chore: rename cleanup"
```

---

### Task 9: GitHub 仓库改名与推送（执行前必须向用户再次确认）

**Files:** 无文件改动，远程操作

- [ ] **Step 1: 向用户确认**："现在执行仓库改名 cone387/HermesPage → cone387/acta 并推送？"未确认不得继续。

- [ ] **Step 2: 改名（gh 已登录）**

```bash
gh repo rename acta --yes
```

- [ ] **Step 3: 更新本地 remote 并推送**

```bash
git remote set-url origin git@github.com:cone387/acta.git
git push
```

- [ ] **Step 4: 提醒用户部署侧迁移**

服务器 `.env`：`HERMES_*` → `ACTA_*`（含 `ACTA_SERVER_URL=https://正式域名`），然后 `docker compose up -d --build`；Hermes agent 模板中 meta 前缀改 `acta-*`。

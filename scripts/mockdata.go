//go:build ignore

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Report struct {
	ID         string   `json:"id"`
	Filename   string   `json:"filename"`
	Title      string   `json:"title"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`
	Size       int64    `json:"size"`
	CreatedAt  string   `json:"created_at"`
	URL        string   `json:"url"`
	Owner      string   `json:"owner"`
	Visibility string   `json:"visibility"`
}

type Metadata struct {
	Reports []Report `json:"reports"`
}

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Role         string `json:"role"`
	Token        string `json:"token"`
	CreatedAt    string `json:"created_at"`
}

type UsersData struct {
	Users []User `json:"users"`
}

type mockDef struct {
	Title      string
	Category   string
	Tags       []string
	HoursAgo  int
	Owner      string
	Visibility string
}

func main() {
	dataDir := "./reports"

	// Create users
	adminID := "u_" + genRandom(8)
	demoID := "u_" + genRandom(8)

	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	demoHash, _ := bcrypt.GenerateFromPassword([]byte("demo"), bcrypt.DefaultCost)

	adminToken := "tok_" + genRandom(32)
	demoToken := "tok_" + genRandom(32)

	users := UsersData{
		Users: []User{
			{
				ID:           adminID,
				Username:     "admin",
				PasswordHash: string(adminHash),
				Role:         "admin",
				Token:        adminToken,
				CreatedAt:    time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339),
			},
			{
				ID:           demoID,
				Username:     "demo",
				PasswordHash: string(demoHash),
				Role:         "user",
				Token:        demoToken,
				CreatedAt:    time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339),
			},
		},
	}

	mocks := []mockDef{
		{"每日数据监控报告 05-18", "daily", []string{"日报", "监控", "数据"}, 1, demoID, "public"},
		{"用户增长分析 - 5月第三周", "analysis", []string{"增长", "用户", "周报"}, 3, demoID, "public"},
		{"API 响应时间监控", "monitoring", []string{"API", "性能", "告警"}, 5, demoID, "private"},
		{"竞品功能对比 - Claude vs GPT", "analysis", []string{"竞品", "AI", "对比"}, 7, demoID, "public"},
		{"每日数据监控报告 05-17", "daily", []string{"日报", "监控", "数据"}, 25, demoID, "public"},
		{"服务器资源使用率报告", "monitoring", []string{"服务器", "CPU", "内存"}, 27, demoID, "private"},
		{"新功能上线效果评估", "analysis", []string{"功能", "上线", "效果"}, 30, demoID, "public"},
		{"每日数据监控报告 05-16", "daily", []string{"日报", "监控", "数据"}, 50, demoID, "public"},
		{"安全扫描报告 - 第20周", "security", []string{"安全", "漏洞", "扫描"}, 72, adminID, "private"},
		{"数据库性能优化建议", "monitoring", []string{"数据库", "性能", "优化"}, 75, adminID, "public"},
		{"每日数据监控报告 05-15", "daily", []string{"日报", "监控", "数据"}, 80, demoID, "public"},
		{"Q1 业务总结报告", "analysis", []string{"季度", "业务", "总结"}, 240, adminID, "public"},
		{"年度安全审计报告", "security", []string{"安全", "审计", "年度"}, 360, adminID, "private"},
		{"系统架构升级方案", "analysis", []string{"架构", "升级", "方案"}, 500, adminID, "public"},
	}

	os.MkdirAll(dataDir, 0755)
	meta := Metadata{Reports: []Report{}}

	for _, m := range mocks {
		id := genID()
		filename := fmt.Sprintf("%s.html", id)
		createdAt := time.Now().Add(-time.Duration(m.HoursAgo) * time.Hour).UTC()

		html := fmt.Sprintf(`<!DOCTYPE html>
<html><head>
<meta charset="UTF-8">
<title>%s</title>
<meta name="hermes-tags" content="%s">
</head><body>
<h1>%s</h1>
<p>这是一份自动生成的示例报告，用于预览 HermesPage 界面效果。</p>
<p>生成时间: %s</p>
<hr>
<h2>摘要</h2>
<ul>
<li>数据点 A: 1,234</li>
<li>数据点 B: 5,678</li>
<li>数据点 C: 91%%</li>
</ul>
<h2>详细数据</h2>
<table border="1" cellpadding="8">
<tr><th>指标</th><th>当前值</th><th>环比</th></tr>
<tr><td>活跃用户</td><td>12,345</td><td>+5.2%%</td></tr>
<tr><td>转化率</td><td>3.8%%</td><td>+0.3%%</td></tr>
<tr><td>响应时间</td><td>142ms</td><td>-8ms</td></tr>
</table>
</body></html>`, m.Title, joinTags(m.Tags), m.Title, createdAt.Format("2006-01-02 15:04"))

		catDir := filepath.Join(dataDir, m.Category)
		os.MkdirAll(catDir, 0755)
		filePath := filepath.Join(catDir, filename)
		os.WriteFile(filePath, []byte(html), 0644)

		report := Report{
			ID:         id,
			Filename:   filename,
			Title:      m.Title,
			Category:   m.Category,
			Tags:       m.Tags,
			Size:       int64(len(html)),
			CreatedAt:  createdAt.Format(time.RFC3339),
			URL:        fmt.Sprintf("/reports/%s/%s", m.Category, filename),
			Owner:      m.Owner,
			Visibility: m.Visibility,
		}
		meta.Reports = append(meta.Reports, report)
		fmt.Printf("  created: %s [%s] vis=%s owner=%s\n", m.Title, m.Category, m.Visibility, m.Owner)
	}

	// Write metadata
	data, _ := json.MarshalIndent(meta, "", "  ")
	os.WriteFile(filepath.Join(dataDir, "metadata.json"), data, 0644)

	// Write users
	userData, _ := json.MarshalIndent(users, "", "  ")
	os.WriteFile(filepath.Join(dataDir, "users.json"), userData, 0644)

	fmt.Printf("\nDone! Created %d mock reports in %s\n", len(mocks), dataDir)
	fmt.Println("\nAccounts:")
	fmt.Printf("  admin / admin  (role: admin, token: %s)\n", adminToken)
	fmt.Printf("  demo  / demo   (role: user,  token: %s)\n", demoToken)
	fmt.Println("\nStart server: go run . serve")
	fmt.Println("Open http://localhost:8080 and login with demo/demo")
}

func genID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func genRandom(length int) string {
	b := make([]byte, length/2+1)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

func joinTags(tags []string) string {
	result := ""
	for i, t := range tags {
		if i > 0 {
			result += ","
		}
		result += t
	}
	return result
}

package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestExtractSummary(t *testing.T) {
	cases := []struct{ name, html, want string }{
		{"main body", `<html><head><title>报告标题</title><style>.hidden{display:none}</style></head><body><nav>首页 设置</nav><main><h1>报告标题</h1><p>这是第一段，包含<strong>重点</strong>和 &amp; 符号。</p><script>const secret = 'not content';</script><p>第二段继续说明。</p></main><footer>版权信息</footer></body></html>`, "这是第一段，包含重点和 & 符号。 第二段继续说明。"},
		{"fragment and attributes", `<p title="a > b">第一段&nbsp;文字</p><p>第二段<br>下一行</p>`, "第一段 文字 第二段 下一行"},
		{"implicit head", `<title>报告标题</title><h1>报告标题</h1><p>正文内容</p>`, "正文内容"},
		{"omitted head end with body", `<!doctype html><html><head><title>报告标题</title><body><p>正文内容</p></body></html>`, "正文内容"},
		{"omitted head and body tags", `<!doctype html><html><head><title>报告标题</title><style>p{color:red}</style><p>正文内容</p></html>`, "正文内容"},
		{"comments and entities", `<!-- ignore me --><p>&lt;script&gt;文字&lt;/script&gt; &#128512;</p>`, "<script>文字</script> 😀"},
		{"plain text", "  摘要\n\t内容  ", "摘要 内容"},
		{"empty content", `<html><head><title>Title</title></head><body><h1>Title</h1><script>run()</script><style>p{color:red}</style></body></html>`, ""},
		{"unclosed script", `<p>保留正文</p><script>do not show`, "保留正文"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ExtractSummary([]byte(tc.html)); got != tc.want {
				t.Fatalf("summary = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExtractSummaryUsesCharactersRatherThanBytes(t *testing.T) {
	text := strings.Repeat("中", 119) + "🧪"
	if got := ExtractSummary([]byte("<p>" + text + "还有更多内容</p>")); got != text+"…" || !utf8.ValidString(got) {
		t.Fatalf("unexpected truncated summary: %q", got)
	}
	if got := ExtractSummary([]byte(text)); got != text {
		t.Fatal("exactly 120 characters must not get an ellipsis")
	}
}

func TestSavePersistsSummary(t *testing.T) {
	dir := t.TempDir()
	store, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	report, err := store.Save([]byte(`<h1>报告标题</h1><p>这是一份新报告的正文。</p>`), "new.html", "报告标题", "daily", "", "owner", "private", "")
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary != "这是一份新报告的正文。" {
		t.Fatalf("summary = %q", report.Summary)
	}
	reopened, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Get(report.ID).Summary != report.Summary {
		t.Fatal("summary was not retained after reload")
	}
}

func TestOldReportsGetSummaryWithoutChangingSourceFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "daily"), 0755); err != nil {
		t.Fatal(err)
	}
	source := []byte(`<html><body><h1>旧报告</h1><p>已有报告的摘要也可以展示。</p></body></html>`)
	path := filepath.Join(dir, "daily", "old.html")
	if err := os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	metadata, err := json.Marshal(Metadata{Reports: []Report{
		{ID: "old", Category: "daily", Filename: "old.html"},
		{ID: "missing", Category: "daily", Filename: "missing.html"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), metadata, 0644); err != nil {
		t.Fatal(err)
	}
	store, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := store.Get("old").Summary; got != "已有报告的摘要也可以展示。" {
		t.Fatalf("old summary = %q", got)
	}
	if store.Get("missing").Summary != "" {
		t.Fatal("missing source should have no summary")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(source) {
		t.Fatal("original report was modified")
	}
}

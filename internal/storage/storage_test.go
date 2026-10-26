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

// intentional legacy fixture: proves hermes-* meta is ignored after the acta clean switch; excluded from brand residual scans
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

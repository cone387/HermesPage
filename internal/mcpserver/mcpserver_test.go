package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func uploadMockAPI(t *testing.T, gotURL *string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/upload" {
			http.NotFound(w, r)
			return
		}
		*gotURL = "http://" + r.Host + r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":       "abcd1234",
			"url":      "/reports/troubleshooting/report.html",
			"title":    "Test Report",
			"category": "troubleshooting",
			"tags":     []string{"bug"},
		})
	}))
	return ts
}

func publishRequest() mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name:      "publish_report",
		Arguments: map[string]any{"html_content": "<html><body>hi</body></html>"},
	}}
}

func resultText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatal("no content in result")
	}
	tc, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("unexpected content type %T", res.Content[0])
	}
	return tc.Text
}

func TestPublishReportUsesPublicURL(t *testing.T) {
	var gotURL string
	ts := uploadMockAPI(t, &gotURL)
	defer ts.Close()

	c := &apiClient{baseURL: ts.URL, publicURL: "https://reports.example.com"}
	res, err := c.handlePublishReport(context.Background(), publishRequest())
	if err != nil {
		t.Fatalf("handlePublishReport: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %s", resultText(t, res))
	}

	text := resultText(t, res)
	want := "https://reports.example.com/reports/troubleshooting/report.html"
	if !strings.Contains(text, want) {
		t.Errorf("result missing public URL %q, got:\n%s", want, text)
	}
	if strings.Contains(text, ts.URL) {
		t.Errorf("result leaks internal API URL %q, got:\n%s", ts.URL, text)
	}
	if gotURL == "" {
		t.Error("internal upload request never hit the API server")
	}
}

func TestPublishReportPublicURLFallback(t *testing.T) {
	var gotURL string
	ts := uploadMockAPI(t, &gotURL)
	defer ts.Close()

	c := &apiClient{baseURL: ts.URL}
	res, err := c.handlePublishReport(context.Background(), publishRequest())
	if err != nil {
		t.Fatalf("handlePublishReport: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %s", resultText(t, res))
	}

	text := resultText(t, res)
	want := ts.URL + "/reports/troubleshooting/report.html"
	if !strings.Contains(text, want) {
		t.Errorf("result missing fallback URL %q, got:\n%s", want, text)
	}
}

package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Run(serverURL, apiKey string) error {
	s := server.NewMCPServer(
		"HermesPage",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	client := &apiClient{baseURL: serverURL, apiKey: apiKey}

	s.AddTool(publishReportTool(), client.handlePublishReport)
	s.AddTool(listReportsTool(), client.handleListReports)
	s.AddTool(deleteReportTool(), client.handleDeleteReport)
	s.AddTool(getReportInfoTool(), client.handleGetReportInfo)

	stdio := server.NewStdioServer(s)
	return stdio.Listen(context.Background(), os.Stdin, os.Stdout)
}

type apiClient struct {
	baseURL string
	apiKey  string
}

func publishReportTool() mcp.Tool {
	return mcp.NewTool("publish_report",
		mcp.WithDescription("发布 HTML 报告到 HermesPage，返回访问 URL"),
		mcp.WithString("html_content", mcp.Description("HTML 内容字符串（与 file_path 二选一）")),
		mcp.WithString("file_path", mcp.Description("本地 HTML 文件路径（与 html_content 二选一）")),
		mcp.WithString("title", mcp.Description("报告标题（可选，不传则自动提取）")),
		mcp.WithString("tags", mcp.Description("逗号分隔的标签（可选）")),
		mcp.WithString("category", mcp.Description("分类名（可选，默认 uncategorized）")),
	)
}

func listReportsTool() mcp.Tool {
	return mcp.NewTool("list_reports",
		mcp.WithDescription("列出所有已发布的报告，支持按分类和关键词筛选"),
		mcp.WithString("category", mcp.Description("按分类筛选（可选）")),
		mcp.WithString("search", mcp.Description("搜索标题关键词（可选）")),
	)
}

func deleteReportTool() mcp.Tool {
	return mcp.NewTool("delete_report",
		mcp.WithDescription("删除指定报告"),
		mcp.WithString("id", mcp.Description("报告 ID"), mcp.Required()),
	)
}

func getReportInfoTool() mcp.Tool {
	return mcp.NewTool("get_report_info",
		mcp.WithDescription("获取单个报告的详细信息"),
		mcp.WithString("id", mcp.Description("报告 ID"), mcp.Required()),
	)
}

func (c *apiClient) handlePublishReport(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	htmlContent, _ := args["html_content"].(string)
	filePath, _ := args["file_path"].(string)
	title, _ := args["title"].(string)
	tags, _ := args["tags"].(string)
	category, _ := args["category"].(string)

	var content []byte
	var filename string

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("无法读取文件: %v", err)), nil
		}
		content = data
		filename = filepath.Base(filePath)
	} else if htmlContent != "" {
		content = []byte(htmlContent)
		filename = "report.html"
	} else {
		return mcp.NewToolResultError("必须提供 html_content 或 file_path"), nil
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("创建表单失败: %v", err)), nil
	}
	part.Write(content)

	if title != "" {
		writer.WriteField("title", title)
	}
	if tags != "" {
		writer.WriteField("tags", tags)
	}
	if category != "" {
		writer.WriteField("category", category)
	}
	writer.Close()

	httpReq, _ := http.NewRequest("POST", c.baseURL+"/api/upload", &body)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("请求失败: %v", err)), nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return mcp.NewToolResultError(fmt.Sprintf("上传失败 (%d): %s", resp.StatusCode, string(respBody))), nil
	}

	var report struct {
		ID       string   `json:"id"`
		URL      string   `json:"url"`
		Title    string   `json:"title"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
	}
	json.Unmarshal(respBody, &report)

	result := fmt.Sprintf("已发布: \"%s\"\nURL: %s%s\nID: %s\n分类: %s | 标签: %s",
		report.Title, c.baseURL, report.URL, report.ID, report.Category, strings.Join(report.Tags, ", "))

	return mcp.NewToolResultText(result), nil
}

func (c *apiClient) handleListReports(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	category, _ := args["category"].(string)
	search, _ := args["search"].(string)

	url := c.baseURL + "/api/list"
	params := []string{}
	if category != "" {
		params = append(params, "category="+category)
	}
	if search != "" {
		params = append(params, "search="+search)
	}
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	resp, err := http.Get(url)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("请求失败: %v", err)), nil
	}
	defer resp.Body.Close()

	var data struct {
		Reports []struct {
			ID        string   `json:"id"`
			Title     string   `json:"title"`
			Category  string   `json:"category"`
			Tags      []string `json:"tags"`
			URL       string   `json:"url"`
			CreatedAt string   `json:"created_at"`
		} `json:"reports"`
		Total int `json:"total"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	if len(data.Reports) == 0 {
		return mcp.NewToolResultText("暂无报告"), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("共 %d 篇报告:\n\n", data.Total))
	for i, r := range data.Reports {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n   分类: %s | 标签: %s | %s\n   URL: %s%s\n\n",
			i+1, r.ID, r.Title, r.Category, strings.Join(r.Tags, ", "), r.CreatedAt, c.baseURL, r.URL))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func (c *apiClient) handleDeleteReport(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	id, _ := args["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id 参数必填"), nil
	}

	httpReq, _ := http.NewRequest("DELETE", c.baseURL+"/api/delete/"+id, nil)
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("请求失败: %v", err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return mcp.NewToolResultError(fmt.Sprintf("报告 %s 不存在", id)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("已删除报告: %s", id)), nil
}

func (c *apiClient) handleGetReportInfo(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	id, _ := args["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id 参数必填"), nil
	}

	resp, err := http.Get(c.baseURL + "/api/report/" + id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("请求失败: %v", err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return mcp.NewToolResultError(fmt.Sprintf("报告 %s 不存在", id)), nil
	}

	var report struct {
		ID        string   `json:"id"`
		Title     string   `json:"title"`
		Category  string   `json:"category"`
		Tags      []string `json:"tags"`
		Size      int64    `json:"size"`
		URL       string   `json:"url"`
		CreatedAt string   `json:"created_at"`
	}
	json.NewDecoder(resp.Body).Decode(&report)

	sizeStr := fmt.Sprintf("%.1f KB", float64(report.Size)/1024)
	if report.Size > 1024*1024 {
		sizeStr = fmt.Sprintf("%.1f MB", float64(report.Size)/(1024*1024))
	}

	result := fmt.Sprintf("标题: %s\nID: %s\n分类: %s\n标签: %s\n大小: %s\n创建时间: %s\nURL: %s%s",
		report.Title, report.ID, report.Category, strings.Join(report.Tags, ", "),
		sizeStr, report.CreatedAt, c.baseURL, report.URL)

	return mcp.NewToolResultText(result), nil
}

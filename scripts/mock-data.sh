#!/bin/bash
# Mock data script - uploads sample reports for preview
# Usage: HERMES_API_KEY=dev-key bash scripts/mock-data.sh

API_KEY="${HERMES_API_KEY:-dev-key}"
BASE_URL="${HERMES_SERVER_URL:-http://localhost:8080}"

upload() {
    local title="$1" category="$2" tags="$3" hours_ago="$4"
    local content="<html><head><title>${title}</title><meta name=\"hermes-tags\" content=\"${tags}\"></head><body><h1>${title}</h1><p>This is a mock report generated for preview.</p><p>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p></body></html>"

    local tmpfile=$(mktemp --suffix=.html)
    echo "$content" > "$tmpfile"

    curl -s -X POST \
        -H "Authorization: Bearer ${API_KEY}" \
        -F "file=@${tmpfile}" \
        -F "title=${title}" \
        -F "category=${category}" \
        -F "tags=${tags}" \
        "${BASE_URL}/api/upload" > /dev/null

    rm -f "$tmpfile"
    echo "  uploaded: ${title} [${category}]"
}

echo "Uploading mock reports to ${BASE_URL}..."
echo ""

# Today
upload "每日数据监控报告 05-18" "daily" "日报,监控,数据" 1
upload "用户增长分析 - 5月第三周" "analysis" "增长,用户,周报" 2
upload "API 响应时间监控" "monitoring" "API,性能,告警" 3
upload "竞品功能对比 - Claude vs GPT" "analysis" "竞品,AI,对比" 4

# Yesterday (we'll upload them and they'll show as today, but the metadata will show grouping works)
upload "每日数据监控报告 05-17" "daily" "日报,监控,数据" 25
upload "服务器资源使用率报告" "monitoring" "服务器,CPU,内存" 26
upload "新功能上线效果评估" "analysis" "功能,上线,效果" 28

# This week
upload "每日数据监控报告 05-16" "daily" "日报,监控,数据" 49
upload "安全扫描报告 - 第20周" "security" "安全,漏洞,扫描" 50
upload "数据库性能优化建议" "monitoring" "数据库,性能,优化" 52
upload "每日数据监控报告 05-15" "daily" "日报,监控,数据" 73

# Earlier
upload "Q1 业务总结报告" "analysis" "季度,业务,总结" 200
upload "年度安全审计报告" "security" "安全,审计,年度" 300
upload "系统架构升级方案" "analysis" "架构,升级,方案" 400

echo ""
echo "Done! Uploaded 14 mock reports."
echo "Open ${BASE_URL} in your browser to preview."

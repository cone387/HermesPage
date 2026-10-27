package storage

import (
	"html"
	"regexp"
	"strings"
)

const summaryLength = 120

var (
	summaryComments = regexp.MustCompile(`(?s)<!--.*?(?:-->|$)`)
	summaryTags     = regexp.MustCompile(`(?s)<[!/?A-Za-z](?:[^>"']|"[^"]*"|'[^']*')*>`)
	summaryBlocks   = regexp.MustCompile(`(?i)^</?(?:address|article|aside|blockquote|br|dd|div|dl|dt|figcaption|figure|h[1-6]|hr|li|main|ol|p|pre|section|table|tbody|td|th|thead|tr|ul)\b`)
	summaryIgnored  = func() []*regexp.Regexp {
		var patterns []*regexp.Regexp
		// Strip non-content elements individually: HTML permits omitting </head>,
		// so removing the entire head block could also consume the report body.
		for _, tag := range []string{"script", "style", "title", "nav", "header", "footer", "svg", "noscript", "template", "h1"} {
			patterns = append(patterns, regexp.MustCompile(`(?is)<`+tag+`\b(?:[^>"']|"[^"]*"|'[^']*')*>.*?(?:</`+tag+`\s*>|$)`))
		}
		return patterns
	}()
)

// ExtractSummary returns a short, best-effort plain-text excerpt, not sanitized HTML.
// Callers displaying it in HTML must still escape it as ordinary text.
func ExtractSummary(content []byte) string {
	text := summaryComments.ReplaceAllString(string(content), " ")
	for _, pattern := range summaryIgnored {
		text = pattern.ReplaceAllString(text, " ")
	}
	text = summaryTags.ReplaceAllStringFunc(text, func(tag string) string {
		if summaryBlocks.MatchString(tag) {
			return " "
		}
		return ""
	})
	text = strings.Join(strings.Fields(html.UnescapeString(text)), " ")
	characters := []rune(text)
	if len(characters) > summaryLength {
		return strings.TrimSpace(string(characters[:summaryLength])) + "…"
	}
	return text
}

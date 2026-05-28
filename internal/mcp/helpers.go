package mcp

import (
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func mcpResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}
}

func mcpError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: msg,
			},
		},
		IsError: true,
	}
}

func splitColumns(columns string) []string {
	result := []string{}
	for _, col := range strings.Split(columns, ",") {
		col = strings.TrimSpace(col)
		if col != "" {
			result = append(result, col)
		}
	}
	return result
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

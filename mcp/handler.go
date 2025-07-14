package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/foomo/contentserver-mcp/scrape"
	"github.com/foomo/contentserver-mcp/service/vo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ScrapeRequest struct {
	URL      string `json:"url"`      // The URL to scrape
	Selector string `json:"selector"` // CSS selector to extract content
}

type ScrapeResponse struct {
	Summary  *vo.DocumentSummary `json:"summary"`  // The extracted content in markdown format
	Markdown string              `json:"markdown"` // The extracted content in markdown format
}

// NewServer creates a new MCP server with the scrape tool
func NewServer() *server.MCPServer {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Content Scraper MCP",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Create the scrape tool
	tool := mcp.NewTool("scrape",
		mcp.WithDescription("Scrape content from a webpage and convert it to markdown"),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("The URL of the webpage to scrape"),
		),
		mcp.WithString("selector",
			mcp.Required(),
			mcp.Description("CSS selector to extract specific content (e.g., '#content', '.article', 'article')"),
		),
	)

	// Add tool handler
	s.AddTool(tool, mcp.NewTypedToolHandler(scrapeHandler))

	return s
}

// scrapeHandler is our typed handler function that receives strongly-typed arguments
func scrapeHandler(ctx context.Context, request mcp.CallToolRequest, args ScrapeRequest) (*mcp.CallToolResult, error) {
	// Validate inputs
	if args.URL == "" {
		return mcp.NewToolResultError("url is required"), nil
	}
	if args.Selector == "" {
		return mcp.NewToolResultError("selector is required"), nil
	}

	// Call the scrape function
	summary, markdown, err := scrape.Scrape(ctx, args.URL, args.Selector)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to scrape content: %v", err)), nil
	}

	// Create response
	response := ScrapeResponse{
		Summary:  summary,
		Markdown: string(markdown),
	}

	// Convert response to JSON
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseBytes)), nil
}

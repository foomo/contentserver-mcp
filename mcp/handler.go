package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/foomo/contentserver-mcp/scrape"
	"github.com/foomo/contentserver-mcp/service"
	"github.com/foomo/contentserver-mcp/service/vo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const Version = "0.0.1"

type ScrapeRequest struct {
	URL      string `json:"url"`      // The URL to scrape
	Selector string `json:"selector"` // CSS selector to extract content
}

type ScrapeResponse struct {
	Summary  *vo.DocumentSummary `json:"summary"`  // The extracted content in markdown format
	Markdown string              `json:"markdown"` // The extracted content in markdown format
}

type GetDocumentRequest struct {
	Path string `json:"path"` // The path to get the document for
}

type GetDocumentResponse struct {
	Document *vo.Document `json:"document"` // The document with full structure
}

// NewServer creates a new MCP server with the scrape and getDocument tools
func NewServer(client *http.Client, serviceInstance service.Service) *server.MCPServer {
	if client == nil {
		client = http.DefaultClient
	}
	// Create a new MCP server
	s := server.NewMCPServer(
		"Content Scraper MCP",
		Version,
		server.WithToolCapabilities(false),
	)

	// Create the scrape tool
	scrapeTool := mcp.NewTool("scrape",
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

	// Add scrape tool handler
	s.AddTool(scrapeTool, mcp.NewTypedToolHandler(getScrapeHandler(client)))

	// Add getDocument tool only if service is provided
	if serviceInstance != nil {
		getDocumentTool := mcp.NewTool("getDocument",
			mcp.WithDescription("Get a document with full structure including breadcrumbs, siblings, and children"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("The path to get the document for"),
			),
		)
		s.AddTool(getDocumentTool, mcp.NewTypedToolHandler(getDocumentHandler(serviceInstance)))
	}

	return s
}

// scrapeHandler is our typed handler function that receives strongly-typed arguments
func getScrapeHandler(client *http.Client) func(ctx context.Context, request mcp.CallToolRequest, args ScrapeRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest, args ScrapeRequest) (*mcp.CallToolResult, error) {
		// Validate inputs
		if args.URL == "" {
			return mcp.NewToolResultError("url is required"), nil
		}
		if args.Selector == "" {
			return mcp.NewToolResultError("selector is required"), nil
		}

		// Call the scrape function
		summary, markdown, err := scrape.Scrape(ctx, client, args.URL, args.Selector)
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
}

// getDocumentHandler is our typed handler function for the getDocument tool
func getDocumentHandler(serviceInstance service.Service) func(ctx context.Context, request mcp.CallToolRequest, args GetDocumentRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest, args GetDocumentRequest) (*mcp.CallToolResult, error) {
		// Validate inputs
		if args.Path == "" {
			return mcp.NewToolResultError("path is required"), nil
		}

		// Create a mock HTTP request with the context
		req, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create request: %v", err)), nil
		}

		// Call the service to get the document
		document, err := serviceInstance.GetDocument(nil, req, args.Path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get document: %v", err)), nil
		}

		// Create response
		response := GetDocumentResponse{
			Document: document,
		}

		// Convert response to JSON
		responseBytes, err := json.Marshal(response)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(responseBytes)), nil
	}
}

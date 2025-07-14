package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/foomo/contentserver-mcp/scrape"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ScrapeRequest struct {
	URL      string `json:"url"`      // The URL to scrape
	Selector string `json:"selector"` // CSS selector to extract content
}

type ScrapeResponse struct {
	Markdown string `json:"markdown"` // The extracted content in markdown format
}

func main() {
	// Define command line flags
	stdioMode := flag.Bool("stdio", true, "Run in stdio mode")
	httpAddr := flag.String("http", "", "HTTP server address (e.g., ':8080')")
	flag.Parse()

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

	// Determine transport mode
	if *httpAddr != "" {
		// Start the HTTP server
		log.Printf("Starting MCP server on HTTP address: %s", *httpAddr)
		httpServer := server.NewStreamableHTTPServer(s)
		if err := httpServer.Start(*httpAddr); err != nil {
			log.Fatal(err)
		}
	} else if *stdioMode {
		// Start the stdio server
		log.Println("Starting MCP server in stdio mode...")
		if err := server.ServeStdio(s); err != nil {
			log.Fatal(err)
		}
	} else {
		// Default to stdio mode if no flags provided
		log.Println("Starting MCP server in stdio mode (default)...")
		if err := server.ServeStdio(s); err != nil {
			log.Fatal(err)
		}
	}
}

// Our typed handler function that receives strongly-typed arguments
func scrapeHandler(ctx context.Context, request mcp.CallToolRequest, args ScrapeRequest) (*mcp.CallToolResult, error) {
	// Validate inputs
	if args.URL == "" {
		return mcp.NewToolResultError("url is required"), nil
	}
	if args.Selector == "" {
		return mcp.NewToolResultError("selector is required"), nil
	}

	// Call the scrape function
	markdown, err := scrape.Scrape(ctx, args.URL, args.Selector)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to scrape content: %v", err)), nil
	}

	// Create response
	response := ScrapeResponse{
		Markdown: string(markdown),
	}

	// Convert response to JSON
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseBytes)), nil
}

package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewServer(t *testing.T) {
	// Test that we can create a server
	server := NewServer(http.DefaultClient, nil)
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
}

func TestScrapeHandler(t *testing.T) {
	// Test the scrape handler with valid arguments
	args := ScrapeRequest{
		URL:      "https://example.com",
		Selector: "body",
	}

	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
		Params: mcp.CallToolParams{
			Name:      "scrape",
			Arguments: args,
		},
	}

	ctx := context.Background()
	scrapeHandler := getScrapeHandler(http.DefaultClient)
	result, err := scrapeHandler(ctx, request, args)
	if err != nil {
		t.Fatalf("scrapeHandler returned error: %v", err)
	}

	if result == nil {
		t.Fatal("scrapeHandler returned nil result")
	}

	// Check that we got a text result (even if it's an error due to network issues)
	if len(result.Content) == 0 {
		t.Fatal("scrapeHandler returned no content")
	}
}

func TestScrapeHandlerValidation(t *testing.T) {
	scrapeHandler := getScrapeHandler(http.DefaultClient)
	// Test validation for missing URL
	args := ScrapeRequest{
		URL:      "",
		Selector: "body",
	}

	request := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
		Params: mcp.CallToolParams{
			Name:      "scrape",
			Arguments: args,
		},
	}

	ctx := context.Background()
	result, err := scrapeHandler(ctx, request, args)
	if err != nil {
		t.Fatalf("scrapeHandler returned error: %v", err)
	}

	if result == nil {
		t.Fatal("scrapeHandler returned nil result")
	}

	// Should return an error result
	if !result.IsError {
		t.Fatal("Expected error result for missing URL")
	}
}

func TestScrapeRequestMarshal(t *testing.T) {
	// Test that ScrapeRequest can be marshaled to JSON
	req := ScrapeRequest{
		URL:      "https://example.com",
		Selector: "#content",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal ScrapeRequest: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}
}

func TestScrapeResponseMarshal(t *testing.T) {
	// Test that ScrapeResponse can be marshaled to JSON
	resp := ScrapeResponse{
		Markdown: "# Test Content\n\nThis is test content.",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal ScrapeResponse: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}
}

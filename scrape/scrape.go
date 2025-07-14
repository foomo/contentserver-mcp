package scrape

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/foomo/contentserver-mcp/service/vo"
	"golang.org/x/net/html"
)

func Scrape(ctx context.Context, url, selector string) (*vo.DocumentSummary, vo.Markdown, error) {
	// Download HTML from URL
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download HTML: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract document metadata
	title := extractTitle(doc)
	description := extractMetaDescription(doc)
	keywords := extractMetaKeywords(doc)

	// Create document summary
	summary := &vo.DocumentSummary{
		URL: url,
		ContentSummary: vo.ContentSummary{
			Title:       title,
			Description: description,
			Keywords:    keywords,
		},
	}

	// Extract node using selector
	selectedNode, err := extractNodeBySelector(doc, selector)
	if err != nil {
		return summary, "", fmt.Errorf("failed to extract node with selector '%s': %w", selector, err)
	}

	// Convert HTML node to markdown
	markdownBytes, err := htmltomarkdown.ConvertNode(selectedNode)
	if err != nil {
		return summary, "", fmt.Errorf("failed to convert HTML to markdown: %w", err)
	}

	return summary, vo.Markdown(string(markdownBytes)), nil
}

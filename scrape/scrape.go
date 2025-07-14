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

func Scrape(ctx context.Context, url, selector string) (vo.Markdown, error) {
	// Download HTML from URL
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download HTML: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract node using selector
	selectedNode, err := extractNodeBySelector(doc, selector)
	if err != nil {
		return "", fmt.Errorf("failed to extract node with selector '%s': %w", selector, err)
	}

	// Convert HTML node to markdown
	markdownBytes, err := htmltomarkdown.ConvertNode(selectedNode)
	if err != nil {
		return "", fmt.Errorf("failed to convert HTML to markdown: %w", err)
	}

	return vo.Markdown(string(markdownBytes)), nil
}

// extractNodeBySelector finds a node in the HTML document using a CSS selector
// This is a simplified implementation - for production use, consider using a proper CSS selector library
func extractNodeBySelector(doc *html.Node, selector string) (*html.Node, error) {
	// For now, we'll implement a basic selector that looks for elements by tag name
	// This can be extended to support more complex CSS selectors
	if strings.HasPrefix(selector, "#") {
		// ID selector
		id := strings.TrimPrefix(selector, "#")
		return findNodeByID(doc, id)
	} else if strings.HasPrefix(selector, ".") {
		// Class selector
		class := strings.TrimPrefix(selector, ".")
		return findNodeByClass(doc, class)
	} else {
		// Tag selector
		return findNodeByTag(doc, selector)
	}
}

func findNodeByID(n *html.Node, id string) (*html.Node, error) {
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if attr.Key == "id" && attr.Val == id {
				return n, nil
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result, err := findNodeByID(c, id); err == nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("element with id '%s' not found", id)
}

func findNodeByClass(n *html.Node, class string) (*html.Node, error) {
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, class) {
				return n, nil
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result, err := findNodeByClass(c, class); err == nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("element with class '%s' not found", class)
}

func findNodeByTag(n *html.Node, tag string) (*html.Node, error) {
	if n.Type == html.ElementNode && n.Data == tag {
		return n, nil
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result, err := findNodeByTag(c, tag); err == nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("element with tag '%s' not found", tag)
}

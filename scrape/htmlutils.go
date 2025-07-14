package scrape

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

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

// extractTitle extracts the title from the HTML document
func extractTitle(doc *html.Node) string {
	var title string
	var findTitle func(*html.Node)

	findTitle = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				title = n.FirstChild.Data
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTitle(c)
		}
	}

	findTitle(doc)
	return title
}

// extractMetaDescription extracts the meta description from the HTML document
func extractMetaDescription(doc *html.Node) string {
	var description string
	var findMeta func(*html.Node)

	findMeta = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, attr := range n.Attr {
				if attr.Key == "name" && attr.Val == "description" {
					name = attr.Val
				}
				if attr.Key == "content" {
					content = attr.Val
				}
			}
			if name == "description" && content != "" {
				description = content
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findMeta(c)
		}
	}

	findMeta(doc)
	return description
}

// extractMetaKeywords extracts the meta keywords from the HTML document
func extractMetaKeywords(doc *html.Node) []string {
	var keywords []string
	var findMeta func(*html.Node)

	findMeta = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, attr := range n.Attr {
				if attr.Key == "name" && attr.Val == "keywords" {
					name = attr.Val
				}
				if attr.Key == "content" {
					content = attr.Val
				}
			}
			if name == "keywords" && content != "" {
				// Split keywords by comma and trim whitespace
				keywordList := strings.Split(content, ",")
				for _, keyword := range keywordList {
					trimmed := strings.TrimSpace(keyword)
					if trimmed != "" {
						keywords = append(keywords, trimmed)
					}
				}
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findMeta(c)
		}
	}

	findMeta(doc)
	return keywords
}

package vo

type (
	Markdown string
	MimeType string

	ContentSummary struct {
		Title       string   `json:"title"`       // Page title
		Name        string   `json:"name"`        // (short) name
		Description string   `json:"description"` // 2-3 sentence abstract
		Keywords    []string `json:"keywords"`    // Keywords
	}

	DocumentSummary struct {
		MimeType       MimeType       `json:"mimeType"`
		ID             string         `json:"id"`
		URL            string         `json:"url"` // Unique identifier (URL hash or custom ID)
		ContentSummary ContentSummary `json:"contentSummary"`
	}
	Document struct {
		DocumentSummary DocumentSummary `json:"documentSummary"`
		Markdown        Markdown        `json:"markdown,omitempty"` // Full content in markdown

		Breadcrump   []DocumentSummary `json:"breadcrump,omitempty"`
		Children     []DocumentSummary `json:"children,omitempty"`     // Child page IDs
		PrevSiblings []DocumentSummary `json:"prevSiblings,omitempty"` // Previous sibling ID
		NextSiblings []DocumentSummary `json:"nextSiblings,omitempty"` // Next sibling ID
	}
)

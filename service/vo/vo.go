package vo

type Markdown string

type ContentSummary struct {
	Title       string   `json:"title"`       // Page title
	Description string   `json:"description"` // 2-3 sentence abstract
	Keywords    []string `json:"keywords"`    // Keywords
}

type DocumentSummary struct {
	URL            string `json:"url"` // Unique identifier (URL hash or custom ID)
	ContentSummary `json:"contentSummary"`
}

// type Article struct {
// 	Anchor string `json:"anchor,omitempty"` // Anchor text
// 	ContentSummary
// }

type Document struct {
	DocumentSummary DocumentSummary
	Markdown        Markdown `json:"markdown,omitempty"` // Full content in markdown
	// 	Articles        []Article `json:"articles,omitempty"`

	Breadcrump   []DocumentSummary `json:"breadcrump,omitempty"`
	Children     []DocumentSummary `json:"children,omitempty"` // Child page IDs
	PrevSiblings []DocumentSummary `json:"prev,omitempty"`     // Previous sibling ID
	NextSiblings []DocumentSummary `json:"next,omitempty"`     // Next sibling ID
}

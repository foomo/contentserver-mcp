package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/foomo/contentserver-mcp/scrape"
	"github.com/foomo/contentserver-mcp/service/vo"
	contentserverclient "github.com/foomo/contentserver/client"
	"github.com/foomo/contentserver/content"
	"github.com/foomo/contentserver/requests"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	GetDocument(w http.ResponseWriter, r *http.Request, path string) (*vo.Document, error)
}

type service struct {
	l                   *zap.Logger
	contentServerClient *contentserverclient.Client
	httpClient          *http.Client
	siteSettings        SiteSettings
	contentScrapers     map[vo.MimeType]ContentScraper
}

type SiteContextService interface {
	GetContext(path string) (string, error)
}

type ContentScraper func(ctx context.Context, httpClient *http.Client, siteSettings SiteSettings, content *content.SiteContent) (vo.Markdown, error)

type SiteSettings struct {
	Env              *requests.Env
	ContentSelector  string
	BaseURL          string
	ContentServerURL string
	MimeTypes        []vo.MimeType
}

func (siteSettings SiteSettings) mimeTypes() []string {
	mimeTypes := make([]string, len(siteSettings.MimeTypes))
	for i, mimeType := range siteSettings.MimeTypes {
		mimeTypes[i] = string(mimeType)
	}
	return mimeTypes
}

func NewService(
	l *zap.Logger,
	siteSettings SiteSettings,
	httpClient *http.Client,
	contentScrapers map[vo.MimeType]ContentScraper,
) Service {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	contentServerClient := contentserverclient.New(
		contentserverclient.NewHTTPTransport(
			siteSettings.ContentServerURL,
			contentserverclient.HTTPTransportWithHTTPClient(httpClient),
		))

	return &service{
		l:                   l,
		siteSettings:        siteSettings,
		httpClient:          httpClient,
		contentServerClient: contentServerClient,
		contentScrapers:     contentScrapers,
	}
}

// isValidURI checks if a URI is valid for processing
func isValidURI(uri string) bool {
	return uri != "" && strings.HasPrefix(uri, "/")
}

func (s *service) GetDocument(w http.ResponseWriter, r *http.Request, path string) (*vo.Document, error) {
	requestID := ""
	if r != nil {
		requestID = r.Header.Get("X-Request-ID")
	}
	if requestID == "" {
		requestID = uuid.New().String()
	}
	l := s.l.With(zap.String("path", path), zap.String("requestID", requestID))
	l.Info("serving GetDocument")

	var ctx context.Context
	if r != nil {
		ctx = r.Context()
	} else {
		ctx = context.Background()
	}

	l.Debug("Getting content from content server")
	content, err := s.contentServerClient.GetContent(ctx, &requests.Content{
		URI:   path,
		Env:   s.siteSettings.Env,
		Nodes: map[string]*requests.Node{},
	})

	if err != nil {
		l.Error("Failed to get content from content server", zap.Error(err))
		return nil, err
	}
	l.Debug("Content retrieved successfully", zap.String("mimeType", content.MimeType), zap.String("itemID", content.Item.ID))

	breadcrump := make([]vo.DocumentSummary, len(content.Path))
	l.Debug("Processing breadcrumb path", zap.Int("pathLength", len(content.Path)))

	for i, item := range content.Path {
		if !isValidURI(item.URI) {
			l.Debug("Skipping invalid URI in breadcrumb", zap.String("uri", item.URI))
			continue
		}
		l.Debug("Scraping breadcrumb item", zap.String("uri", item.URI), zap.Int("index", i))
		summary, _, err := scrape.Scrape(ctx, s.httpClient, s.siteSettings.BaseURL+item.URI, s.siteSettings.ContentSelector)
		if err != nil {
			l.Error("Failed to scrape breadcrumb item", zap.String("uri", item.URI), zap.Error(err))
			return nil, err
		}
		breadcrump[len(content.Path)-i-1] = *summary
	}

	l.Debug("Scraping main document", zap.String("url", s.siteSettings.BaseURL+path))
	summary, markdown, err := scrape.Scrape(ctx, s.httpClient, s.siteSettings.BaseURL+path, s.siteSettings.ContentSelector)
	if err != nil {
		l.Error("Failed to scrape main document", zap.Error(err))
		return nil, err
	}
	l.Debug("Main document scraped successfully")

	contentScraper, ok := s.contentScrapers[vo.MimeType(content.MimeType)]
	if ok {
		l.Debug("Applying content scraper", zap.String("mimeType", content.MimeType))
		markdown, err = contentScraper(ctx, s.httpClient, s.siteSettings, content)
		if err != nil {
			l.Error("Content scraper failed", zap.String("mimeType", content.MimeType), zap.Error(err))
			return nil, err
		}
		l.Debug("Content scraper applied successfully", zap.String("mimeType", content.MimeType))
	} else {
		l.Debug("No content scraper found for mime type", zap.String("mimeType", content.MimeType))
	}

	s.loadItemData(summary, content.Item)
	doc := &vo.Document{
		DocumentSummary: *summary,
		Breadcrump:      breadcrump,
		Markdown:        markdown,
	}

	isPrevious := true
	if len(content.Path) > 0 {
		l.Debug("Processing siblings", zap.String("parentID", content.Path[0].ID))
		parent := content.Path[0]
		nodes, err := s.contentServerClient.GetNodes(ctx, s.siteSettings.Env, map[string]*requests.Node{
			parent.ID: {
				ID:        parent.ID,
				MimeTypes: s.siteSettings.mimeTypes(),
			},
		})
		if err != nil {
			l.Error("Failed to get parent nodes", zap.String("parentID", parent.ID), zap.Error(err))
			return nil, err
		}
		parentNode, ok := nodes[parent.ID]
		if !ok {
			l.Error("Parent node not found", zap.String("parentID", parent.ID))
			return nil, errors.New("parent node not found")
		}
		l.Debug("Processing sibling nodes", zap.Int("siblingCount", len(parentNode.Index)))

		for _, id := range parentNode.Index {
			if id == content.Item.ID {
				l.Debug("Found current item in siblings, switching to next siblings", zap.String("itemID", id))
				isPrevious = false
				continue
			}

			siblingNode, ok := parentNode.Nodes[id]
			if !ok {
				l.Error("Sibling node not found", zap.String("nodeID", id))
				return nil, errors.New("sibling node not found")
			}
			if !isValidURI(siblingNode.Item.URI) {
				l.Debug("Skipping sibling with invalid URI", zap.String("uri", siblingNode.Item.URI))
				continue
			}

			l.Debug("Scraping sibling", zap.String("uri", siblingNode.Item.URI), zap.Bool("isPrevious", isPrevious))
			siblingSummary, _, err := scrape.Scrape(ctx, s.httpClient, s.siteSettings.BaseURL+siblingNode.Item.URI, s.siteSettings.ContentSelector)
			if err != nil {
				l.Error("Failed to scrape sibling", zap.String("uri", siblingNode.Item.URI), zap.Error(err))
				return nil, err
			}
			s.loadItemData(siblingSummary, siblingNode.Item)
			if isPrevious {
				doc.PrevSiblings = append(doc.PrevSiblings, *siblingSummary)
			} else {
				doc.NextSiblings = append(doc.NextSiblings, *siblingSummary)
			}
		}
		l.Debug("Siblings processed", zap.Int("prevSiblings", len(doc.PrevSiblings)), zap.Int("nextSiblings", len(doc.NextSiblings)))
	}

	l.Debug("Getting child nodes", zap.String("itemID", content.Item.ID))
	nodes, err := s.contentServerClient.GetNodes(ctx, s.siteSettings.Env, map[string]*requests.Node{
		content.Item.ID: {
			ID:        content.Item.ID,
			MimeTypes: s.siteSettings.mimeTypes(),
		},
	})
	if err != nil {
		l.Error("Failed to get child nodes", zap.String("itemID", content.Item.ID), zap.Error(err))
		return nil, err
	}

	contentNode, ok := nodes[content.Item.ID]
	if !ok {
		l.Error("Content node not found", zap.String("itemID", content.Item.ID))
		return nil, errors.New("content node not found")
	}

	l.Debug("Processing child nodes", zap.Int("childCount", len(contentNode.Index)))
	for _, id := range contentNode.Index {
		childNode, ok := contentNode.Nodes[id]
		if !ok {
			l.Error("Child node not found", zap.String("nodeID", id))
			return nil, errors.New("child node not found")
		}
		l.Debug("Scraping child", zap.String("uri", childNode.Item.URI))
		childSummary, _, err := scrape.Scrape(ctx, s.httpClient, s.siteSettings.BaseURL+childNode.Item.URI, s.siteSettings.ContentSelector)
		if err != nil {
			l.Error("Failed to scrape child", zap.String("uri", childNode.Item.URI), zap.Error(err))
			return nil, err
		}
		s.loadItemData(childSummary, childNode.Item)
		doc.Children = append(doc.Children, *childSummary)
	}

	l.Info("GetDocument completed successfully",
		zap.Int("breadcrumbLength", len(doc.Breadcrump)),
		zap.Int("prevSiblings", len(doc.PrevSiblings)),
		zap.Int("nextSiblings", len(doc.NextSiblings)),
		zap.Int("children", len(doc.Children)))

	return doc, nil
}

func (s *service) loadItemData(d *vo.DocumentSummary, item *content.Item) {
	d.MimeType = vo.MimeType(item.MimeType)
	d.ID = item.ID
	d.URL = s.siteSettings.BaseURL + item.URI
}

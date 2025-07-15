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
)

type Service interface {
	GetDocument(w http.ResponseWriter, r *http.Request, path string) (*vo.Document, error)
}

type service struct {
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
	var ctx context.Context
	if r != nil {
		ctx = r.Context()
	} else {
		ctx = context.Background()
	}
	content, err := s.contentServerClient.GetContent(ctx, &requests.Content{
		URI:   path,
		Env:   s.siteSettings.Env,
		Nodes: map[string]*requests.Node{},
	})

	if err != nil {
		return nil, err
	}

	breadcrump := make([]vo.DocumentSummary, len(content.Path))

	for i, item := range content.Path {
		if !isValidURI(item.URI) {
			continue
		}
		summary, _, err := scrape.Scrape(ctx, s.httpClient, s.siteSettings.BaseURL+item.URI, s.siteSettings.ContentSelector)
		if err != nil {
			return nil, err
		}
		breadcrump[len(content.Path)-i-1] = *summary
	}

	summary, markdown, err := scrape.Scrape(ctx, s.httpClient, s.siteSettings.BaseURL+path, s.siteSettings.ContentSelector)
	if err != nil {
		return nil, err
	}

	contentScraper, ok := s.contentScrapers[vo.MimeType(content.MimeType)]
	if ok {
		markdown, err = contentScraper(ctx, s.httpClient, s.siteSettings, content)
		if err != nil {
			return nil, err
		}
	}
	s.loadItemData(summary, content.Item)
	doc := &vo.Document{
		DocumentSummary: *summary,
		Breadcrump:      breadcrump,
		Markdown:        markdown,
	}

	isPrevious := true
	if len(content.Path) > 0 {
		parent := content.Path[0]
		nodes, err := s.contentServerClient.GetNodes(ctx, s.siteSettings.Env, map[string]*requests.Node{
			parent.ID: {
				ID:        parent.ID,
				MimeTypes: s.siteSettings.mimeTypes(),
			},
		})
		if err != nil {
			return nil, err
		}
		parentNode, ok := nodes[parent.ID]
		if !ok {
			return nil, errors.New("parent node not found")
		}
		for _, id := range parentNode.Index {
			if id == content.Item.ID {
				isPrevious = false
				continue
			}

			siblingNode, ok := parentNode.Nodes[id]
			if !ok {
				return nil, errors.New("sibling node not found")
			}
			if !isValidURI(siblingNode.Item.URI) {
				continue
			}

			siblingSummary, _, err := scrape.Scrape(ctx, s.httpClient, s.siteSettings.BaseURL+siblingNode.Item.URI, s.siteSettings.ContentSelector)
			if err != nil {
				return nil, err
			}
			s.loadItemData(siblingSummary, siblingNode.Item)
			if isPrevious {
				doc.PrevSiblings = append(doc.PrevSiblings, *siblingSummary)
			} else {
				doc.NextSiblings = append(doc.NextSiblings, *siblingSummary)
			}

		}

	}

	nodes, err := s.contentServerClient.GetNodes(ctx, s.siteSettings.Env, map[string]*requests.Node{
		content.Item.ID: {
			ID:        content.Item.ID,
			MimeTypes: s.siteSettings.mimeTypes(),
		},
	})

	contentNode, ok := nodes[content.Item.ID]
	if !ok {
		return nil, errors.New("content node not found")
	}

	for _, id := range contentNode.Index {
		childNode, ok := contentNode.Nodes[id]
		if !ok {
			return nil, errors.New("child node not found")
		}
		childSummary, _, err := scrape.Scrape(ctx, s.httpClient, s.siteSettings.BaseURL+childNode.Item.URI, s.siteSettings.ContentSelector)
		if err != nil {
			return nil, err
		}
		s.loadItemData(childSummary, childNode.Item)
		doc.Children = append(doc.Children, *childSummary)
	}
	return doc, nil
}

func (s *service) loadItemData(d *vo.DocumentSummary, item *content.Item) {
	d.MimeType = vo.MimeType(item.MimeType)
	d.ID = item.ID
	d.URL = s.siteSettings.BaseURL + item.URI
}

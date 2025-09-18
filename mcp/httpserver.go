package mcp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/foomo/contentserver-mcp/service"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type McpHTTPServer struct {
	server   *server.MCPServer
	endpoint string
}

// httpRequestKey is a custom context key for storing the original HTTP request
type httpRequestKey struct{}

// withHTTPRequest adds the original HTTP request to the context
func withHTTPRequest(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey{}, req)
}

// httpRequestFromContext extracts the original HTTP request from the context
func httpRequestFromContext(ctx context.Context) (*http.Request, bool) {
	req, ok := ctx.Value(httpRequestKey{}).(*http.Request)
	return req, ok
}

// httpContextFunc extracts the original HTTP request and adds it to the context
func httpContextFunc(ctx context.Context, r *http.Request) context.Context {
	return withHTTPRequest(ctx, r)
}

// NewMcpHTTPServer creates a new MCP HTTP server with traditional MCP endpoints
func NewMcpHTTPServer(s *server.MCPServer, endpoint string) *server.StreamableHTTPServer {
	return server.NewStreamableHTTPServer(
		s,
		server.WithEndpointPath(endpoint),
		server.WithHTTPContextFunc(httpContextFunc),
	)
}

// NewMcpHTTPSSEServer creates a new MCP server with both HTTP and SSE capabilities
func NewMcpHTTPSSEServer(logger *zap.Logger, s *server.MCPServer, serviceInstance service.Service, httpClient *http.Client, endpoint string, config *SSEServerConfig) *McpHTTPSSEServer {
	// Create the SSE server
	sseServer := NewMCPSSEServer(logger, s, serviceInstance, httpClient, config)

	// Create HTTP mux for both MCP and SSE endpoints
	mux := http.NewServeMux()

	// Add MCP server endpoint
	mcpHandler := server.NewStreamableHTTPServer(
		s,
		server.WithEndpointPath(endpoint),
		server.WithHTTPContextFunc(httpContextFunc),
	)
	mux.Handle(endpoint, mcpHandler)

	// Add SSE endpoints
	mux.HandleFunc(endpoint+"/sse", sseServer.HandleSSE)
	mux.HandleFunc(endpoint+"/sse/scrape", sseServer.HandleScrapeSSE)
	mux.HandleFunc(endpoint+"/sse/document", sseServer.HandleGetDocumentSSE)
	mux.HandleFunc(endpoint+"/sse/clients", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		clients := sseServer.GetConnectedClients()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"connectedClients": len(clients),
			"clients":          clients,
		})
	})
	mux.HandleFunc(endpoint+"/sse/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		stats := sseServer.GetStats()
		json.NewEncoder(w).Encode(stats)
	})

	return &McpHTTPSSEServer{
		mux:       mux,
		sseServer: sseServer,
	}
}

// McpHTTPSSEServer combines MCP HTTP server with SSE capabilities
type McpHTTPSSEServer struct {
	mux       *http.ServeMux
	sseServer *MCPSSEServer
}

// ServeHTTP implements http.Handler
func (s *McpHTTPSSEServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// GetSSEServer returns the underlying SSE server for direct access
func (s *McpHTTPSSEServer) GetSSEServer() *MCPSSEServer {
	return s.sseServer
}

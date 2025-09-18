package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/foomo/contentserver-mcp/scrape"
	"github.com/foomo/contentserver-mcp/service"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// SSEEvent represents an SSE event structure
type SSEEvent struct {
	ID        string      `json:"id"`
	Event     string      `json:"event"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// SSEClient represents a connected SSE client
type SSEClient struct {
	ID       string
	Writer   http.ResponseWriter
	Flusher  http.Flusher
	Done     chan struct{}
	LastSeen time.Time
}

// MCPSSEServer wraps the MCP server with SSE capabilities
type MCPSSEServer struct {
	logger       *zap.Logger
	mcpServer    *server.MCPServer
	service      service.Service
	httpClient   *http.Client
	clients      map[string]*SSEClient
	clientsMutex sync.RWMutex
	broadcast    chan SSEEvent
	nextClientID int
}

// SSEServerConfig holds configuration for the SSE server
type SSEServerConfig struct {
	KeepaliveInterval time.Duration
	BufferSize        int
	ClientTimeout     time.Duration
}

// DefaultSSEServerConfig returns the default configuration for SSE server
func DefaultSSEServerConfig() *SSEServerConfig {
	return &SSEServerConfig{
		KeepaliveInterval: 30 * time.Second,
		BufferSize:        100,
		ClientTimeout:     60 * time.Second,
	}
}

// NewMCPSSEServer creates a new MCP SSE server
func NewMCPSSEServer(logger *zap.Logger, mcpServer *server.MCPServer, serviceInstance service.Service, httpClient *http.Client, config *SSEServerConfig) *MCPSSEServer {
	if config == nil {
		config = DefaultSSEServerConfig()
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	sseServer := &MCPSSEServer{
		logger:     logger,
		mcpServer:  mcpServer,
		service:    serviceInstance,
		httpClient: httpClient,
		clients:    make(map[string]*SSEClient),
		broadcast:  make(chan SSEEvent, config.BufferSize),
	}

	// Start the broadcast loop
	go sseServer.broadcastLoop(config)

	return sseServer
}

// broadcastLoop handles broadcasting events to all connected clients
func (s *MCPSSEServer) broadcastLoop(config *SSEServerConfig) {
	for event := range s.broadcast {
		s.clientsMutex.RLock()
		for clientID, client := range s.clients {
			select {
			case <-client.Done:
				// Client disconnected, remove it
				s.clientsMutex.RUnlock()
				s.removeClient(clientID)
				s.clientsMutex.RLock()
				continue
			default:
				// Send event to client
				if err := s.sendEventToClient(client, event); err != nil {
					s.logger.Error("failed to send event to client", zap.String("clientID", clientID), zap.Error(err))
					s.clientsMutex.RUnlock()
					s.removeClient(clientID)
					s.clientsMutex.RLock()
				}
			}
		}
		s.clientsMutex.RUnlock()
	}
}

// sendEventToClient sends an SSE event to a specific client
func (s *MCPSSEServer) sendEventToClient(client *SSEClient, event SSEEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Format as SSE
	fmt.Fprintf(client.Writer, "id: %s\n", event.ID)
	fmt.Fprintf(client.Writer, "event: %s\n", event.Event)
	fmt.Fprintf(client.Writer, "data: %s\n\n", string(eventJSON))

	client.Flusher.Flush()
	client.LastSeen = time.Now()

	return nil
}

// addClient adds a new SSE client
func (s *MCPSSEServer) addClient(w http.ResponseWriter, r *http.Request) *SSEClient {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return nil
	}

	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	s.nextClientID++
	clientID := fmt.Sprintf("client_%d_%d", time.Now().Unix(), s.nextClientID)

	client := &SSEClient{
		ID:       clientID,
		Writer:   w,
		Flusher:  flusher,
		Done:     make(chan struct{}),
		LastSeen: time.Now(),
	}

	s.clients[clientID] = client

	// Send connection confirmation
	connectEvent := SSEEvent{
		ID:        fmt.Sprintf("connect_%d", time.Now().UnixNano()),
		Event:     "connected",
		Data:      map[string]string{"clientID": clientID, "message": "Connected to MCP SSE server"},
		Timestamp: time.Now(),
	}

	if err := s.sendEventToClient(client, connectEvent); err != nil {
		s.logger.Error("failed to send connection event", zap.String("clientID", clientID), zap.Error(err))
		delete(s.clients, clientID)
		return nil
	}

	s.logger.Info("SSE client connected", zap.String("clientID", clientID))
	return client
}

// removeClient removes a client from the server
func (s *MCPSSEServer) removeClient(clientID string) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	if client, exists := s.clients[clientID]; exists {
		close(client.Done)
		delete(s.clients, clientID)
		s.logger.Info("SSE client disconnected", zap.String("clientID", clientID))
	}
}

// broadcastEvent sends an event to all connected clients
func (s *MCPSSEServer) broadcastEvent(event SSEEvent) {
	select {
	case s.broadcast <- event:
	default:
		s.logger.Warn("broadcast channel full, dropping event", zap.String("eventID", event.ID))
	}
}

// HandleSSE handles SSE client connections
func (s *MCPSSEServer) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	client := s.addClient(w, r)
	if client == nil {
		return
	}

	// Keep connection alive and handle client disconnect
	ctx := r.Context()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.removeClient(client.ID)
				return
			case <-client.Done:
				return
			case <-ticker.C:
				// Send keepalive
				keepaliveEvent := SSEEvent{
					ID:        fmt.Sprintf("keepalive_%d", time.Now().UnixNano()),
					Event:     "keepalive",
					Data:      map[string]interface{}{"timestamp": time.Now()},
					Timestamp: time.Now(),
				}
				if err := s.sendEventToClient(client, keepaliveEvent); err != nil {
					s.removeClient(client.ID)
					return
				}
			}
		}
	}()

	// Wait for client to disconnect
	<-client.Done
}

// HandleScrapeSSE handles scrape requests via SSE
func (s *MCPSSEServer) HandleScrapeSSE(w http.ResponseWriter, r *http.Request) {
	var request struct {
		URL      string `json:"url"`
		Selector string `json:"selector"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.URL == "" || request.Selector == "" {
		http.Error(w, "url and selector are required", http.StatusBadRequest)
		return
	}

	// Create a temporary client for this request
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send start event
	startEvent := SSEEvent{
		ID:        fmt.Sprintf("scrape_start_%d", time.Now().UnixNano()),
		Event:     "scrape_start",
		Data:      map[string]string{"url": request.URL, "selector": request.Selector},
		Timestamp: time.Now(),
	}

	startJSON, _ := json.Marshal(startEvent)
	fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", startEvent.ID, startEvent.Event, string(startJSON))
	flusher.Flush()

	// Execute scrape in a goroutine
	go func() {
		ctx := context.Background()

		// Call the scrape function
		summary, markdown, err := scrape.Scrape(ctx, s.httpClient, request.URL, request.Selector)

		if err != nil {
			errorEvent := SSEEvent{
				ID:        fmt.Sprintf("scrape_error_%d", time.Now().UnixNano()),
				Event:     "scrape_error",
				Data:      map[string]string{"error": err.Error()},
				Timestamp: time.Now(),
			}
			errorJSON, _ := json.Marshal(errorEvent)
			fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", errorEvent.ID, errorEvent.Event, string(errorJSON))
			flusher.Flush()
			return
		}

		// Send result event
		resultEvent := SSEEvent{
			ID:    fmt.Sprintf("scrape_result_%d", time.Now().UnixNano()),
			Event: "scrape_result",
			Data: map[string]interface{}{
				"summary":  summary,
				"markdown": string(markdown),
			},
			Timestamp: time.Now(),
		}
		resultJSON, _ := json.Marshal(resultEvent)
		fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", resultEvent.ID, resultEvent.Event, string(resultJSON))
		flusher.Flush()

		// Send completion event
		completeEvent := SSEEvent{
			ID:        fmt.Sprintf("scrape_complete_%d", time.Now().UnixNano()),
			Event:     "scrape_complete",
			Data:      map[string]string{"status": "completed"},
			Timestamp: time.Now(),
		}
		completeJSON, _ := json.Marshal(completeEvent)
		fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", completeEvent.ID, completeEvent.Event, string(completeJSON))
		flusher.Flush()
	}()
}

// HandleGetDocumentSSE handles getDocument requests via SSE
func (s *MCPSSEServer) HandleGetDocumentSSE(w http.ResponseWriter, r *http.Request) {
	if s.service == nil {
		http.Error(w, "Document service not available", http.StatusServiceUnavailable)
		return
	}

	var request struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}

	// Create a temporary client for this request
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send start event
	startEvent := SSEEvent{
		ID:        fmt.Sprintf("document_start_%d", time.Now().UnixNano()),
		Event:     "document_start",
		Data:      map[string]string{"path": request.Path},
		Timestamp: time.Now(),
	}

	startJSON, _ := json.Marshal(startEvent)
	fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", startEvent.ID, startEvent.Event, string(startJSON))
	flusher.Flush()

	// Execute getDocument in a goroutine
	go func() {
		ctx := context.Background()

		// Create a request for the service
		req, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
		if err != nil {
			errorEvent := SSEEvent{
				ID:        fmt.Sprintf("document_error_%d", time.Now().UnixNano()),
				Event:     "document_error",
				Data:      map[string]string{"error": fmt.Sprintf("failed to create request: %v", err)},
				Timestamp: time.Now(),
			}
			errorJSON, _ := json.Marshal(errorEvent)
			fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", errorEvent.ID, errorEvent.Event, string(errorJSON))
			flusher.Flush()
			return
		}

		// Call the service to get the document
		document, err := s.service.GetDocument(nil, req, request.Path)

		if err != nil {
			errorEvent := SSEEvent{
				ID:        fmt.Sprintf("document_error_%d", time.Now().UnixNano()),
				Event:     "document_error",
				Data:      map[string]string{"error": err.Error()},
				Timestamp: time.Now(),
			}
			errorJSON, _ := json.Marshal(errorEvent)
			fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", errorEvent.ID, errorEvent.Event, string(errorJSON))
			flusher.Flush()
			return
		}

		// Send result event
		resultEvent := SSEEvent{
			ID:    fmt.Sprintf("document_result_%d", time.Now().UnixNano()),
			Event: "document_result",
			Data: map[string]interface{}{
				"document": document,
			},
			Timestamp: time.Now(),
		}
		resultJSON, _ := json.Marshal(resultEvent)
		fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", resultEvent.ID, resultEvent.Event, string(resultJSON))
		flusher.Flush()

		// Send completion event
		completeEvent := SSEEvent{
			ID:        fmt.Sprintf("document_complete_%d", time.Now().UnixNano()),
			Event:     "document_complete",
			Data:      map[string]string{"status": "completed"},
			Timestamp: time.Now(),
		}
		completeJSON, _ := json.Marshal(completeEvent)
		fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", completeEvent.ID, completeEvent.Event, string(completeJSON))
		flusher.Flush()
	}()
}

// GetConnectedClients returns information about connected clients
func (s *MCPSSEServer) GetConnectedClients() []map[string]interface{} {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	clients := make([]map[string]interface{}, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, map[string]interface{}{
			"id":        client.ID,
			"lastSeen":  client.LastSeen,
			"connected": time.Since(client.LastSeen) < 60*time.Second,
		})
	}
	return clients
}

// GetStats returns server statistics
func (s *MCPSSEServer) GetStats() map[string]interface{} {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	return map[string]interface{}{
		"connectedClients": len(s.clients),
		"bufferSize":       len(s.broadcast),
		"serverVersion":    Version,
	}
}

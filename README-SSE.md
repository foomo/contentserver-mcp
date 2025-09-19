# MCP Server with Server-Sent Events (SSE) Support

This package extends the MCP (Model Context Protocol) server to support Server-Sent Events (SSE), enabling real-time communication with clients.

## Overview

The `github.com/foomo/contentserver-mcp/mcp` package now provides both traditional MCP HTTP transport and SSE transport capabilities, allowing clients to receive real-time updates and stream MCP tool execution results.

## Architecture

### Core Components

1. **MCPSSEServer**: The core SSE server that manages client connections and event broadcasting
2. **McpHTTPSSEServer**: A combined HTTP/SSE server that serves both traditional MCP and SSE endpoints
3. **SSE Configuration**: Configurable parameters for keepalive, buffer sizes, and timeouts

### Package Structure

```
mcp/
├── handler.go      # MCP tool handlers (scrape, getDocument)
├── httpserver.go   # HTTP server with SSE integration
└── sse.go          # SSE server implementation
```

## Usage

### Basic Setup

```go
package main

import (
    "github.com/foomo/contentserver-mcp/mcp"
    "github.com/foomo/contentserver-mcp/service"
)

func main() {
    // Create your service instance
    serviceInstance := service.NewService(...)

    // Create MCP server
    mcpServer := mcp.NewServer(nil, serviceInstance)

    // Create HTTP server with SSE capabilities
    httpSSEServer := mcp.NewMcpHTTPSSEServer(
        logger,
        mcpServer,
        serviceInstance,
        httpClient,
        "/services/mcp",
        mcp.DefaultSSEServerConfig(),
    )

    // Use httpSSEServer as your HTTP handler
    http.Handle("/", httpSSEServer)
}
```

### Configuration

```go
config := &mcp.SSEServerConfig{
    KeepaliveInterval: 30 * time.Second,
    BufferSize:        100,
    ClientTimeout:     60 * time.Second,
}

server := mcp.NewMcpHTTPSSEServer(logger, mcpServer, service, httpClient, "/mcp", config)
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/services/mcp` | POST | Traditional MCP HTTP endpoint |
| `/services/mcp/sse` | GET | SSE connection endpoint for real-time events |
| `/services/mcp/sse/scrape` | POST | SSE-enabled scrape endpoint |
| `/services/mcp/sse/document` | POST | SSE-enabled document endpoint |
| `/services/mcp/sse/clients` | GET | Get information about connected SSE clients |
| `/services/mcp/sse/stats` | GET | Get server statistics |

## SSE Event Types

### Connection Events
- `connected`: Sent when a client successfully connects
- `keepalive`: Sent every 30 seconds to keep connections alive

### Scrape Events
- `scrape_start`: Sent when a scrape operation begins
- `scrape_result`: Sent with the scrape results (summary and markdown)
- `scrape_error`: Sent if a scrape operation fails
- `scrape_complete`: Sent when a scrape operation finishes

### Document Events
- `document_start`: Sent when a document request begins
- `document_result`: Sent with the document data
- `document_error`: Sent if a document request fails
- `document_complete`: Sent when a document request finishes

## Client Integration

### JavaScript Example

```javascript
// Connect to SSE stream
const eventSource = new EventSource('http://localhost:8000/services/mcp/sse');

// Listen for events
eventSource.addEventListener('connected', function(event) {
    console.log('Connected:', JSON.parse(event.data));
});

eventSource.addEventListener('scrape_result', function(event) {
    const result = JSON.parse(event.data);
    console.log('Scrape result:', result);
});

// Execute scrape via SSE
fetch('http://localhost:8000/services/mcp/sse/scrape', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        url: 'https://example.com',
        selector: 'main'
    })
});

// Execute document request via SSE
fetch('http://localhost:8000/services/mcp/sse/document', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        path: '/messer/kuechenmesser'
    })
});
```

### Go Client Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

func main() {
    // Connect to SSE stream
    req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://localhost:8000/services/mcp/sse", nil)
    resp, _ := http.DefaultClient.Do(req)

    // Read SSE events
    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "data: ") {
            var event SSEEvent
            json.Unmarshal([]byte(line[6:]), &event)
            fmt.Printf("Event: %s, Data: %v\n", event.Event, event.Data)
        }
    }
}
```

## Event Format

All SSE events follow this structure:

```
id: <unique_event_id>
event: <event_type>
data: <json_data>

```

The JSON data contains:
```json
{
    "id": "event_id",
    "event": "event_type",
    "data": { /* event-specific data */ },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

## Configuration Options

### SSEServerConfig

```go
type SSEServerConfig struct {
    KeepaliveInterval time.Duration // How often to send keepalive events
    BufferSize        int           // Size of the broadcast channel buffer
    ClientTimeout     time.Duration // When to consider clients disconnected
}
```

### Default Configuration

```go
func DefaultSSEServerConfig() *SSEServerConfig {
    return &SSEServerConfig{
        KeepaliveInterval: 30 * time.Second,
        BufferSize:        100,
        ClientTimeout:     60 * time.Second,
    }
}
```

## Advanced Usage

### Custom Event Broadcasting

```go
// Get the SSE server instance
sseServer := httpSSEServer.GetSSEServer()

// Broadcast custom events
sseServer.broadcastEvent(mcp.SSEEvent{
    ID:        "custom_event_123",
    Event:     "custom_event",
    Data:      map[string]string{"message": "Hello from server"},
    Timestamp: time.Now(),
})
```

### Client Management

```go
// Get connected clients
clients := sseServer.GetConnectedClients()

// Get server statistics
stats := sseServer.GetStats()
```

## Testing

### Using the Test Client

1. Start your MCP server
2. Open `test-sse-client.html` in a web browser
3. Connect to the SSE stream
4. Test scrape and document operations
5. Monitor events in the event log

### Manual Testing with curl

```bash
# Connect to SSE stream
curl -N -H "Accept: text/event-stream" \
  http://localhost:8000/services/mcp/sse

# Test scrape endpoint
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","selector":"main"}' \
  http://localhost:8000/services/mcp/sse/scrape

# Test document endpoint
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"path":"/messer/kuechenmesser"}' \
  http://localhost:8000/services/mcp/sse/document

# Check connected clients
curl http://localhost:8000/services/mcp/sse/clients

# Get server stats
curl http://localhost:8000/services/mcp/sse/stats
```

## Features

### Client Management
- Automatic client registration and cleanup
- Connection status tracking
- Keepalive mechanism to detect disconnected clients
- Client information API

### Error Handling
- Graceful handling of client disconnections
- Error events for failed operations
- Automatic cleanup of dead connections

### Performance
- Efficient broadcasting to multiple clients
- Non-blocking event delivery
- Configurable channel buffer sizes

### Security
- CORS headers for cross-origin requests
- Input validation on all endpoints
- Context-aware request handling

## Limitations

1. **Unidirectional**: SSE only allows server-to-client communication
2. **Browser connections**: Limited to 6 concurrent SSE connections per browser
3. **No message replay**: Events are not stored for late-connecting clients
4. **Memory usage**: All connected clients are kept in memory

## Best Practices

1. **Connection Management**: Monitor client connections and clean up disconnected clients
2. **Error Handling**: Always handle connection errors and implement reconnection logic
3. **Rate Limiting**: Consider implementing rate limiting for expensive operations
4. **Authentication**: Add authentication for production use
5. **Monitoring**: Monitor server statistics and client counts

## Troubleshooting

### Common Issues

1. **Connection fails**: Check server URL and CORS settings
2. **No events received**: Verify the server is running and check browser console
3. **Memory leaks**: Monitor server memory usage with many connected clients
4. **Event ordering**: Events may arrive out of order; use timestamps for ordering

### Debug Mode

Enable debug logging to see detailed information about client connections and event broadcasting.

## Contributing

When extending the SSE functionality:

1. Add new event types to the SSE event system
2. Update the test client to support new features
3. Add appropriate error handling
4. Update documentation
5. Add tests for new functionality

## License

This package is part of the foomo/contentserver-mcp project.

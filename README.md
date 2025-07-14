# Content Scraper MCP

A Model Context Protocol (MCP) server that provides a tool to scrape web content and convert it to markdown format.

## Features

- **Web Scraping**: Download HTML content from any URL
- **CSS Selector Support**: Extract specific content using CSS selectors
- **Markdown Conversion**: Convert HTML content to markdown format
- **Multiple Transport Modes**: Support for both stdio and HTTP interfaces
- **Type Safety**: Strongly typed tool handlers

## Installation

```bash
go build .
```

## Package Structure

The project is organized into reusable packages:

- `mcp/` - MCP server implementation with scrape tool
- `scrape/` - Web scraping functionality
- `service/` - Service layer and value objects

## Usage

### Stdio Mode (Default)

Run the server in stdio mode for integration with MCP clients:

```bash
./contentserver-mcp
# or explicitly
./contentserver-mcp --stdio
```

### HTTP Mode

Run the server as an HTTP service for remote access:

```bash
./contentserver-mcp --http :8080
```

This will start the MCP server on `http://localhost:8080/mcp`.

## Tool: scrape

The server provides a `scrape` tool with the following parameters:

- **url** (required): The URL of the webpage to scrape
- **selector** (required): CSS selector to extract specific content

### Supported Selectors

- **ID selectors**: `#content`, `#main`
- **Class selectors**: `.article`, `.content`
- **Tag selectors**: `article`, `div`, `main`

### Example Usage

#### Via MCP Client

```json
{
  "name": "scrape",
  "arguments": {
    "url": "https://example.com",
    "selector": "#content"
  }
}
```

#### Via HTTP API

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "scrape",
      "arguments": {
        "url": "https://example.com",
        "selector": "#content"
      }
    }
  }'
```

### Response Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"markdown\":\"# Extracted Content\\n\\nThis is the markdown content...\"}"
      }
    ]
  }
}
```

## Error Handling

The tool provides detailed error messages for common issues:

- Missing required parameters
- Invalid URLs
- Network errors
- HTML parsing errors
- Selector not found

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o contentserver-mcp .
```

## Reusing the MCP Package

The MCP package can be reused in other projects. See the `examples/` directory for complete examples:

### Simple HTTP Server
```go
package main

import (
    "flag"
    "log"
    "github.com/foomo/contentserver-mcp/mcp"
    "github.com/mark3labs/mcp-go/server"
)

func main() {
    port := flag.String("port", "8080", "HTTP server port")
    flag.Parse()
    
    s := mcp.NewServer()
    httpServer := server.NewStreamableHTTPServer(s)
    
    log.Printf("Starting MCP HTTP server on port %s...", *port)
    if err := httpServer.Start(":" + *port); err != nil {
        log.Fatal(err)
    }
}
```

### Advanced HTTP Server with Middleware
```go
package main

import (
    "net/http"
    "github.com/foomo/contentserver-mcp/mcp"
    "github.com/mark3labs/mcp-go/server"
)

func main() {
    s := mcp.NewServer()
    httpServer := server.NewStreamableHTTPServer(s)
    
    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthCheckHandler)
    mux.Handle("/mcp", httpServer)
    
    // Add custom middleware
    handler := loggingMiddleware(corsMiddleware(mux))
    
    http.ListenAndServe(":8080", handler)
}
```

See `examples/` directory for complete working examples.

## License

This project is part of the foomo content server ecosystem. 
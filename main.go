package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/foomo/contentserver-mcp/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Define command line flags
	stdioMode := flag.Bool("stdio", true, "Run in stdio mode")
	httpAddr := flag.String("http", "", "HTTP server address (e.g., ':8080')")
	flag.Parse()

	// Create MCP server using the extracted package
	// Note: Service parameter is nil, so only the scrape tool will be available
	// To enable the getDocument tool, create a service instance and pass it here
	s := mcp.NewServer(http.DefaultClient, nil)

	if *httpAddr != "" {
		// Start the HTTP server
		log.Printf("Starting MCP server on HTTP address: %s", *httpAddr)
		httpServer := server.NewStreamableHTTPServer(s)
		if err := httpServer.Start(*httpAddr); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}
	// Start the stdio server
	if *stdioMode {
		log.Println("Starting MCP server in stdio mode...")
	} else {
		log.Println("Starting MCP server in stdio mode (default)...")
	}
	if err := server.ServeStdio(s); err != nil {
		log.Fatal(err)
	}
}

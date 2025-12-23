package main

import (
	"flag"
	"log"
	"os"

	"github.com/Prodro21/video-mcp/internal/client"
	"github.com/Prodro21/video-mcp/internal/handlers"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Configuration flags
	apiURL := flag.String("api-url", "http://localhost:8080", "Video platform API base URL")
	flag.Parse()

	// Check for environment variable override
	if envURL := os.Getenv("VIDEO_PLATFORM_URL"); envURL != "" {
		*apiURL = envURL
	}

	// Create API client
	apiClient := client.New(*apiURL)

	// Create MCP server
	s := server.NewMCPServer(
		"video-platform",
		"1.0.0",
		server.WithResourceCapabilities(true, false),
		server.WithPromptCapabilities(true),
	)

	// Register handlers
	handlers.RegisterTools(s, apiClient)
	handlers.RegisterResources(s, apiClient)
	handlers.RegisterPrompts(s)

	// Start stdio server
	log.Println("Starting video-platform MCP server...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hermespage/hermespage/internal/config"
	"github.com/hermespage/hermespage/internal/handler"
	"github.com/hermespage/hermespage/internal/mcpserver"
	"github.com/hermespage/hermespage/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: hermespage <command>")
		fmt.Println("Commands:")
		fmt.Println("  serve   Start the web server")
		fmt.Println("  mcp     Start the MCP server (stdio)")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		runServe()
	case "mcp":
		runMCP()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runServe() {
	cfg := config.Load()

	if cfg.APIKey == "" {
		log.Fatal("HERMES_API_KEY environment variable is required")
	}

	store, err := storage.New(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	mux := http.NewServeMux()
	h := handler.New(store, cfg)
	h.RegisterRoutes(mux)

	addr := ":" + cfg.Port
	log.Printf("HermesPage server starting on %s", addr)
	log.Printf("  Data dir: %s", cfg.DataDir)
	log.Printf("  Web dir:  %s", cfg.WebDir)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func runMCP() {
	cfg := config.Load()
	serverURL := os.Getenv("HERMES_SERVER_URL")
	if serverURL == "" {
		serverURL = fmt.Sprintf("http://localhost:%s", cfg.Port)
	}
	apiKey := cfg.APIKey

	if err := mcpserver.Run(serverURL, apiKey); err != nil {
		log.Fatalf("MCP server failed: %v", err)
	}
}

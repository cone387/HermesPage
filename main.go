package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hermespage/hermespage/internal/auth"
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

	store, err := storage.New(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	users, err := auth.NewUserStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize user store: %v", err)
	}

	jwtSvc := auth.NewJWTService(cfg.JWTSecret)

	// auto-create admin from env if no users exist
	if !users.HasUsers() && cfg.AdminUser != "" && cfg.AdminPass != "" {
		user, err := users.CreateUser(cfg.AdminUser, cfg.AdminPass, "admin")
		if err != nil {
			log.Fatalf("Failed to create admin from env: %v", err)
		}
		log.Printf("Created admin user '%s' from environment (token: %s)", user.Username, user.Token)
	}

	serverURL := fmt.Sprintf("http://localhost:%s", cfg.Port)
	mcpHandler := mcpserver.NewHTTPHandler(serverURL)

	mux := http.NewServeMux()
	h := handler.New(store, users, jwtSvc, cfg, mcpHandler)
	h.RegisterRoutes(mux)

	addr := ":" + cfg.Port
	if !users.HasUsers() {
		log.Printf("No users found - setup mode active at http://localhost%s", addr)
	}
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
	token := os.Getenv("HERMES_TOKEN")

	if err := mcpserver.Run(serverURL, token); err != nil {
		log.Fatalf("MCP server failed: %v", err)
	}
}

// Command proxmox-mcp is an MCP server that exposes Proxmox VE cluster
// operations as MCP tools.
//
// Required environment variables:
//
//	PROXMOX_API_URL      Base URL of the Proxmox JSON API (e.g. https://pve:8006/api2/json)
//	PROXMOX_TOKEN_ID     API token ID (e.g. root@pam!mcp)
//	PROXMOX_TOKEN_SECRET API token UUID secret
//
// Optional environment variables:
//
//	PROXMOX_INSECURE          Set to "true" to skip TLS certificate verification
//	PROXMOX_ALLOW_DESTRUCTIVE Set to "true" to enable delete_vm and delete_container tools
//
// Flags:
//
//	--transport   Transport to use: "stdio" (default) or "http"
//	--addr        Listen address when --transport=http (default: localhost:8080)
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/gordcurrie/proxmox-mcp/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

// run contains all application logic and returns an error on failure.
// Keeping it separate from main allows defers to execute normally and
// makes the entrypoint independently testable.
func run() error {
	transport := flag.String("transport", "stdio", "Transport: stdio or http")
	addr := flag.String("addr", "localhost:8080", "Listen address (http transport only)")
	flag.Parse()

	apiURL, err := requireEnv("PROXMOX_API_URL")
	if err != nil {
		return err
	}
	tokenID, err := requireEnv("PROXMOX_TOKEN_ID")
	if err != nil {
		return err
	}
	tokenSecret, err := requireEnv("PROXMOX_TOKEN_SECRET")
	if err != nil {
		return err
	}
	insecure := os.Getenv("PROXMOX_INSECURE") == "true"
	allowDestructive := os.Getenv("PROXMOX_ALLOW_DESTRUCTIVE") == "true"

	client, err := proxmox.NewClient(apiURL, tokenID, tokenSecret, insecure)
	if err != nil {
		return fmt.Errorf("creating Proxmox client: %w", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "proxmox-mcp",
		Version: "v0.1.0",
	}, nil)

	tools.RegisterAll(server, client, tools.Config{AllowDestructive: allowDestructive})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	switch *transport {
	case "stdio":
		if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			return fmt.Errorf("stdio server: %w", err)
		}
	case "http":
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)
		httpServer := &http.Server{
			Addr:              *addr,
			Handler:           handler,
			ReadHeaderTimeout: 30 * time.Second,
		}
		slog.Info("proxmox-mcp listening", "addr", *addr, "transport", "http")
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if shutdownErr := httpServer.Shutdown(shutdownCtx); shutdownErr != nil {
				slog.Warn("HTTP server shutdown error", "err", shutdownErr)
			}
		}()
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("http server: %w", err)
		}
	default:
		return fmt.Errorf("unknown transport %q: must be 'stdio' or 'http'", *transport)
	}

	return nil
}

// requireEnv returns the value of the named environment variable or an error
// if it is unset or empty.
func requireEnv(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("required environment variable %s is not set", name)
	}
	return v, nil
}

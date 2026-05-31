package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/byte-v-forge/common-lib/envx"
)

type config struct {
	ListenAddr         string
	DashboardStaticDir string
	N8NInternalURL     string
	N8NPublicURL       string
	N8NAPIKey          string
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := loadConfig()
	server := &dashboardServer{
		staticDir:   cfg.DashboardStaticDir,
		projections: newWorkflowProjectionStore(),
		workflowRuntime: newWorkflowRuntimeClient(workflowRuntimeConfig{
			InternalURL: cfg.N8NInternalURL,
			EditorURL:   cfg.N8NPublicURL,
			APIKey:      cfg.N8NAPIKey,
		}),
	}

	httpServer := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           withCORS(server.routes()),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("workflow-runtime dashboard listening", "addr", cfg.ListenAddr)
	if err := httpServer.ListenAndServe(); err != nil {
		logger.Error("workflow-runtime stopped", "error", err)
		os.Exit(1)
	}
}

func loadConfig() config {
	return config{
		ListenAddr:         envx.StringDefault("WORKFLOW_RUNTIME_HTTP_ADDR", ":8080"),
		DashboardStaticDir: envx.StringDefault("WORKFLOW_RUNTIME_DASHBOARD_STATIC_DIR", "/app/dashboard/workflow-runtime"),
		N8NInternalURL:     envx.StringDefault("N8N_INTERNAL_URL", ""),
		N8NPublicURL:       envx.StringDefault("N8N_PUBLIC_URL", ""),
		N8NAPIKey:          envx.StringDefault("N8N_API_KEY", ""),
	}
}

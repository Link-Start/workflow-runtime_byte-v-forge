package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"google.golang.org/protobuf/proto"
)

type dashboardServer struct {
	staticDir       string
	workflowRuntime *workflowRuntimeClient
}

func (s *dashboardServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/workflow-runtime/summary", s.handleWorkflowRuntimeSummary)
	mux.Handle("/mf/workflow-runtime/", http.StripPrefix("/mf/workflow-runtime/", noCacheFileServer(s.staticDir)))
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/health", s.handleHealth)
	return mux
}

func (s *dashboardServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *dashboardServer) handleWorkflowRuntimeSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeProtoJSON(w, http.StatusOK, s.workflowRuntime.summary(r.Context()))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,PUT,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func noCacheFileServer(dir string) http.Handler {
	root := filepath.Clean(dir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		cleanPath := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		target := filepath.Join(root, filepath.FromSlash(cleanPath))
		if !isPathWithin(root, target) {
			http.NotFound(w, r)
			return
		}
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			http.ServeFile(w, r, target)
			return
		}
		http.NotFound(w, r)
	})
}

func isPathWithin(root string, target string) bool {
	rel, err := filepath.Rel(root, target)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, "../") && !filepath.IsAbs(rel)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeProtoJSON(w http.ResponseWriter, status int, value proto.Message) {
	_ = protojsonhttp.WriteResponse(w, status, value)
}

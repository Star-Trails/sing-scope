package app

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"sing-scope/internal/store"
)

// NewAssetHandler creates a unified HTTP handler serving both the embedded SPA and the Go AppService APIs.
func NewAssetHandler(svc *AppService, assets fs.FS) http.Handler {
	mux := http.NewServeMux()

	// API Endpoints
	mux.HandleFunc("/api/connection", func(w http.ResponseWriter, r *http.Request) {
		info := svc.GetConnectionInfo()
		writeJSON(w, info)
	})

	mux.HandleFunc("/api/connect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			URL    string `json:"url"`
			Secret string `json:"secret"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ok := svc.ConnectServer(body.URL, body.Secret)
		writeJSON(w, map[string]bool{"ok": ok})
	})

	mux.HandleFunc("/api/disconnect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ok := svc.DisconnectServer()
		writeJSON(w, map[string]bool{"ok": ok})
	})

	mux.HandleFunc("/api/overview", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		overview := svc.GetOverviewSummary(filter)
		writeJSON(w, overview)
	})

	mux.HandleFunc("/api/flows", func(w http.ResponseWriter, r *http.Request) {
		var opts store.QueryOptions
		if r.Method == http.MethodPost && r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&opts)
		}
		res := svc.GetFlows(opts)
		writeJSON(w, res)
	})

	mux.HandleFunc("/api/analytics", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		topNStr := r.URL.Query().Get("topN")
		topN := 100
		if n, err := strconv.Atoi(topNStr); err == nil && n > 0 {
			topN = n
		}
		res := svc.GetBatchAnalytics(filter, topN)
		writeJSON(w, res)
	})

	mux.HandleFunc("/api/rules", func(w http.ResponseWriter, r *http.Request) {
		rules := svc.GetRules()
		writeJSON(w, rules)
	})

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		st := svc.GetSystemStatus()
		writeJSON(w, st)
	})

	mux.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		limitStr := r.URL.Query().Get("limit")
		limit := 100
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
		logs := svc.GetLogs(limit)
		writeJSON(w, logs)
	})

	mux.HandleFunc("/api/clear-logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ok := svc.ClearLogs()
		writeJSON(w, map[string]bool{"ok": ok})
	})

	mux.HandleFunc("/api/groups", func(w http.ResponseWriter, r *http.Request) {
		groups := svc.GetGroups()
		writeJSON(w, groups)
	})

	mux.HandleFunc("/api/close", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Query().Get("id")
		ok := svc.CloseConnection(id)
		writeJSON(w, map[string]bool{"ok": ok})
	})

	mux.HandleFunc("/api/close-all", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ok := svc.CloseAllConnections()
		writeJSON(w, map[string]bool{"ok": ok})
	})

	mux.HandleFunc("/api/select-outbound", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			GroupTag    string `json:"groupTag"`
			OutboundTag string `json:"outboundTag"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ok := svc.SelectOutbound(body.GroupTag, body.OutboundTag)
		writeJSON(w, map[string]bool{"ok": ok})
	})

	mux.HandleFunc("/api/url-test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tag := r.URL.Query().Get("tag")
		ok := svc.URLTest(tag)
		writeJSON(w, map[string]bool{"ok": ok})
	})
	mux.HandleFunc("/api/probe-latency", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		latency := svc.ProbeLatency(target)
		writeJSON(w, map[string]int{"latency": latency})
	})


	mux.HandleFunc("/api/clash-mode", func(w http.ResponseWriter, r *http.Request) {
		st := svc.GetClashModeStatus()
		writeJSON(w, st)
	})

	mux.HandleFunc("/api/set-clash-mode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Mode string `json:"mode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ok := svc.SetClashMode(body.Mode)
		writeJSON(w, map[string]bool{"ok": ok})
	})

	mux.HandleFunc("/api/started-at", func(w http.ResponseWriter, r *http.Request) {
		ts := svc.GetStartedAt()
		writeJSON(w, map[string]int64{"startedAt": ts})
	})

	mux.HandleFunc("/api/filter", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		filter := r.URL.Query().Get("filter")
		ok := svc.SetInboundFilter(filter)
		writeJSON(w, map[string]bool{"ok": ok})
	})

	// Static Assets File Server
	fileServer := http.FileServer(http.FS(assets))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path != "" {
			if _, err := fs.Stat(assets, path); err != nil {
				// Fallback to index.html for SPA router paths
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	return mux
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_ = json.NewEncoder(w).Encode(data)
}

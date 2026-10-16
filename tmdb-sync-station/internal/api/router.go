package api

import (
	"net/http"
	"strings"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5173")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	switch r.URL.Path {
	case "/api/health":
		s.health(w, r)
	case "/api/dashboard":
		s.dashboard(w, r)
	case "/api/catalog":
		s.catalog(w, r)
	case "/api/catalog/detail":
		s.catalogDetail(w, r)
	case "/api/catalog/search":
		s.catalogSearch(w, r)
	case "/api/sync/runs":
		s.syncRuns(w, r)
	case "/api/sync/status":
		s.syncStatus(w, r)
	case "/api/sync/start":
		s.startSync(w, r)
	default:
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeError(w, http.StatusNotFound, "未找到接口")
			return
		}
		serveWeb(w, r, s.webRoot)
	}
}

package api

import (
	"context"
	"net/http"
	"time"
)

func (s *Server) catalogSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "仅支持 GET 请求")
		return
	}
	query := r.URL.Query()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	results, err := s.store.SearchCatalog(ctx, query.Get("kind"), query.Get("q"), query.Get("year"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

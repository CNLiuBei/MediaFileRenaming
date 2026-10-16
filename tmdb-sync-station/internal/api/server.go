package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"tmdb-sync-station/internal/catalog"
)

type SyncController interface {
	Start() error
	Configured() bool
}

type SyncSettings struct {
	APIBase        string
	QPS            float64
	AuthConfigured bool
	AutoSync       bool
	WebRoot        string
}

type Server struct {
	store          *catalog.Store
	syncController SyncController
	syncSettings   SyncSettings
	webRoot        string
}

func NewServer(store *catalog.Store, syncController SyncController, settings SyncSettings) *Server {
	return &Server{store: store, syncController: syncController, syncSettings: settings, webRoot: settings.WebRoot}
}

func (s *Server) catalogDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "仅支持 GET 请求")
		return
	}
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	kind := r.URL.Query().Get("kind")
	if err != nil || id < 1 || (kind != "movie" && kind != "tv") {
		writeError(w, http.StatusBadRequest, "影视类型或 TMDB ID 无效")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	item, err := s.store.CatalogDetail(ctx, kind, id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "片库中未找到该影视条目")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取影视详情失败: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	result, err := s.store.Health(ctx)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	result, err := s.store.Dashboard(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.decorateSyncStatus(&result.Sync)
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) catalog(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, hasMore, err := s.store.Catalog(ctx, r.URL.Query().Get("type"), r.URL.Query().Get("q"), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "hasMore": hasMore, "limit": limit, "offset": offset})
}

func (s *Server) syncRuns(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	runs, err := s.store.SyncRuns(ctx, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (s *Server) syncStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	status, err := s.store.SyncStatus(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.decorateSyncStatus(&status)
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) startSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "仅支持 POST 请求")
		return
	}
	if s.syncController == nil || !s.syncController.Configured() {
		writeError(w, http.StatusServiceUnavailable, "TMDB 同步尚未配置")
		return
	}
	if err := s.syncController.Start(); err != nil {
		if errors.Is(err, catalog.ErrSyncAlreadyRunning) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"state": "running", "message": "同步任务已启动"})
}

func (s *Server) decorateSyncStatus(status *catalog.SyncStatus) {
	status.Configured = s.syncController != nil && s.syncController.Configured()
	status.APIBase = s.syncSettings.APIBase
	status.RateLimit = s.syncSettings.QPS
	status.AuthConfigured = s.syncSettings.AuthConfigured
	status.AutoSync = s.syncSettings.AutoSync
	if !status.AutoSync && status.State != "running" {
		status.NextSync = "自动同步已关闭"
	} else if status.State == "failed" {
		status.NextSync = "等待下次自动同步"
	}
	if !status.Configured {
		status.Description = "请在服务端配置 TMDB API"
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

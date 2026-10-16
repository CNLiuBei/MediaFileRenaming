package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tmdb-sync-station/internal/api"
	"tmdb-sync-station/internal/catalog"
	"tmdb-sync-station/internal/syncer"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	address := os.Getenv("TMDB_STATION_ADDR")
	if address == "" {
		address = "127.0.0.1:8788"
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://localhost:5432/tmdb_sync_station?sslmode=disable"
	}
	webRoot := os.Getenv("TMDB_STATION_WEB_ROOT")
	if webRoot == "" {
		webRoot = "frontend/dist"
	}
	store, err := catalog.Open(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	if err := store.MarkInterruptedSyncRuns(ctx); err != nil {
		log.Fatal(err)
	}
	tmdbConfig, err := syncer.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	syncManager := syncer.NewManager(ctx, store, tmdbConfig)
	syncManager.Schedule()
	server := &http.Server{
		Addr: address,
		Handler: api.NewServer(store, syncManager, api.SyncSettings{
			APIBase: tmdbConfig.APIBase, QPS: tmdbConfig.QPS,
			AuthConfigured: tmdbConfig.APIKey != "" || tmdbConfig.ReadAccessToken != "",
			AutoSync:       tmdbConfig.AutoStart,
			WebRoot:        webRoot,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("关闭 HTTP 服务失败: %v", err)
		}
	}()
	log.Printf("TMDB 同步站 API 已连接 PostgreSQL 并启动：http://%s（TMDB 限速 %.2f QPS）", address, tmdbConfig.QPS)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP 服务已退出: %v", err)
	}
	stop()
	syncManager.Wait()
}

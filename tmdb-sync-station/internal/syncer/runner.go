package syncer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"tmdb-sync-station/internal/catalog"
)

var ErrAlreadyRunning = catalog.ErrSyncAlreadyRunning

type changeRef struct {
	kind string
	id   int64
}

type Manager struct {
	ctx         context.Context
	store       *catalog.Store
	client      *client
	config      Config
	mu          sync.Mutex
	running     bool
	wg          sync.WaitGroup
	schedulerWG sync.WaitGroup
}

func NewManager(ctx context.Context, store *catalog.Store, config Config) *Manager {
	return &Manager{ctx: ctx, store: store, client: newClient(config), config: config}
}

func (m *Manager) Configured() bool { return m.config.APIBase != "" }

func (m *Manager) Start() error {
	if !m.Configured() {
		return errors.New("未配置 TMDB API 地址")
	}
	m.mu.Lock()
	if m.ctx.Err() != nil {
		m.mu.Unlock()
		return m.ctx.Err()
	}
	if m.running {
		m.mu.Unlock()
		return ErrAlreadyRunning
	}
	m.running = true
	m.wg.Add(1)
	m.mu.Unlock()
	runID, err := m.store.BeginSyncRun(m.ctx)
	if err != nil {
		m.mu.Lock()
		m.running = false
		m.mu.Unlock()
		m.wg.Done()
		return fmt.Errorf("创建同步记录失败: %w", err)
	}
	go func() {
		defer m.wg.Done()
		defer func() { m.mu.Lock(); m.running = false; m.mu.Unlock() }()
		if err := m.run(m.ctx, runID); err != nil {
			log.Printf("TMDB 增量同步失败: %v", err)
		}
	}()
	return nil
}

func (m *Manager) Wait() { m.schedulerWG.Wait(); m.wg.Wait() }

func (m *Manager) Schedule() {
	if !m.config.AutoStart {
		return
	}
	m.schedulerWG.Add(1)
	go func() {
		defer m.schedulerWG.Done()
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-m.ctx.Done():
				return
			case <-timer.C:
				if err := m.Start(); err != nil && !errors.Is(err, ErrAlreadyRunning) {
					log.Printf("启动定时同步失败: %v", err)
				}
				timer.Reset(m.config.Interval)
			}
		}
	}()
}

func (m *Manager) run(ctx context.Context, runID int64) (runErr error) {
	var err error
	var processed, total int64
	var finishStage string
	defer func() {
		state, message := "success", ""
		if finishStage == "" {
			finishStage = "同步完成"
		}
		if runErr != nil {
			state, message, finishStage = "failed", runErr.Error(), "同步失败"
		}
		finishCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if finishErr := m.store.FinishSyncRun(finishCtx, runID, processed, total, state, message, finishStage); finishErr != nil {
			log.Printf("写入同步结束状态失败: %v", finishErr)
		}
	}()

	from, to, clipped, err := m.syncWindow(ctx)
	if err != nil {
		return err
	}
	if from.After(to) {
		finishStage = "没有新的完整日期需要同步"
		return nil
	}
	batch, err := m.store.GetOrCreateSyncBatch(ctx, from, to, clipped)
	if err != nil {
		return err
	}
	from, to, clipped = batch.WindowStart, batch.WindowEnd, batch.Clipped
	if !batch.ManifestComplete {
		batch, err = m.changedItems(ctx, runID, batch)
		if err != nil {
			return err
		}
	}
	if !batch.ManifestComplete {
		if err = m.store.CompleteSyncManifest(ctx, batch.ID); err != nil {
			return fmt.Errorf("提交变更清单检查点失败: %w", err)
		}
	}
	processed, total, err = m.store.SyncBatchItemCounts(ctx, batch.ID)
	if err != nil {
		return fmt.Errorf("统计待同步影视条目失败: %w", err)
	}
	if err = m.store.UpdateSyncRun(ctx, runID, processed, total, "读取变更详情"); err != nil {
		return err
	}
	items, err := m.store.PendingSyncBatchItems(ctx, batch.ID)
	if err != nil {
		return fmt.Errorf("读取待同步影视条目失败: %w", err)
	}
	for _, item := range items {
		ref := changeRef{kind: item.MediaType, id: item.TMDBID}
		if err = m.syncItem(ctx, batch.ID, ref, runID, processed, total); err != nil {
			return err
		}
		processed++
		if err = m.store.UpdateSyncRun(ctx, runID, processed, total, fmt.Sprintf("已处理 %d / %d 条", processed, total)); err != nil {
			return err
		}
	}
	if err = m.store.CompleteSyncBatch(ctx, batch.ID, to.AddDate(0, 0, -1)); err != nil {
		return fmt.Errorf("提交同步检查点和水位失败: %w", err)
	}
	if clipped {
		finishStage = "同步完成；TMDB 仅允许查询最近 14 天，已从窗口起点开始"
	}
	return nil
}

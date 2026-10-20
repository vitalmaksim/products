package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrSyncAlreadyRunning = errors.New("sync already running")

type SyncRunner interface {
	Sync(ctx context.Context) error
}

type SyncState string

const (
	SyncIdle      SyncState = "idle"
	SyncRunning   SyncState = "running"
	SyncCompleted SyncState = "completed"
	SyncFailed    SyncState = "failed"
)

type SyncStatus struct {
	Status     SyncState  `json:"status"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	LastError  string     `json:"last_error"`
}

type SyncManager struct {
	runner SyncRunner

	mu         sync.Mutex
	status     SyncState
	startedAt  time.Time
	finishedAt time.Time
	lastError  string
}

func NewSyncManager(runner SyncRunner) *SyncManager {
	return &SyncManager{runner: runner, status: SyncIdle}
}

func (m *SyncManager) Start(ctx context.Context) error {
	m.mu.Lock()
	// Проверка и смена состояния неделимы: иначе два запроса могут запустить работу одновременно.
	if m.status == SyncRunning {
		m.mu.Unlock()
		return ErrSyncAlreadyRunning
	}
	if err := ctx.Err(); err != nil {
		m.mu.Unlock()
		return err
	}

	m.status = SyncRunning
	m.startedAt = time.Now().UTC()
	m.finishedAt = time.Time{}
	m.lastError = ""
	m.mu.Unlock()

	go func() {
		// Сеть и БД работают без mutex, чтобы запросы статуса не ждали окончания синхронизации.
		err := m.runner.Sync(ctx)

		m.mu.Lock()
		defer m.mu.Unlock()

		m.finishedAt = time.Now().UTC()
		if err != nil {
			m.status = SyncFailed
			m.lastError = err.Error()
			return
		}
		m.status = SyncCompleted
	}()

	return nil
}

func (m *SyncManager) GetStatus() SyncStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	status := SyncStatus{Status: m.status, LastError: m.lastError}
	// Указатели ведут на копии времени: вызывающий не сможет изменить состояние менеджера.
	if !m.startedAt.IsZero() {
		startedAt := m.startedAt
		status.StartedAt = &startedAt
	}
	if !m.finishedAt.IsZero() {
		finishedAt := m.finishedAt
		status.FinishedAt = &finishedAt
	}
	return status
}

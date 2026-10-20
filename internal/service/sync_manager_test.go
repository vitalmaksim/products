package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeSyncRunner struct {
	started chan struct{}
	release chan struct{}
}

func (f *fakeSyncRunner) Sync(ctx context.Context) error {
	close(f.started)
	select {
	case <-f.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestSyncManager(t *testing.T) {
	runner := &fakeSyncRunner{started: make(chan struct{}), release: make(chan struct{})}
	manager := NewSyncManager(runner)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(func() {
		// Cleanup сработает и после Fatal, чтобы fake не оставил работающую горутину.
		cancel()
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := manager.Shutdown(shutdownCtx); err != nil {
			t.Errorf("cleanup sync manager: %v", err)
		}
	})

	if status := manager.GetStatus().Status; status != SyncIdle {
		t.Fatalf("initial status: got %q, want %q", status, SyncIdle)
	}
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("start sync: %v", err)
	}
	// Канал подтверждает вход в fake; время используется только как защита от зависания.
	select {
	case <-runner.started:
	case <-ctx.Done():
		t.Fatal("sync runner did not start before timeout")
	}
	if status := manager.GetStatus().Status; status != SyncRunning {
		t.Fatalf("running status: got %q, want %q", status, SyncRunning)
	}
	if err := manager.Start(ctx); !errors.Is(err, ErrSyncAlreadyRunning) {
		t.Fatalf("second start: got %v, want ErrSyncAlreadyRunning", err)
	}

	close(runner.release)
	// Fake уже может завершиться, но только Shutdown подтверждает и обновление статуса менеджером.
	if err := manager.Shutdown(ctx); err != nil {
		t.Fatalf("wait for sync completion: %v", err)
	}
	if status := manager.GetStatus().Status; status != SyncCompleted {
		t.Fatalf("completed status: got %q, want %q", status, SyncCompleted)
	}
}

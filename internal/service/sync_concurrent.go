package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/vitalmaksim/products/internal/supplier"
	"github.com/vitalmaksim/products/model"
	"golang.org/x/sync/errgroup"
)

const syncWorkerCount = 5

func (s *SyncService) SyncConcurrent(ctx context.Context) error {
	firstPage, err := s.firstPage(ctx)
	if err != nil {
		return err
	}
	batch, err := s.appendBatch(ctx, make([]model.Product, 0, syncBatchSize), firstPage.Items)
	if err != nil {
		return err
	}

	// Отмена нужна и агрегатору: ошибка записи должна остановить загрузку страниц.
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	// errgroup передает ошибку загрузчика и отменяет запросы остальных участников.
	group, groupCtx := errgroup.WithContext(runCtx)

	// Общие небуферизованные каналы не позволяют накопить все задания и страницы в памяти.
	jobs := make(chan int)
	results := make(chan supplier.ProductsResponse)

	group.Go(func() error {
		// Только источник отправляет задания, поэтому он же закрывает jobs.
		defer close(jobs)
		for page := 2; page <= firstPage.TotalPages; page++ {
			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
			case jobs <- page:
			}
		}
		return nil
	})

	var workers sync.WaitGroup
	workers.Add(syncWorkerCount)
	for i := 0; i < syncWorkerCount; i++ {
		group.Go(func() error {
			defer workers.Done()
			for {
				select {
				case <-groupCtx.Done():
					return groupCtx.Err()
				case page, ok := <-jobs:
					if !ok {
						return nil
					}

					response, err := s.productSupplier.GetProducts(groupCtx, page)
					if err != nil {
						return fmt.Errorf("get product page %d: %w", page, err)
					}

					// При ошибке БД агрегатор перестает читать: отмена освобождает отправителя.
					select {
					case <-groupCtx.Done():
						return groupCtx.Err()
					case results <- response:
					}
				}
			}
		})
	}

	group.Go(func() error {
		// Выход общий: закрывать его можно только после завершения всех отправителей.
		workers.Wait()
		close(results)
		return nil
	})

	// Только эта горутина владеет batch и пишет в БД, поэтому mutex не нужен.
	for response := range results {
		batch, err = s.appendBatch(groupCtx, batch, response.Items)
		if err != nil {
			alreadyCanceled := groupCtx.Err() != nil
			cancel()
			loadErr := group.Wait()
			// Если БД прервана из-за сбоя загрузчика, сохраняем исходную ошибку с номером страницы.
			if alreadyCanceled && loadErr != nil {
				return loadErr
			}
			return err
		}
	}

	// Wait вызывается после чтения или отмены: воркеры не останутся ждать получателя.
	if err := group.Wait(); err != nil {
		return err
	}

	// Даже успешный Wait отменяет groupCtx; остаток сохраняем с контекстом вызывающего.
	return s.saveBatch(ctx, batch)
}

package app

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	totalJobs   = 24
	workerCount = 4
)

type Result struct {
	ID    int
	Value int
}

// Имитирует внешнюю операцию.
func process(ctx context.Context, id int, failOn int) (Result, error) {
	delay := time.Duration(100+(id%4)*50) * time.Millisecond

	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-time.After(delay):
	}

	if failOn != 0 && id == failOn {
		return Result{}, fmt.Errorf("job %d failed", id)
	}

	return Result{ID: id, Value: id * id}, nil
}

// Запускает Worker Pool и возвращает результаты.
func run(ctx context.Context, totalJobs int, workerCount int, failOn int) ([]Result, error) {
	jobs := make(chan int)
	results := make(chan Result)

	go func() {
		defer close(jobs)
		for i := 1; i <= totalJobs; i++ {
			jobs <- i
		}
	}()

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				result, err := process(ctx, job, failOn)
				if err != nil {
					continue
				}

				results <- result
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	response := make([]Result, 0)
	for result := range results {
		response = append(response, result)
	}

	return response, nil
}

func main() {
	// TODO
}

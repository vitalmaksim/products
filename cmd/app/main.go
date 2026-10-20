package app

import (
	"context"
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
	// TODO
}

// Запускает Worker Pool и возвращает результаты.
func run(
	ctx context.Context,
	totalJobs int,
	workerCount int,
	failOn int,
) ([]Result, error) {
	// TODO
}

func main() {
	// TODO
}

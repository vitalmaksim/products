package service

import (
	"context"
	"fmt"
)

type UserBalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (int64, error)
}

type BalanceService struct {
	userRepository UserBalanceRepository
}

func NewBalanceService(userRepository UserBalanceRepository) *BalanceService {
	return &BalanceService{userRepository: userRepository}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID int64) (int64, error) {
	balance, err := s.userRepository.GetBalance(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("get balance %w", err)
	}

	return balance, nil
}

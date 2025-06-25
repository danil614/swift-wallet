package service

import (
	"context"

	"github.com/google/uuid"

	"swiftwallet/internal/model"
	"swiftwallet/internal/repository"
)

type Service interface {
	Operate(ctx context.Context, id uuid.UUID, op model.OperationType, amount int64) (int64, error)
	Balance(ctx context.Context, id uuid.UUID) (int64, error)
}

type service struct {
	repo repository.Repository
}

func New(r repository.Repository) Service { return &service{repo: r} }

func (s *service) Operate(ctx context.Context, id uuid.UUID, op model.OperationType, amount int64) (int64, error) {
	return s.repo.ChangeBalance(ctx, id, op, amount)
}

func (s *service) Balance(ctx context.Context, id uuid.UUID) (int64, error) {
	return s.repo.GetBalance(ctx, id)
}

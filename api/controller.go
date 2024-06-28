package api

import (
	"context"
	"kiln-tezos-delegation/repository"
)

const YearNotSpecified = 0

type Repository interface {
	GetDelegations(context.Context) ([]repository.Delegation, error)
	GetDelegationsOfYear(context.Context, int) ([]repository.Delegation, error)
}

type TezosController struct {
	repo Repository
}

func NewController(repo Repository) TezosController {
	return TezosController{
		repo: repo,
	}
}

func (c TezosController) GetDelegations(ctx context.Context, year int) ([]repository.Delegation, error) {
	if year != YearNotSpecified {
		return c.repo.GetDelegationsOfYear(ctx, year)
	}
	return c.repo.GetDelegations(ctx)
}

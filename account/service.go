package account

import (
	"context"
	"github.com/segmentio/ksuid"
)

type Service interface {
	PostAccount(ctx context.Context, name string) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	ListAccounts(ctx context.Context, skip uint, take uint64) ([]Account, error)
}

type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type accountService struct {
	repository Repository
}

func newService(r Repository) Service {
	return &accountService{
		repository: r,
	}
}

func (s *accountService) PostAccount(ctx context.Context, name string) (*Account, error) {
	a := Account{
		Name: name,
		ID:   ksuid.New().String(),
	}

	if err := s.repository.PutAccount(ctx, *a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *accountService) GetAccountByID(ctx context.Context, id string) (*Account, error) {

	return s.repository.GetAccountByID(ctx, id)
}

func (s *accountService) GetAccounts(ctx context.Context, skip uint, take uint64) ([]Account, error) {

	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	return s.repository.ListAccounts(ctx, skip, take)
}
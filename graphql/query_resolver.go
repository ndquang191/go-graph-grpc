package main

import (
	"context"
	"log"
	"time"
)

type queryResolver struct {
	server *Server
}

func (r *queryResolver) Account(ctx context.Context, pagination *PaginationInput, id *string) ([]*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if id != nil {
		r, err := r.server.accountClient.GetAccount(ctx, *id)

		if err != nil {
			log.Print(err)
			return nil, err
		}

		return []*Account{{
			ID:   r.ID,
			Name: r.Name,
		}}, nil
	}

	skip, take := uint64(0), uint64(0)
	if pagination.Skip != nil && pagination.Take != nil {
		skip, take = pagination.bounds()
	}

	accountList, err := r.server.accountClient.GetAccounts(ctx, skip, take)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	var accounts []*Account

	for _, a := range accountList {
		account := &Account{
			ID:   a.ID,
			Name: a.Name,
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *queryResolver) Products(ctx context.Context, pagination *PaginationInput, query *string, id *string) ([]*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)

	defer cancel()

	if id != nil {
		r, err := r.server.catalogClient.GetProduct(ctx, *id)

		if err != nil {
			log.Print(err)
			return nil, err
		}

		return []*Product{{
			ID:          r.Id,
			Name:        r.Name,
			Description: r.Description,
			Price:       r.Price,
		}}, nil
	}

	skip, take := uint64(0), uint64(0)
	if pagination.Skip != nil && pagination.Take != nil {
		skip, take = pagination.bounds()
	}

	q := ""
	if query != nil {
		q = *query
	}

	productList, err := r.server.catalogClient.GetProducts(ctx, skip, take, nil, q)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var products []*Product

	for _, a := range productList {
		product := &Product{
			ID:          a.ID,
			Name:        a.Name,
			Description: a.Description,
			Price:       a.Price,
		}
		products = append(products, product)
	}

	return products, nil
}

func (p PaginationInput) bounds() (uint64, uint64) {
	skipV := uint64(0)
	takeV := uint64(0)

	if p.Skip != nil {
		skipV = uint64(*p.Skip)
	}
	if p.Take != nil {
		takeV = uint64(*p.Take)
	}

	return skipV, takeV
}

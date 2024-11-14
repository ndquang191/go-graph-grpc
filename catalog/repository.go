package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	elastic "github.com/olivere/elastic/v7"
)

var (
	ErrNotFound = errors.New("Entity not found")
)

type Repository interface {
	Close()
	PutProduct(ctx context.Context, product *Product) (*Product, error)
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type elasticRepository struct {
	client *elastic.Client
}

type productDocument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
}

func NewElasticRepository(url string) (Repository, error) {
	client, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(false))

	if err != nil {
		return nil, err
	}

	return &elasticRepository(client), nil
}

func (r *elasticRepository) Close() {
	r.client.Stop()
}

func (r *elasticRepository) PutProduct(ctx context.Context, product *Product) (*Product, error) {
	_, err := r.client.Index().
		Index("catalog").
		Type("product").
		Id(product.ID).
		BodyJson(productDocument{
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		})

	Do(ctx)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *elasticRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	res, err := r.client.Get().
		Index("catalog").
		Type("product").
		Id(id).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	if !res.Found {
		return nil, ErrNotFound
	}

	P := productDocument{}
	if err = json.Unmarshal(*res.Source, &P); err != nil {
		return nil, err
	}

	return &Product{
		ID:          res.Id,
		Name:        P.Name,
		Description: P.Description,
		Price:       P.Price,
	}, nil
}

func (r *elasticRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	res, err := r.client.Search().
		Index("catalog").
		Type("product").Query(elastic.NewMatchAllQuery()).
		From(int(skip)).Size(int(take)).
		Do(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	products := []Product{}

	for _, hit := range res.Hits.Hits {
		P := productDocument{}
		if err := json.Unmarshal(*hit.Source , &p ); err == nil {
			
		}


}

func (r *elasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
}

func (r *elasticRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
}

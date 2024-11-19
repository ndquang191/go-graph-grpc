package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	elastic "github.com/elastic/go-elasticsearch/v8"
)

var (
	ErrNotFound = errors.New("Entity not found")
)

type Repository interface {
	Close()
	PutProduct(ctx context.Context, product *Product) error
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type elasticRepository struct {
	client *elastic.Client
}

type productDocument struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

// NewElasticRepository initializes the repository with Elasticsearch v8
func NewElasticRepository(url string) (Repository, error) {
	client, err := elastic.NewClient(elastic.Config{
		Addresses: []string{url},
	})
	if err != nil {
		return nil, err
	}

	return &elasticRepository{client: client}, nil
}

// Close cleans up resources (no explicit close needed for go-elasticsearch)
func (r *elasticRepository) Close() {
	// No specific method for closing in go-elasticsearch
}

// PutProduct indexes a product document
func (r *elasticRepository) PutProduct(ctx context.Context, product *Product) error {
	doc := productDocument{
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	res, err := r.client.Index(
		"products",            // Index name
		bytes.NewReader(body), // Document body
		r.client.Index.WithContext(ctx),
		r.client.Index.WithDocumentID(product.ID), // Optional: specify document ID
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.New("error indexing product")
	}

	return nil
}

// GetProductByID retrieves a product by its ID
func (r *elasticRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	res, err := r.client.Get(
		"products", // Index name
		id,         // Document ID
		r.client.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, ErrNotFound
		}
		return nil, errors.New("error retrieving product")
	}

	var doc productDocument
	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return nil, err
	}

	return &Product{
		ID:          id,
		Name:        doc.Name,
		Description: doc.Description,
		Price:       doc.Price,
	}, nil
}

// ListProducts retrieves a list of products with pagination
func (r *elasticRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	query := map[string]interface{}{
		"from": skip,
		"size": take,
		"sort": []map[string]interface{}{
			{"name.keyword": map[string]string{"order": "asc"}},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("products"),
		r.client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New("error listing products")
	}

	var searchResult struct {
		Hits struct {
			Hits []struct {
				ID     string          `json:"_id"`
				Source productDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, err
	}

	products := make([]Product, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		products = append(products, Product{
			ID:          hit.ID,
			Name:        hit.Source.Name,
			Description: hit.Source.Description,
			Price:       hit.Source.Price,
		})
	}
	return products, nil
}

// ListProductsWithIDs retrieves products by their IDs
func (r *elasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"ids": map[string]interface{}{
				"values": ids,
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("products"),
		r.client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New("error listing products by IDs")
	}

	var searchResult struct {
		Hits struct {
			Hits []struct {
				ID     string          `json:"_id"`
				Source productDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, err
	}

	products := make([]Product, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		products = append(products, Product{
			ID:          hit.ID,
			Name:        hit.Source.Name,
			Description: hit.Source.Description,
			Price:       hit.Source.Price,
		})
	}
	return products, nil
}

// SearchProducts searches for products based on a query string
func (r *elasticRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	searchQuery := map[string]interface{}{
		"from": skip,
		"size": take,
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     query,
				"fields":    []string{"name", "description"},
				"fuzziness": "AUTO",
			},
		},
	}

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("products"),
		r.client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New("error searching products")
	}

	var searchResult struct {
		Hits struct {
			Hits []struct {
				ID     string          `json:"_id"`
				Source productDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, err
	}

	products := make([]Product, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		products = append(products, Product{
			ID:          hit.ID,
			Name:        hit.Source.Name,
			Description: hit.Source.Description,
			Price:       hit.Source.Price,
		})
	}
	return products, nil
}

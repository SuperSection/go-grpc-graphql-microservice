package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

var (
	ErrNotFound        = errors.New("entity not found")
	ErrFailedIndexing  = errors.New("failed to index document")
	ErrFailedFetching  = errors.New("failed to fetch document(s)")
	ErrFailedSearching = errors.New("failed to search documents")
)

type Repository interface {
	Close()
	PutProduct(ctx context.Context, p Product) error
	GetProductById(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type elasticRepository struct {
	Client *elastic.Client
}

type productDocument struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func NewElasticRepository(url string) (Repository, error) {
	client, err := elastic.NewClient(
		elastic.Config{
			Addresses: []string{
				url,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return &elasticRepository{Client: client}, nil
}

func (r *elasticRepository) Close() {

}

func (r *elasticRepository) PutProduct(ctx context.Context, p Product) error {
	doc := productDocument{
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}

	res, err := r.Client.Index(
		"catalog",
		esutil.NewJSONReader(doc),
		r.Client.Index.WithContext(ctx),
		r.Client.Index.WithDocumentID(p.ID),
		r.Client.Index.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return ErrFailedIndexing
	}
	return nil
}

func (r *elasticRepository) GetProductById(ctx context.Context, id string) (*Product, error) {
	res, err := esapi.GetRequest{
		Index:      "catalog",
		DocumentID: id,
	}.Do(ctx, r.Client)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !res.IsError() {
		if res.StatusCode == 404 {
			return nil, ErrNotFound
		}
		return nil, ErrFailedFetching
	}

	// Parse the response
	var doc struct {
		Source productDocument `json:"_source"`
	}
	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return nil, err
	}

	// Extract and return products
	return &Product{
		ID:          id,
		Name:        doc.Source.Name,
		Description: doc.Source.Description,
		Price:       doc.Source.Price,
	}, nil
}

func (r *elasticRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	query := map[string]interface{}{
		"from": skip,
		"size": take,
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	res, err := r.Client.Search(
		r.Client.Search.WithContext(ctx),
		r.Client.Search.WithIndex("catalog"),
		r.Client.Search.WithBody(esutil.NewJSONReader(query)),
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, ErrFailedSearching
	}

	// Parse the response
	var result struct {
		Hits []struct {
			ID     string          `json:"_id"`
			Source productDocument `json:"_source"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Extract and return products
	products := []Product{}
	for _, hit := range result.Hits {
		products = append(products, Product{
			ID:          hit.ID,
			Name:        hit.Source.Name,
			Description: hit.Source.Description,
			Price:       hit.Source.Price,
		})
	}
	return products, nil
}

func (r *elasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"ids": map[string]interface{}{
				"values": ids,
			},
		},
	}

	res, err := r.Client.Search(
		r.Client.Search.WithContext(ctx),
		r.Client.Search.WithIndex("catalog"),
		r.Client.Search.WithBody(esutil.NewJSONReader(query)),
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, ErrFailedFetching
	}

	// Parse the response
	var result struct {
		Hits []struct {
			ID     string          `json:"_id"`
			Source productDocument `json:"_source"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Extract and return products
	products := []Product{}
	for _, hit := range result.Hits {
		products = append(products, Product{
			ID:          hit.ID,
			Name:        hit.Source.Name,
			Description: hit.Source.Description,
			Price:       hit.Source.Price,
		})
	}
	return products, nil
}

func (r *elasticRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name", "description"},
			},
		},
		"from": skip,
		"size": take,
	}

	res, err := r.Client.Search(
		r.Client.Search.WithContext(ctx),
		r.Client.Search.WithIndex("catalog"),
		r.Client.Search.WithBody(esutil.NewJSONReader(searchQuery)),
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, ErrFailedSearching
	}

	// Parse the response
	var result struct {
		Hits []struct {
			ID     string          `json:"_id"`
			Source productDocument `json:"_source"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Extract and return products
	products := []Product{}
	for _, hit := range result.Hits {
		products = append(products, Product{
			ID:          hit.ID,
			Name:        hit.Source.Name,
			Description: hit.Source.Description,
			Price:       hit.Source.Price,
		})
	}
	return products, nil
}

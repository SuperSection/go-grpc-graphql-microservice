package catalog

import (
	"context"

	"github.com/SuperSection/go-grpc-graphql-microservice/catalog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.CatalogServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := pb.NewCatalogServiceClient(conn)
	return &Client{
		conn:    conn,
		service: c,
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostProduct(ctx context.Context, name string, description string, price float64) (*Product, error) {
	r, err := c.service.PostProduct(
		ctx,
		&pb.PostProductRequest{
			Name: name,
			Description: description,
			Price: price,
		},
	)
	if err != nil {
		return nil, err
	}

	return &Product{
		ID: r.Product.Id,
		Name: r.Product.Name,
		Description: r.Product.Description,
		Price: r.Product.Price,
	}, nil
}

func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	res, err := c.service.GetProduct(
		ctx,
		&pb.GetProductRequest{
			Id: id,
		},
	)
	if err != nil {
		return nil, err
	}

	return &Product{
		ID: res.Product.Id,
		Name: res.Product.Name,
		Description: res.Product.Description,
		Price: res.Product.Price,
	}, nil
}

func (c *Client) GetProducts(ctx context.Context, ids []string, query string, skip uint64, take uint64) ([]Product, error) {
	res, err := c.service.GetProducts(
		ctx,
		&pb.GetProductsRequest{
			Ids: ids,
			Query: query,
			Skip: skip,
			Take: take,
		},
	)
	if err != nil {
		return nil, err
	}

	products := []Product{}
	for _, p := range res.Products{
		products = append(products, Product{
			ID: p.Id,
			Name: p.Name,
			Description: p.Description,
			Price: p.Price,
		})
	}
	return products, nil
}

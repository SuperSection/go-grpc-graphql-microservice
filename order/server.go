package order

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	account "github.com/SuperSection/go-grpc-graphql-microservice/account"
	catalog "github.com/SuperSection/go-grpc-graphql-microservice/catalog"
	"github.com/SuperSection/go-grpc-graphql-microservice/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
  pb.UnimplementedOrderServiceServer
	service       Service
	accountClient *account.Client
	catalogClient *catalog.Client
}

func ListenGRPC(s Service, accountURL, catalogURL string, port int) error {
	accountClient, err := account.NewClient(accountURL)
	if err != nil {
		return err
	}

	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		accountClient.Close()
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		return err
	}

	server := grpc.NewServer()
	pb.RegisterOrderServiceServer(server, &grpcServer{
		service: s,
		accountClient: accountClient,
		catalogClient: catalogClient,
	})
	reflection.Register(server)
	return server.Serve(lis)
}

func (s *grpcServer) PostOrder(ctx context.Context, req *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	_, err := s.accountClient.GetAccount(ctx, req.AccountId)
	if err != nil {
		log.Println("Error getting account:", err)
		return nil, errors.New("account not found")
	}

	productIDs := []string{}
	orderedProducts, err := s.catalogClient.GetProducts(ctx, productIDs, "", 0, 0)
	if err != nil {
		log.Println("Error getting products:", err)
		return nil, errors.New("product not found")
	}

	products := []OrderedProduct{}
	for _, p := range orderedProducts {
		product := OrderedProduct{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    0,
		}
		for _, rp := range req.Products {
			if rp.ProductId == p.ID {
				product.Quantity = rp.Quantity
				break
			}
		}
		if product.Quantity != 0 {
			products = append(products, product)
		}
	}

	order, err := s.service.PostOrder(ctx, req.AccountId, products)
	if err != nil {
		log.Println("Error posting order:", err)
		return nil, errors.New("could not post order")
	}

	orderProto := &pb.Order{
		Id:         order.ID,
		AccountId:  order.AccountID,
		TotalPrice: order.TotalPrice,
		Products:   []*pb.Order_OrderProduct{},
	}
	orderProto.CreatedAt, _ = order.CreatedAt.MarshalBinary()
	for _, p := range order.Products {
		orderProto.Products = append(orderProto.Products,
			&pb.Order_OrderProduct{
				Id:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
				Quantity:    p.Quantity,
			})
	}
	return &pb.PostOrderResponse{
		Order: orderProto,
	}, nil
}

func (s *grpcServer) GetOrdersForAccount(ctx context.Context, req *pb.GetOrdersForAccountRequest) (*pb.GetOrdersForAccountResponse, error) {
	accountOrders, err := s.service.GetOrdersForAccount(ctx, req.AccountId)
	if err != nil {
		log.Println("Error getting orders for the account:", err)
		return nil, err
	}

	productIDMap := map[string]bool{}
	for _, o := range accountOrders {
		for _, p := range o.Products {
			productIDMap[p.ID] = true
		}
	}
	productIDs := []string{}
	for id := range productIDMap {
		productIDs = append(productIDs, id)
	}
	products, err := s.catalogClient.GetProducts(ctx, productIDs, "", 0, 0)
	if err != nil {
		log.Println("Error getting account products:", err)
		return nil, err
	}

	orders := []*pb.Order{}
	for _, o := range accountOrders {
		orderedProduct := &pb.Order{
			AccountId:  o.AccountID,
			Id:         o.ID,
			TotalPrice: o.TotalPrice,
			Products:   []*pb.Order_OrderProduct{},
		}
		orderedProduct.CreatedAt, _ = o.CreatedAt.MarshalBinary()

		for _, product := range o.Products {
			for _, p := range products {
				if p.ID == product.ID {
					product.Name = p.Name
					product.Description = p.Description
					product.Price = p.Price
					break
				}
			}
			orderedProduct.Products = append(orderedProduct.Products,
				&pb.Order_OrderProduct{
					Id:          product.ID,
					Name:        product.Name,
					Description: product.Description,
					Price:       product.Price,
					Quantity:    product.Quantity,
				})
		}
		orders = append(orders, orderedProduct)
	}

	return &pb.GetOrdersForAccountResponse{Orders: orders}, nil
}

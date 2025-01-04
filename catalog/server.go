package catalog

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/SuperSection/go-grpc-graphql-microservice/catalog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)


type grpcServer struct {
	pb.UnimplementedCatalogServiceServer
	service Service
}

func ListenGRPC(s Service, port int) error  {
	lis, err := net.Listen("tcp", fmt.Sprintf("%d", port))
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	pb.RegisterCatalogServiceServer(server, &grpcServer{
		service: s,
	})
	reflection.Register(server)
	return server.Serve(lis)
}


func (s *grpcServer) PostProduct(ctx context.Context, req *pb.PostProductRequest) (*pb.PostProductResponse, error) {
	p, err := s.service.PostProduct(ctx, req.Name, req.Description, req.Price)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &pb.PostProductResponse{
		Product: &pb.Product{
			Id: p.ID,
			Name: p.Name,
			Description: p.Description,
			Price: p.Price,
		},
	}, nil
}

func (s *grpcServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	p, err := s.service.GetProduct(ctx, req.Id)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &pb.GetProductResponse{
		Product: &pb.Product{
			Id: p.ID,
			Name: p.Name,
			Description: p.Description,
			Price: p.Price,
		},
	}, nil
}

func (s *grpcServer) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	var res []Product
	var err error
	if req.Query != "" {
		res, err = s.service.SearchProducts(ctx, req.Query, req.Skip, req.Take)
	} else if len(req.Ids) != 0 {
		res, err = s.service.GetProductsByIDs(ctx, req.Ids)
	} else {
		res, err = s.service.GetProducts(ctx, req.Skip, req.Take)
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}
	products := []*pb.Product{}
	for _, p := range res {
		products = append(products,
		&pb.Product{
			Id: p.ID,
			Name: p.Name,
			Description: p.Description,
			Price: p.Price,
		})
	}
	return &pb.GetProductsResponse{Products: products}, nil
}

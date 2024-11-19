package catalog

import (
	"context"
	"fmt"
	"github.com/ndquang191/go-graph-grpc/catalog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

type grpcServer struct {
	pb.UnimplementedCatalogServiceServer
	service Service
}

func ListenGRPC(service Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprint(":", port))
	if err != nil {
		return err
	}
	serv := grpc.NewServer()
	pb.RegisterCatalogServiceServer(serv, &grpcServer{
		UnimplementedCatalogServiceServer: pb.UnimplementedCatalogServiceServer{},
		service:                           service})
	reflection.Register(serv)
	return serv.Serve(lis)

}

func (s *grpcServer) PostProduct(ctx context.Context, rq *pb.PostProductRequest) (*pb.PostProductResponse, error) {
	p, err := s.service.PostProduct(ctx, rq.Name, rq.Description, rq.Price)
	if err != nil {
		return nil, err
	}

	return &pb.PostProductResponse{Product: &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}}, nil
}

func (s *grpcServer) GetProduct(ctx context.Context, rq *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	p, err := s.service.GetProduct(ctx, rq.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetProductResponse{Product: &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}}, nil
}

func (s *grpcServer) GetProducts(ctx context.Context, r *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	var res []Product
	var err error
	if r.Query != "" {
		res, err = s.service.SearchProducts(ctx, r.Query, r.Skip, r.Take)
	} else if len(r.Ids) != 0 {
		res, err = s.service.GetProductsByIDs(ctx, r.Ids)
	} else {
		res, err = s.service.GetProducts(ctx, r.Skip, r.Take)
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}

	products := []*pb.Product{}

	for _, p := range res {
		products = append(products, &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		})
	}

	return &pb.GetProductsResponse{Products: products}, nil

}

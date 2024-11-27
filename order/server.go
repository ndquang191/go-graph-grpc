package order

import (
	"context"
	"fmt"
	"github.com/ndquang191/go-graph-grpc/account"
	"github.com/ndquang191/go-graph-grpc/catalog"
	"github.com/ndquang191/go-graph-grpc/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
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

	serv := grpc.NewServer()
	pb.RegisterOrderServiceServer(serv, &grpcServer{
		UnimplementedOrderServiceServer: pb.UnimplementedOrderServiceServer{},
		service:                         s,
		accountClient:                   accountClient,
		catalogClient:                   catalogClient,
	})

	reflection.Register(serv)
	return serv.Serve(lis)
}

func (s *grpcServer) PostOrder(ctx context.Context, req *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {

	_, err := s.accountClient.GetAccount(ctx, req.AccountId)
	if err != nil {
		log.Print("Error getting account: ", err)
		return nil, err
	}

	productIDs := []string{}
	orderedProducts, err := s.catalogClient.GetProducts(ctx, 0, 0, productIDs, "")

	if err != nil {
		log.Print("Error getting products: ", err)
		return nil, err
	}

	products := []OrderedProduct{}

	for _, p := range orderedProducts {
		product := OrderedProduct{
			ID:          p.ID,
			Name:        p.Name,
			Price:       p.Price,
			Quantity:    0,
			Description: p.Description,
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
		log.Print("Error posting order: ", err)
		return nil, err
	}

	orderProto := &pb.Order{
		Id:         order.ID,
		AccountId:  order.AccountID,
		TotalPrice: order.TotalPrice,
		Products:   []*pb.Order_OrderedProduct{},
	}

	binaryData, err := order.CreatedAt.MarshalBinary()
	if err != nil {
		log.Fatalf("Failed to marshal CreatedAt: %v", err)
	}
	orderProto.CreatedAt = string(binaryData)

	for _, p := range order.Products {

		orderProto.Products = append(orderProto.Products, &pb.Order_OrderedProduct{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Quantity:    p.Quantity,
			Price:       p.Price,
		})
	}

	return &pb.PostOrderResponse{
		Order: orderProto,
	}, nil

}

func (s *grpcServer) GetOrdersForAccount(ctx context.Context, req *pb.GetOrdersForAccountRequest) (*pb.GetOrdersForAccountResponse, error) {

	accountOrders, err := s.service.GetOrdersForAccount(ctx, req.AccountId)

	if err != nil {
		log.Print("Error getting orders: ", err)
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

	products, err := s.catalogClient.GetProducts(ctx, 0, 0, productIDs, "")

	if err != nil {
		log.Print("Error getting products: ", err)
		return nil, err
	}

	orders := []*pb.Order{}
	for _, o := range accountOrders {
		op := &pb.Order{
			Id:         o.ID,
			AccountId:  o.AccountID,
			TotalPrice: o.TotalPrice,
			Products:   []*pb.Order_OrderedProduct{},
		}

		binaryData, err := o.CreatedAt.MarshalBinary()
		if err != nil {
			log.Fatalf("Failed to marshal CreatedAt: %v", err)
		}
		op.CreatedAt = string(binaryData)

		for _, product := range o.Products {
			for _, p := range products {

				if p.ID == product.ID {
					product.Name = p.Name
					product.Description = p.Description
					product.Price = p.Price
					break
				}
			}

			op.Products = append(op.Products, &pb.Order_OrderedProduct{
				Id:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Quantity:    product.Quantity,
				Price:       product.Price,
			})

		}
		orders = append(orders, op)
	}

	return &pb.GetOrdersForAccountResponse{
		Orders: orders,
	}, nil
}

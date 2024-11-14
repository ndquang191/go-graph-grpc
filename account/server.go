package account

import (
	"context"
	"fmt"
	"github.com/ndquang191/go-graph-grpc/account/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type grpcServer struct {
	pb.UnimplementedAccountServiceServer
	service Service
}

func ListenGRPC(service Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprint(":", port))
	if err != nil {
		return err
	}
	serv := grpc.NewServer()
	pb.RegisterAccountServiceServer(serv, &grpcServer{service: service})
	reflection.Register(serv)
	return serv.Serve(lis)

}

func (s *grpcServer) PostAccount(ctx context.Context, rq *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	a, err := s.service.PostAccount(ctx, rq.Name)
	if err != nil {
		return nil, err
	}

	return &pb.PostAccountResponse{Account: &pb.Account{
		Id:   a.ID,
		Name: a.Name,
	}}, nil
}
func (s *grpcServer) GetAccount(ctx context.Context, rq *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	a, err := s.service.GetAccountByID(ctx, rq.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetAccountResponse{Account: &pb.Account{
		Id:   a.ID,
		Name: a.Name,
	}}, nil
}

func (s *grpcServer) GetAccounts(ctx context.Context, rq *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	res, err := s.service.ListAccounts(ctx, rq.Skip, rq.Take)
	if err != nil {
		return nil, err
	}

	accounts := []*pb.Account{}
	for _, p := range res {
		accounts = append(accounts, &pb.Account{
			Id:   p.ID,
			Name: p.Name,
		})
	}

	return &pb.GetAccountsResponse{Accounts: accounts}, nil
}

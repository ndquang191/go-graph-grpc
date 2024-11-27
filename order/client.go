package order

import (
	"context"
	"time"

	"github.com/ndquang191/go-graph-grpc/order/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := pb.NewOrderServiceClient(conn)
	return &Client{
		conn:    conn,
		service: c,
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostOder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {
	protoProducts := []*pb.PostOrderRequest_OrderedProduct{}

	for _, p := range products {
		protoProducts = append(protoProducts, &pb.PostOrderRequest_OrderedProduct{
			ProductId: p.ID,
			Quantity:  p.Quantity,
		})
	}
	res, err := c.service.PostOrder(ctx, &pb.PostOrderRequest{
		AccountId: accountID,
		Products:  protoProducts,
	})

	if err != nil {
		return nil, err
	}

	newOrder := res.Order

	newOrderCreatedAt := time.Time{}

	newOrderCreatedAt.UnmarshalBinary([]byte(newOrder.CreatedAt))

	return &Order{
		ID:         newOrder.Id,
		CreatedAt:  newOrderCreatedAt,
		TotalPrice: newOrder.TotalPrice,
		AccountID:  newOrder.AccountId,
		Products:   products,
	}, nil
}

func (c *Client) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	res, err := c.service.GetOrdersForAccount(ctx, &pb.GetOrdersForAccountRequest{
		AccountId: accountID,
	})

	if err != nil {
		return nil, err
	}

	orders := []Order{}

	for _, orderProto := range res.Orders {
		newOrder := Order{
			ID:         orderProto.Id,
			TotalPrice: orderProto.TotalPrice,
			AccountID:  orderProto.AccountId,
		}

		newOrder.CreatedAt = time.Time{}
		newOrder.CreatedAt.UnmarshalBinary([]byte(orderProto.CreatedAt))

		products := []OrderedProduct{}

		for _, p := range orderProto.Products {
			products = append(products, OrderedProduct{
				ID:          p.Id,
				Quantity:    p.Quantity,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}

		newOrder.Products = products
		orders = append(orders, newOrder)

	}

	return orders, nil
}

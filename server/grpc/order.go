package grpc

import (
	"admin/models"
	pb "admin/protos"
	"context"
	"errors"
	"fmt"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/paymentintent"
	"github.com/stripe/stripe-go/v71/paymentmethod"
)

func (s *Server) CreateOrder(ctx context.Context, in *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {

	c, cErr := s.GetOrCreateCustomerByUserId(ctx, &pb.Id{Id: in.GetUserId()})
	if cErr != nil {
		return nil, cErr
	}
	sess, sessErr := s.GetSessionById(ctx, &pb.GetSessionByIdRequest{Id: in.GetSessionId(), UserId: in.GetUserId()})
	if sessErr != nil {
		return nil, sessErr
	}
	if sess.GetPurchased() {
		return nil, errors.New("user has already purchased this session")
	}

	// validate amount paid is equal to session cost if the transaction type is not a donation
	if !(in.GetTransactionType() == pb.TransactionType_DONATION) {
		if sess.GetCost() != in.GetCost() {
			return nil, fmt.Errorf("session cost in backend: %d does not match session cost from client: %d", sess.GetCost(), in.GetCost())
		}
	}

	// default to purchase
	if in.GetTransactionType() == pb.TransactionType_NONE {
		in.TransactionType = pb.TransactionType_PURCHASE
	}

	params := &stripe.PaymentIntentParams{
		Customer: stripe.String(c.GetPaymentPlatformCustomerId()),
		Amount:   stripe.Int64(in.GetCost()),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		PaymentMethodTypes: []*string{
			stripe.String(paymentMethodDefault),
		},
		Description: stripe.String(fmt.Sprintf("%d-%s", sess.GetId(), sess.GetName())),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	order := &pb.Order{
		CustomerId:                c.GetId(),
		PaymentPlatform:           pb.PaymentPlatforms_STRIPE,
		Status:                    pb.Order_CREATED,
		SessionId:                 sess.GetId(),
		PaymentPlatformOrderId:    pi.ID,
		PaymentPlatformCustomerId: c.GetPaymentPlatformCustomerId(),
		Amount:                    in.GetCost(),

		SessionName:        sess.GetName(),
		SessionDescription: sess.GetDescription(),
		Tags:               sess.GetTags(),

		Email:           c.GetEmail(),
		TransactionType: in.GetTransactionType(),
	}

	if order.GetCustomerId() == 0 || order.GetSessionId() == 0 ||
		order.GetPaymentPlatformOrderId() == "" ||
		order.GetPaymentPlatformCustomerId() == "" ||
		order.Amount <= 0 {
		return nil, fmt.Errorf("error, required fields are missing in order: %+v", order)
	}

	if pi.ClientSecret == "" {
		return nil, errors.New("no client secret returned")
	}

	toCreate := &models.Order{}
	toCreate.ProtoUnMarshal(order)

	out, createErr := s.repo.CreateOrder(toCreate)
	if createErr != nil {
		return nil, createErr
	}

	// check if the customer has an existing paymentmethod to attach to the payment intent
	pmParams := &stripe.PaymentMethodListParams{
		Customer: stripe.String(c.GetPaymentPlatformCustomerId()),
		Type:     stripe.String("card"),
	}
	i := paymentmethod.List(pmParams)
	methods := []*pb.PaymentMethodData{}
	for i.Next() {
		pmi := i.PaymentMethod()
		methods = append(methods, &pb.PaymentMethodData{
			PaymentMethodId: pmi.ID,
			LastFour:        pmi.Card.Last4,
			ExpMonth:        int64(pmi.Card.ExpMonth),
			ExpYear:         int64(pmi.Card.ExpYear),
			Network:         pmi.Card.Issuer,
			Brand:           string(pmi.Card.Brand),
		})
	}
	return &pb.CreateOrderResponse{
		Request:        in,
		Order:          out.ProtoMarshal(),
		ClientSecret:   pi.ClientSecret,
		PaymentMethods: methods,
	}, nil

}

// Gets a order by ID.
func (s *Server) GetOrderById(ctx context.Context, in *pb.Id) (*pb.Order, error) {
	u, err := s.repo.GetOrderById(uint(in.GetId()))
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshal(), nil
}

// Gets a user by ID.
func (s *Server) GetOrderByPaymentPlatformOrderId(ctx context.Context, in *pb.PaymentPlatformOrderId) (*pb.Order, error) {
	u, err := s.repo.GetOrderByPaymentPlatformOrderId(in.GetId())
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshal(), nil
}

func (s *Server) UpdateOrderPaymentMethodId(ctx context.Context, in *pb.UpdateOrderPaymentMethodIdRequest) (*pb.Order, error) {
	o, err := s.repo.UpdateOrderPaymentMethodId(uint(in.GetOrderId()), in.GetPaymentMethodId())
	if err != nil {
		return nil, err
	}
	return o.ProtoMarshal(), nil
}

// updates a order status
func (s *Server) UpdateOrderStatusByOrderId(ctx context.Context, in *pb.Order) (*pb.Order, error) {
	toUpdate := &models.Order{}
	toUpdate.ProtoUnMarshal(in)
	u, err := s.repo.UpdateOrderStatusByOrderId(toUpdate)
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshal(), nil
}

func (s *Server) GetSucceededOrdersByCustomerId(ctx context.Context, in *pb.Id) (*pb.Orders, error) {
	orders, err := s.repo.GetSucceededOrdersByCustomerId(uint(in.GetId()))
	if err != nil {
		return nil, err
	}
	return orders.ProtoMarshal(), nil
}

package grpc

import (
	"admin/models"
	pb "admin/protos"
	"context"
	"fmt"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/customer"
)

// Gets a user by SubID.
func (s *Server) GetOrCreateCustomerByUserId(ctx context.Context, in *pb.Id) (*pb.Customer, error) {

	c, err := s.repo.GetOrCreateCustomerByUserId(uint(in.GetId()))
	if err != nil {
		return nil, err
	}

	u, uErr := s.repo.GetUserById(uint(in.GetId()))
	if uErr != nil {
		return nil, err
	}

	if c.PaymentPlatformCustomerId != "" {
		return c.ProtoMarshal(), nil
	}

	params := &stripe.CustomerParams{
		Email:       stripe.String(u.Email),
		Description: stripe.String(fmt.Sprintf("%s %s", u.FirstName, u.LastName)),
	}
	customer, cErr := customer.New(params)
	if cErr != nil {
		return nil, cErr
	}
	if customer.ID == "" {
		return nil, fmt.Errorf("error: customer ID is empty returned from stripe, for user: %v", u)
	}
	c.PaymentPlatformCustomerId = customer.ID
	c.PaymentPlatform = pb.PaymentPlatforms_STRIPE.Enum().String()

	c, err = s.repo.UpdateCustomer(c)
	if err != nil {
		return nil, err
	}
	return c.ProtoMarshal(), nil
}

// updates a customer by SubID.
func (s *Server) UpdateCustomer(ctx context.Context, in *pb.Customer) (*pb.Customer, error) {
	toUpdate := &models.Customer{}
	toUpdate.ProtoUnMarshal(in)
	u, err := s.repo.UpdateCustomer(toUpdate)
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshal(), nil
}

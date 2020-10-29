package grpc

import (
	pb "admin/protos"
	"context"
	"fmt"
)

const (
	testStripeClientId = "ca_HaYN6yH8fXZ6suMre4v8W5u0XcUE4xmW"
)

func (s *Server) GetStripeAppInstallUrl(ctx context.Context, in *pb.StripeAppInstallUrlRequest) (*pb.StripeAppInstallUrlResponse, error) {
	p, err := s.GetUserByIdPrivate(ctx, &pb.Id{Id: in.GetUserId()})
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("https://connect.stripe.com/express/oauth/authorize?redirect_uri=https://connect.stripe.com/connect/default/oauth/test&client_id=%s&state=%d&stripe_user[email]=%s", testStripeClientId, p.GetId(), p.GetEmail())
	return &pb.StripeAppInstallUrlResponse{Url: u}, nil
}

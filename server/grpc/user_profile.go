package grpc

import (
	pb "admin/protos"
	"context"
	"golang.org/x/sync/errgroup"
)

func (s *Server) GetUserProfile(ctx context.Context, in *pb.Id) (*pb.UserProfile, error) {

	user, err := s.GetUserByIdPrivate(ctx, &pb.Id{Id: in.GetId()})
	if err != nil {
		return nil, err
	}

	// For user profile we want to hydrate
	//	- previous/upcoming/live hosted sessions   (2 queries)
	//  - previous/upcoming/live registered sessions (2 queries)
	//  - most up to date zoom token  (queries + zoom API call)
	// We can do all these asynchronously so lets do them each in their own go routine and combine into the user profile when all are complete.

	var gtProfile errgroup.Group
	var userOrderdSessions *pb.User
	var zoomToken *pb.ZoomToken
	gtProfile.Go(func() error {
		user, err = s.hydrateUser(user)
		if err != nil {
			return err
		}
		return nil
	})

	gtProfile.Go(func() error {
		userOrderdSessions, err = s.hydrateUserOrderedSessions(user)
		if err != nil {
			return err
		}
		return nil
	})

	var hasToken = false

	gtProfile.Go(func() error {
		if !user.GetZoomAppInstalled() {
			return nil
		}
		if zoomToken, err = s.GetZoomTokenByUserId(ctx, &pb.Id{Id: user.GetId()}); err == nil {
			hasToken = true
		}
		return nil
	})

	// Wait for all hydrations to complete.
	if gtErr := gtProfile.Wait(); gtErr != nil {
		return nil, gtErr
	}

	user.OrderedPreviousSessions = userOrderdSessions.OrderedPreviousSessions
	user.OrderedUpcomingSessions = userOrderdSessions.OrderedUpcomingSessions
	user.OrderedLiveSessions = userOrderdSessions.OrderedLiveSessions

	profile := &pb.UserProfile{
		User: user,
	}

	if hasToken {
		profile.ZoomToken = zoomToken
	}
	return profile, nil

}

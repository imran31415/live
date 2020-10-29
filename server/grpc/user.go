package grpc

import (
	"admin/models"
	pb "admin/protos"
	"context"
	"golang.org/x/sync/errgroup"
	"time"
)

func (s *Server) GetUserById(ctx context.Context, in *pb.Id) (*pb.User, error) {
	u, err := s.repo.GetUserById(uint(in.GetId()))
	if err != nil {
		return nil, err
	}
	return s.hydrateUser(u.ProtoMarshalPrivate())
}

func (s *Server) GetUserByIdPrivate(ctx context.Context, in *pb.Id) (*pb.User, error) {
	u, err := s.repo.GetUserById(uint(in.GetId()))
	if err != nil {
		return nil, err
	}
	return s.hydrateUser(u.ProtoMarshalPrivate())
}

func (s *Server) GetUserByIdPublic(ctx context.Context, in *pb.Id) (*pb.User, error) {
	u, err := s.repo.GetUserById(uint(in.GetId()))
	if err != nil {
		return nil, err
	}
	return s.hydrateUser(u.ProtoMarshalPublic())
}

// TODO: Depricate?
func (s *Server) GetUserByEmail(ctx context.Context, in *pb.Email) (*pb.User, error) {
	u, err := s.repo.GetUserByEmail(in.GetEmail())
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshalPublic(), nil
}

func (s *Server) GetUserByZoomAccountId(ctx context.Context, in *pb.ZoomAccountId) (*pb.User, error) {
	u, err := s.repo.GetUserByZoomAccountId(in.GetId())
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshalPrivate(), nil
}

// Gets a user by SubID.
func (s *Server) GetOrCreateUserBySubId(ctx context.Context, in *pb.SubId) (*pb.User, error) {
	u, err := s.repo.GetOrCreateUserBySubId(in.GetId())
	if err != nil {
		return nil, err
	}
	return s.hydrateUser(u.ProtoMarshalPrivate())
}

// updates a user by SubID.
func (s *Server) UpdateUserBySubId(ctx context.Context, in *pb.User) (*pb.User, error) {
	toUpdate := &models.User{}
	toUpdate.ProtoUnMarshal(in)
	u, err := s.repo.UpdateUserBySubId(toUpdate)
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshalPrivate(), nil
}

// updates a user by SubID.
func (s *Server) UpdateUserById(ctx context.Context, in *pb.User) (*pb.User, error) {
	toUpdate := &models.User{}
	toUpdate.ProtoUnMarshal(in)
	u, err := s.repo.UpdateUserById(toUpdate)
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshalPrivate(), nil
}

// updates a user by SubID.
func (s *Server) UpdateUserZoomDeAuthorized(ctx context.Context, in *pb.Id) (*pb.Empty, error) {
	err := s.repo.UpdateUserZoomDeAuthorized(uint(in.GetId()))
	return &pb.Empty{}, err
}

func (s *Server) hydrateUser(hydrated *pb.User) (*pb.User, error) {
	u, err := s.hydrateUserHostedSessions(hydrated)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Server) hydrateUserHostedSessions(hydrated *pb.User) (*pb.User, error) {
	// run the repo queries in go routines to reduce latency
	var gtSessions errgroup.Group
	var upcoming, previous models.Sessions

	var cErr, pErr, hcuErr, hpcErr error

	gtSessions.Go(func() error {
		upcoming, cErr = s.repo.GetUpcomingSessionsByUserIdAndDate(uint(hydrated.GetId()), time.Now().Unix())
		if cErr != nil {
			return cErr
		}

		return nil
	})
	gtSessions.Go(func() error {
		previous, pErr = s.repo.GetPreviousSessionsByUserIdAndDate(uint(hydrated.GetId()), time.Now().Unix())
		if pErr != nil {
			return pErr
		}
		return nil
	})
	// Wait for all fetches to complete.
	if getSessionsErr := gtSessions.Wait(); getSessionsErr != nil {
		return nil, getSessionsErr
	}

	var hydrateClasses errgroup.Group

	hydrateClasses.Go(func() error {
		hydrated.UpcomingSessions, hcuErr = s.hydrateSessions(context.Background(), upcoming.ProtoMarshalPrivate())
		if hcuErr != nil {
			return hcuErr
		}
		return nil
	})

	// We need to parse live sessions out of previous based on duration so we create a temp array

	// initialize array fields so they are not nil and can be appended
	hydrated.LiveSessions = &pb.Sessions{Sessions: []*pb.Session{}}
	hydrated.PreviousSessions = &pb.Sessions{Sessions: []*pb.Session{}}
	var prevSessions *pb.Sessions
	hydrateClasses.Go(func() error {
		prevSessions, hpcErr = s.hydrateSessions(context.Background(), previous.ProtoMarshalPrivate())
		if hpcErr != nil {
			return hpcErr
		}
		return nil
	})

	// Wait for all hydrations to complete.
	if hydrSessErr := hydrateClasses.Wait(); hydrSessErr != nil {
		return nil, hydrSessErr
	}

	// parse out live from previous sesssions
	now := time.Now().Unix()
	for _, i := range prevSessions.GetSessions() {
		endTime := i.GetStartTime() + i.GetDuration()*60
		if now < endTime {
			hydrated.LiveSessions.Sessions = append(hydrated.LiveSessions.Sessions, i)
		} else {
			hydrated.PreviousSessions.Sessions = append(hydrated.PreviousSessions.Sessions, i)
		}

	}

	return hydrated, nil
}

func (s *Server) hydrateUserOrderedSessions(hydrated *pb.User) (*pb.User, error) {
	customer, err := s.GetOrCreateCustomerByUserId(context.Background(), &pb.Id{Id: hydrated.GetId()})
	if err != nil {
		return nil, err
	}

	orders, ordersErr := s.GetSucceededOrdersByCustomerId(context.Background(), &pb.Id{Id: customer.GetId()})
	if ordersErr != nil {
		return nil, ordersErr
	}

	if len(orders.GetOrders()) == 0 {
		// no ordered sessions to hydrate
		return hydrated, nil
	}

	orderedSesionIds := []uint{}
	for _, i := range orders.GetOrders() {
		orderedSesionIds = append(orderedSesionIds, uint(i.GetSessionId()))
	}

	//orders, err := s.
	// run the repo queries in go routines to reduce latency
	var gtSessions errgroup.Group
	var upcoming, previous models.Sessions

	var cErr, pErr, hcuErr, hpcErr error

	gtSessions.Go(func() error {
		upcoming, cErr = s.repo.GetUpcomingSessionsByIds(orderedSesionIds)
		if cErr != nil {
			return cErr
		}

		return nil
	})
	gtSessions.Go(func() error {
		previous, pErr = s.repo.GetPreviousSessionsByIds(orderedSesionIds)
		if pErr != nil {
			return pErr
		}
		return nil
	})
	// Wait for all fetches to complete.
	if getSessionsErr := gtSessions.Wait(); getSessionsErr != nil {
		return nil, getSessionsErr
	}

	var hydrateClasses errgroup.Group

	hydrateClasses.Go(func() error {
		hydrated.OrderedUpcomingSessions, hcuErr = s.hydrateSessions(context.Background(), upcoming.ProtoMarshalPrivate())
		if hcuErr != nil {
			return hcuErr
		}
		return nil
	})

	// We need to parse live sessions out of previous based on duration so we create a temp array

	// initialize array fields so they are not nil and can be appended
	hydrated.OrderedLiveSessions = &pb.Sessions{Sessions: []*pb.Session{}}
	hydrated.OrderedPreviousSessions = &pb.Sessions{Sessions: []*pb.Session{}}
	var prevSessions *pb.Sessions
	hydrateClasses.Go(func() error {
		prevSessions, hpcErr = s.hydrateSessions(context.Background(), previous.ProtoMarshalPrivate())
		if hpcErr != nil {
			return hpcErr
		}
		return nil
	})

	// Wait for all hydrations to complete.
	if hydrSessErr := hydrateClasses.Wait(); hydrSessErr != nil {
		return nil, hydrSessErr
	}

	// parse out live from previous sesssions
	now := time.Now().Unix()
	for _, i := range prevSessions.GetSessions() {
		endTime := i.GetStartTime() + i.GetDuration()*60
		if now < endTime {
			hydrated.OrderedLiveSessions.Sessions = append(hydrated.OrderedLiveSessions.Sessions, i)
		} else {
			hydrated.OrderedPreviousSessions.Sessions = append(hydrated.OrderedPreviousSessions.Sessions, i)
		}

	}

	return hydrated, nil
}

package grpc

import (
	"admin/models"
	pb "admin/protos"
	"admin/server/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// updates a Session by Id.
func (s *Server) UpdateSession(ctx context.Context, in *pb.Session) (*pb.Session, error) {
	toUpdate := &models.Session{}
	toUpdate.ProtoUnMarshal(in)
	u, err := s.repo.UpdateSession(toUpdate)
	if err != nil {
		return nil, err
	}
	if u.ProtoMarshalPrivate().GetZoomSyncEnabled().GetValue() && in.GetShouldZoomSyncOnUpdate() {
		updateInZoomErr := s.updateMeetingInZoom(ctx, u.ProtoMarshalPrivate())
		if updateInZoomErr != nil {
			log.Println("error updating meeting in zoom, err", updateInZoomErr)
		}
	}

	return s.hydrateSession(u.ProtoMarshalPrivate()), nil
}

func (s *Server) createSessionFromZoomMeetingInfo(meetingInfo *utils.MeetingInfo, userId int64) (int64, []string) {
	success := int64(0)
	errs := []string{}
	sessionToCreate := utils.SessionFromZoomMeetingWebHookCreated("", meetingInfo, userId)

	ex, exEr := s.GetSessionByUserIdAndZoomMeetingId(context.Background(), &pb.GetSessionByUserIdAndZoomMeetingIdRequest{
		UserId: userId,
		ZoomId: sessionToCreate.GetZoomMeetingId().GetValue(),
	})

	// if the existing meeting is a different type than what is in zoom lets delete and re-create
	if exEr == nil && ex.GetZoomMeetingType() != int64(meetingInfo.Type) {
		log.Println("Existing meeting has a different type then zoom meeting type, deleting the existing meeting for recreation")
		if dErr := s.repo.DeleteSessionsByUserIdAndZoomMeetingId(uint(userId), uint(ex.GetZoomMeetingId().GetValue())); dErr != nil {
			return 0, []string{dErr.Error()}
		}
	}
	switch sessionToCreate.GetZoomMeetingType() {
	// re-occurring meeting we have to parse  the meetings occurrences array and create meetings based on those times and durations
	case utils.ZoomMeetingTypeReocurring:
		if meetingInfo.Occurrences == nil || len(meetingInfo.Occurrences) < 1 {
			errs = append(errs, errors.New("error processWebHookMeetingCreated: meeting type is 8 but Occurrences array is empty, not processing").Error())
		}
		foundOccurrences := []string{}

		for _, o := range meetingInfo.Occurrences {
			foundOccurrences = append(foundOccurrences, o.OccurrenceID)

			sess := sessionToCreate
			sess.StartTime = o.StartTime.Unix()
			sess.Duration = int64(o.Duration)
			sess.ZoomOccurrenceId = o.OccurrenceID
			existing, exErr := s.GetSessionByUserIdAndZoomMeetingIdAndOccurrenceId(context.Background(), &pb.GetSessionByUserIdAndZoomMeetingIdAndOccurrenceIdRequest{
				UserId:       userId,
				ZoomId:       sess.GetZoomMeetingId().GetValue(),
				OccurrenceId: o.OccurrenceID,
			})
			if exErr == nil {
				// if the status is deleted lets delete it and continue
				if o.Status == "deleted" {
					s.repo.DeleteSessionById(uint(existing.GetId()))
					continue
				}
				// meeting with occurrence exists in our system
				// check whether the existing meeting fields are the same as the new meeting fields:
				if existing.GetStartTime() == o.StartTime.Unix() && existing.GetDuration() == int64(o.Duration) && meetingInfo.Topic == existing.GetName() {
					log.Println("Ignoring ocurrence as there are no changes to this session")
					// there is no change to apply, no need to create the session for this occurrence
					continue
				}

				// the occurrence has changed lets delete the existing one and re-create
				if dErr := s.repo.DeleteSessionById(uint(existing.GetId())); dErr != nil {
					errs = append(errs, dErr.Error())
					log.Println("Unable to delete existing session, not re-creating and continuing to next occurrence")
				}
			}
			if o.Status == "deleted" {
				// dont create if its deleted
				continue
			}
			created, cErr := s.CreateSession(context.Background(), sessionToCreate)
			if cErr != nil {
				errs = append(errs, cErr.Error())
			} else {
				log.Println("Successfully processWebHookMeetingCreated, meeting id: ", created.GetId(), " for user: ", created.GetUserId())
				success += 1
			}
		}
		// TOOD: Delete all sessions with occurrence IDs that do not match the most updated list.
		if delErr := s.repo.DeleteSessionsByUserIdAndZoomMeetingIdAndNotOccurrenceIds(uint(userId), uint(meetingInfo.ID), foundOccurrences); delErr != nil {
			log.Println("Err occurred deleting old occurrences: ", delErr)
		}
	case utils.ZoomMeetingTypeSingular:
		if exEr == nil {
			if ex.GetStartTime() == meetingInfo.StartTime.Unix() && ex.GetDuration() == int64(meetingInfo.Duration) && meetingInfo.Topic == ex.GetName() {
				log.Println("Ignoring creating meeting as there are no changes")
				// there is no change to apply, no need to create the session for this occurrence
				return 0, nil
			}
		}

		created, cErr := s.CreateSession(context.Background(), sessionToCreate)
		if cErr != nil {
			errs = append(errs, fmt.Errorf("error: :%s  received when creating meeting with fields %+v for meeting info with payload %+v", cErr, sessionToCreate, meetingInfo).Error())
		} else {
			success += 1
			log.Println("Successfully processWebHookMeetingCreated, meeting id: ", created.GetId(), " for user: ", created.GetUserId())
		}
	default:
		log.Println("createSessionFromZoomMeetingInfo received un-handled meeting type: ", sessionToCreate.GetZoomMeetingType(), " skipping msg")

	}
	return success, errs
}

// creates Session with supplied values
func (s *Server) CreateSession(ctx context.Context, in *pb.Session) (*pb.Session, error) {
	toCreate := &models.Session{}
	toCreate.ProtoUnMarshal(in)

	// Validate Fields
	if toCreate.UserId == 0 {
		return nil, errors.New("can not create session without user_id set")
	}

	if toCreate.Name == "" {
		return nil, errors.New("can not create Session with empty name")
	}

	if toCreate.Date < 0 {
		return nil, errors.New("can not create Session with date value less than 0")
	}

	// Dedupe by name and user id and date
	if existing, err := s.repo.GetSessionByUserIdAndNameAndDateAndDuration(toCreate.UserId, toCreate.Date, toCreate.Duration, toCreate.Name); err == nil && existing.ID != 0 {
		return nil, errors.New("session with the same name, user_id, duration and date already exists")
	}

	// Meeting ID is set and non-zero
	if toCreate.ZoomMeetingId != nil && *toCreate.ZoomMeetingId != 0 {
		if toCreate.ZoomOccurrenceId != "" {
			if existing, err := s.repo.GetSessionByUserIdAndZoomMeetingIdAndOccurrenceId(toCreate.UserId, *toCreate.ZoomMeetingId, toCreate.ZoomOccurrenceId); err == nil && existing.ID != 0 {
				return nil, errors.New("session with the same user_id and zoom_meeting_id and occurrence id already exists")
			}
		} else {
			if existing, err := s.repo.GetSessionByUserIdAndZoomMeetingId(toCreate.UserId, *toCreate.ZoomMeetingId); err == nil && existing.ID != 0 {
				return nil, errors.New("session with the same user_id and zoom_meeting_id already exists")
			}
		}

	}

	// Validate supplied user exists
	_, err := s.GetUserById(ctx, &pb.Id{Id: toCreate.UserId})
	if err != nil {
		return nil, err
	}

	// validate unix time
	timeStr := strconv.Itoa(int(toCreate.Date))
	_, cErr := strconv.ParseInt(timeStr, 10, 64)
	if cErr != nil {
		log.Println("Can not create Session without valid time stamp")
		return nil, cErr
	}

	if toCreate.Duration < 1 || toCreate.Duration > 10000 {
		return nil, errors.New("session duration must be between 1 and 10,000")
	}

	// Get unsplash if images are empty and created from webhook:
	shouldGetUnsplash := false

	if (toCreate.ProfileImgUrl == "" || toCreate.BannerImgUrl == "") &&
		toCreate.Source == pb.Session_ZOOM_WEBHOOK.Enum().String() {
		shouldGetUnsplash = true
	}
	if shouldGetUnsplash {
		unsplashBytes, uErr := utils.GetUnsplashImages(toCreate.Name)
		if uErr != nil {
			log.Println("Unable to get unsplash image response, error: ", uErr, " ignoring error and will continue to save session")
		}
		if toCreate.ProfileImgUrl == "" {
			if toCreate.ProfileImgUrl, err = utils.RandomImageFromUnsplashResponse(unsplashBytes); err != nil {
				log.Println("Error parsing image for Profile from unsplash response: ", err)
				toCreate.ProfileImgUrl = "/Shared/via_banner_default.png"
			}
		}
		if toCreate.BannerImgUrl == "" {
			if toCreate.BannerImgUrl, err = utils.RandomImageFromUnsplashResponse(unsplashBytes); err != nil {
				log.Println("Error parsing image for Banner from unsplash response: ", err)
				toCreate.ProfileImgUrl = "/Shared/via_banner_default.png"
			}
		}
	}

	u, crErr := s.repo.CreateSession(toCreate)
	if crErr != nil {
		return nil, crErr
	}

	createdSession := u.ProtoMarshalPrivate()
	// decide if we need to create in zoom
	if !in.GetCreateMeetingInZoom() {
		return createdSession, nil
	}

	zmToken, zmTokenErr := s.GetZoomTokenByUserId(ctx, &pb.Id{Id: createdSession.GetUserId()})
	if zmTokenErr != nil {
		log.Println("CreateZoomMeeting is true but returned err retrieving zoom token", zmTokenErr)
		return createdSession, zmTokenErr
	}

	zm, zmErr := s.CreateMeetingInZoom(ctx, &pb.CreateMeetingInZoomRequest{
		Fields: &pb.CreateZoomMeetingFields{
			Topic:     u.Name,
			Type:      2,
			StartTime: time.Unix(createdSession.GetStartTime(), 0).Format(time.RFC3339),
			// get first 10 characters of random due to zoom limit
			Password: strconv.Itoa(seededRand.Int())[0:9],
			Duration: createdSession.GetDuration(),
		},
		AccessToken: zmToken.GetAccessToken(),
	})
	if zmErr != nil {
		return nil, zmErr
	}

	updateSessionWithZoom := &pb.Session{
		Id:                  createdSession.GetId(),
		ZoomMeetingId:       &pb.OptionalInt64{Value: zm.GetId()},
		ZoomMeetingJoinUrl:  &pb.OptionalString{Value: zm.GetJoinUrl()},
		ZoomMeetingStartUrl: &pb.OptionalString{Value: zm.GetStartUrl()},
		MeetingUrl:          &pb.OptionalString{Value: zm.GetJoinUrl()},
		ZoomSyncEnabled:     &pb.OptionalBool{Value: true},
		ZoomMeetingType:     utils.ZoomMeetingTypeSingular,
	}

	update, updateErr := s.UpdateSession(ctx, updateSessionWithZoom)
	if updateErr != nil {
		return nil, err
	}
	return update, nil

}

// Gets a Session by ID.
func (s *Server) GetSessionById(ctx context.Context, in *pb.GetSessionByIdRequest) (*pb.Session, error) {
	c, err := s.repo.GetSessionById(uint(in.GetId()))
	if err != nil {
		return nil, err
	}
	hydrated := s.hydrateSession(c.ProtoMarshalPrivate())

	if b, bErr := s.repo.IsSessionIsPurchasedByUser(uint(in.GetUserId()), uint(in.GetId())); bErr == nil {
		hydrated.Purchased = b
	}
	return hydrated, nil
}

// Gets a GetUpcomingSessionsByUserIdAndDate by User Id and date.
// can be used for pagination
func (s *Server) GetUpcomingSessionsByUserIdAndDate(ctx context.Context, in *pb.SessionsByIdAndDateRequest) (*pb.Sessions, error) {
	c, err := s.repo.GetUpcomingSessionsByUserIdAndDate(uint(in.GetUserId()), in.GetDate())
	if err != nil {
		return nil, err
	}
	hydrated, hErr := s.hydrateSessions(ctx, c.ProtoMarshalPrivate())
	if hErr != nil {
		return nil, hErr
	}
	return hydrated, nil
}

func (s *Server) UpdateSessionsNoZoomSyncByUserId(ctx context.Context, in *pb.Id) (*pb.Empty, error) {
	err := s.repo.UpdateSessionsNoZoomSyncByUserId(in.GetId())
	return &pb.Empty{}, err
}

// Gets a GetUpcomingSessionsByUserIdAndDate by User Id and date.
// can be used for pagination
func (s *Server) GetPreviousSessionsByUserIdAndDate(ctx context.Context, in *pb.SessionsByIdAndDateRequest) (*pb.Sessions, error) {
	c, err := s.repo.GetPreviousSessionsByUserIdAndDate(uint(in.GetUserId()), in.GetDate())
	if err != nil {
		return nil, err
	}
	hydrated, hErr := s.hydrateSessions(ctx, c.ProtoMarshalPrivate())
	if hErr != nil {
		return nil, hErr
	}
	return hydrated, nil
}

func (s *Server) GetSessionByUserIdAndZoomMeetingId(ctx context.Context, in *pb.GetSessionByUserIdAndZoomMeetingIdRequest) (*pb.Session, error) {
	c, err := s.repo.GetSessionByUserIdAndZoomMeetingId(in.GetUserId(), in.GetZoomId())
	if err != nil {
		return nil, err
	}
	hydrated := s.hydrateSession(c.ProtoMarshalPrivate())
	return hydrated, nil
}

func (s *Server) GetSessionByUserIdAndZoomMeetingIdAndOccurrenceId(ctx context.Context, in *pb.GetSessionByUserIdAndZoomMeetingIdAndOccurrenceIdRequest) (*pb.Session, error) {
	c, err := s.repo.GetSessionByUserIdAndZoomMeetingIdAndOccurrenceId(in.GetUserId(), in.GetZoomId(), in.GetOccurrenceId())
	if err != nil {
		return nil, err
	}
	hydrated := s.hydrateSession(c.ProtoMarshalPrivate())
	return hydrated, nil
}

func (s *Server) GetSessionByUserIdAndZoomMeetingIdAndRawPayloadId(ctx context.Context, in *pb.GetSessionByUserIdAndZoomMeetingIdAndRawPayloadIdRequest) (*pb.Session, error) {
	c, err := s.repo.GetSessionByUserIdAndZoomMeetingIdAndRawPayloadId(in.GetUserId(), in.GetZoomId(), in.GetRawPayload())
	if err != nil {
		return nil, err
	}
	hydrated := s.hydrateSession(c.ProtoMarshalPrivate())
	return hydrated, nil
}

// deletes a Session by ID.
func (s *Server) DeleteSessionById(ctx context.Context, in *pb.GetSessionByIdRequest) (*pb.Session, error) {
	existing, gtErr := s.GetSessionById(ctx, in)
	if gtErr != nil {
		return nil, gtErr
	}

	err := s.repo.DeleteSessionById(uint(in.GetId()))
	if err != nil {
		return nil, err
	}
	if existing.GetZoomSyncEnabled().GetValue() {
		deleteErr := s.deleteMeetingInZoom(ctx, existing)
		if deleteErr != nil {
			log.Println("error deleting meeting in zoom, err", deleteErr)
		}
	}
	return &pb.Session{}, nil
}

// deletes a Session by Zoom Meeting ID.
func (s *Server) DeleteSessionsByZoomMeetingId(ctx context.Context, in *pb.DeleteSessionsByZoomMeetingIdRequest) (*pb.Empty, error) {

	err := s.repo.DeleteSessionsByUserIdAndZoomMeetingId(uint(in.GetUserId()), uint(in.GetZoomMeetingId()))
	if err != nil {
		return nil, err
	}
	// Note this endpoint is meant to be hit upon zoom webhook deletion so no need to sync to zoom
	return &pb.Empty{}, nil
}

func (s *Server) GetSessionsByStartDate(ctx context.Context, in *pb.SessionsRequest) (*pb.Sessions, error) {
	sess, err := s.repo.GetSessionsByStartDate(in.GetDate(), int(in.GetLimit()))
	if err != nil {
		return nil, err
	}
	return s.hydrateSessions(ctx, sess.ProtoMarshalPublic())
}

func (s *Server) GetSessionsByStartDateAndTag(ctx context.Context, in *pb.SessionsByTagRequest) (*pb.Sessions, error) {
	sess, err := s.repo.GetSessionsByStartDateAndTag(in.GetDate(), int(in.GetLimit()), in.GetTag())
	if err != nil {
		return nil, err
	}
	return s.hydrateSessions(ctx, sess.ProtoMarshalPublic())
}

func (s *Server) GetSessionsByUserId(ctx context.Context, in *pb.Id) (*pb.Sessions, error) {
	u, err := s.repo.GetSessionsByUserId(uint(in.GetId()))
	if err != nil {
		return nil, err
	}
	return s.hydrateSessions(ctx, u.ProtoMarshalPrivate())
}

func (s *Server) hydrateSessions(ctx context.Context, hydrated *pb.Sessions) (*pb.Sessions, error) {
	// Hydrate users
	userIds := []uint{}
	for _, c := range hydrated.GetSessions() {
		userIds = append(userIds, uint(c.GetUserId()))
	}

	usersMap := map[uint]*models.User{}
	users, uErr := s.repo.GetUsersById(userIds)
	if uErr != nil {
		return nil, uErr
	}
	for _, u := range users.Users {
		usersMap[u.ID] = u
	}

	out := []*pb.Session{}
	for _, i := range hydrated.GetSessions() {
		if u, uOk := usersMap[uint(i.GetUserId())]; uOk {
			i.User = u.ProtoMarshalPublic()
		}
		out = append(out, i)
	}

	return &pb.Sessions{Sessions: out}, nil

}

func (s *Server) hydrateSession(hydrated *pb.Session) *pb.Session {
	if sessions, err := s.hydrateSessions(context.Background(), &pb.Sessions{Sessions: []*pb.Session{hydrated}}); err == nil {
		if len(sessions.GetSessions()) == 1 && sessions.GetSessions()[0] != nil {
			return sessions.GetSessions()[0]
		}
	}
	log.Println("Err hydrating session, returning unhydrated")
	return hydrated

}

func (s *Server) updateMeetingInZoom(ctx context.Context, in *pb.Session) error {
	if in.GetZoomMeetingId().GetValue() <= 0 {
		return errors.New("error, can not update meeting in zoom with invalid zoom meeting id")
	}
	token, tErr := s.GetZoomTokenByUserId(ctx, &pb.Id{Id: in.GetUserId()})
	if tErr != nil {
		return tErr
	}

	meetingPayloadForUpdate := utils.UpdateMeetingInZoomPayloadFromSession(in)
	_, updateErr := s.UpdateMeetingInZoom(ctx, &pb.UpdateMeetingInZoomRequest{
		Fields:       meetingPayloadForUpdate,
		AccessToken:  token.GetAccessToken(),
		MeetingId:    in.GetZoomMeetingId().GetValue(),
		OccurrenceId: in.GetZoomOccurrenceId(),
	})

	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (s *Server) deleteMeetingInZoom(ctx context.Context, in *pb.Session) error {
	if in.GetZoomMeetingId().GetValue() <= 0 {
		return errors.New("error, can not delete meeting in zoom with invalid zoom meeting id")
	}
	token, tErr := s.GetZoomTokenByUserId(ctx, &pb.Id{Id: in.GetUserId()})
	if tErr != nil {
		return tErr
	}

	_, deleteErr := s.DeleteMeetingInZoom(ctx, &pb.DeleteMeetingInZoomRequest{
		AccessToken:  token.GetAccessToken(),
		MeetingId:    in.GetZoomMeetingId().GetValue(),
		OccurrenceId: in.GetZoomOccurrenceId(),
	})

	if deleteErr != nil {
		return deleteErr
	}

	return nil
}

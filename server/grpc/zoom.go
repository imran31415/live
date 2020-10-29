package grpc

import (
	"admin/models"
	pb "admin/protos"
	"admin/server/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

func (s *Server) GetZoomAppInstallUrl(ctx context.Context, in *pb.ZoomAppInstallUrlRequest) (*pb.ZoomAppInstallUrlResponse, error) {
	p, err := s.GetUserByIdPrivate(ctx, &pb.Id{Id: in.GetUserId()})
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("https://zoom.us/oauth/authorize?client_id=%s&response_type=code&redirect_uri=%s&state=%d", s.zoomClientKey, url.QueryEscape(s.zoomRedirectUri), p.GetId())
	return &pb.ZoomAppInstallUrlResponse{Url: u}, nil
}

func (s *Server) CreateZoomAccessToken(ctx context.Context, in *pb.ZoomToken) (*pb.ZoomToken, error) {

	toCreate := &models.ZoomToken{}
	toCreate.ProtoUnMarshal(in)
	u, err := s.repo.CreateZoomAccessToken(toCreate)
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshal(), nil
}

func (s *Server) GetZoomTokenById(ctx context.Context, in *pb.Id) (*pb.ZoomToken, error) {
	z, err := s.repo.GetZoomTokenById(uint(in.GetId()))
	if err != nil {
		return nil, err
	}
	return z.ProtoMarshal(), nil
}

func (s *Server) UpdateZoomTokenById(ctx context.Context, in *pb.ZoomToken) (*pb.ZoomToken, error) {
	toUpdate := &models.ZoomToken{}
	toUpdate.ProtoUnMarshal(in)
	u, err := s.repo.UpdateZoomTokenById(toUpdate)
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshal(), nil
}

func (s *Server) DeleteZoomAccessTokensByUserId(ctx context.Context, id *pb.Id) (*pb.Empty, error) {
	err := s.repo.DeleteZoomAccessTokensByUserId(id.GetId())
	return &pb.Empty{}, err
}

func (s *Server) GetZoomTokenByUserId(ctx context.Context, in *pb.Id) (*pb.ZoomToken, error) {
	z, err := s.repo.GetZoomTokenByUserId(uint(in.GetId()))
	if err != nil {
		return nil, err
	}

	refreshed, refreshedErr := s.RefreshZoomToken(ctx, &pb.Id{Id: z.ProtoMarshal().GetId()})
	if refreshedErr != nil {
		return nil, refreshedErr
	}
	return refreshed, nil
}

func (s *Server) GetZoomRedirectSuccessUri(ctx context.Context, in *pb.Empty) (*pb.RedirectUri, error) {
	return &pb.RedirectUri{Uri: s.zoomreDirectSuccessUri}, nil
}

func (s *Server) RefreshZoomToken(ctx context.Context, in *pb.Id) (*pb.ZoomToken, error) {
	existingToken, err := s.GetZoomTokenById(context.Background(), in)
	if err != nil {
		return nil, err
	}

	if existingToken.GetRefreshToken() == "" {
		return nil, fmt.Errorf("existing token has empty refresh token: %+v", existingToken)
	}

	refreshedToken, refreshErr := s.getRefreshedZoomToken(existingToken.GetRefreshToken())

	if refreshErr != nil {
		return nil, refreshErr
	}
	updateToken := refreshedToken.ProtoMarshal()
	updateToken.Id = existingToken.Id
	updateToken.UserId = existingToken.UserId

	updatedToken, updateErr := s.UpdateZoomTokenById(context.Background(), updateToken)
	if updateErr != nil {
		return nil, updateErr
	}
	return updatedToken, nil

}

func (s *Server) CreateMeetingInZoom(ctx context.Context, in *pb.CreateMeetingInZoomRequest) (*pb.CreateMeetingInZoomResponse, error) {
	u := "https://api.zoom.us/v2/users/me/meetings"
	in.Fields.Settings = &pb.CreateZoomMeetingFields_Settings{}

	// Defaulted fields for settings
	in.Fields.Settings.HostVideo = "true"
	in.Fields.Settings.ParticipantVideo = "false"
	in.Fields.Settings.JoinBeforeHost = "false"
	in.Fields.Settings.MuteUponEntry = "true"
	in.Fields.Settings.WaitingRoom = "false"

	jsonStr, mErr := json.Marshal(in.Fields)
	if mErr != nil {
		return nil, mErr
	}

	req, err := http.NewRequest("POST", u, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", in.GetAccessToken()))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, bErr := ioutil.ReadAll(resp.Body)
	if bErr != nil {
		return nil, bErr
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		return nil, fmt.Errorf("received status code: %d from zoom, err: %s", resp.StatusCode, string(body))
	}
	z := &utils.CreateZoomMeetingApiResponse{}

	if err = json.Unmarshal(body, z); err != nil {
		return nil, err
	}

	createResp := &pb.CreateMeetingInZoomResponse{
		Id:       z.ID,
		StartUrl: z.StartURL,
		JoinUrl:  z.JoinURL,
	}

	return createResp, nil
}

func (s *Server) UpdateMeetingInZoom(ctx context.Context, in *pb.UpdateMeetingInZoomRequest) (*pb.UpdateMeetingInZoomResponse, error) {
	u := fmt.Sprintf("https://api.zoom.us/v2/meetings/%d", in.GetMeetingId())

	if in.GetOccurrenceId() != "" {
		u += fmt.Sprintf("?occurrence_id=%s", in.GetOccurrenceId())
	}

	jsonStr, mErr := json.Marshal(in.Fields)
	if mErr != nil {
		return nil, mErr
	}

	req, err := http.NewRequest("PATCH", u, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", in.GetAccessToken()))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, bErr := ioutil.ReadAll(resp.Body)
	if bErr != nil {
		return nil, bErr
	}

	if resp.StatusCode != 204 {
		return nil, fmt.Errorf("patch error received status code: %d from zoom, err: %s", resp.StatusCode, string(body))
	}
	return &pb.UpdateMeetingInZoomResponse{}, nil
}

func (s *Server) DeleteMeetingInZoom(ctx context.Context, in *pb.DeleteMeetingInZoomRequest) (*pb.DeleteMeetingInZoomResponse, error) {
	u := fmt.Sprintf("https://api.zoom.us/v2/meetings/%d", in.GetMeetingId())
	if in.GetOccurrenceId() != "" {
		// todo: better way to add query params but since we just need one this works for now
		u = fmt.Sprintf("%s?occurrence_id=%s", u, in.GetOccurrenceId())
	}
	empty := &struct{}{}
	jsonStr, _ := json.Marshal(empty)
	req, err := http.NewRequest("DELETE", u, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", in.GetAccessToken()))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, bErr := ioutil.ReadAll(resp.Body)
	if bErr != nil {
		return nil, bErr
	}

	if resp.StatusCode != 204 {
		return nil, fmt.Errorf("delete error received status code: %d from zoom, err: %s", resp.StatusCode, string(body))
	}
	return &pb.DeleteMeetingInZoomResponse{}, nil
}

func (s *Server) SyncZoomMeetingIds(ctx context.Context, in *pb.SyncZoomMeetingIdsRequest) (*pb.SyncZoomMeetingIdsResponse, error) {
	errs := []string{}
	success := int64(0)
	time.Sleep(time.Second * 2)
	zt, err := s.GetZoomTokenByUserId(ctx, &pb.Id{Id: in.GetUserId()})
	if err != nil {
		return nil, err
	}
	for _, i := range in.GetMeetingIds() {
		// Note this is running in a go routine and not blocking the response, so lets rate limit our syncs to 1 per second
		time.Sleep(time.Second)
		meetingInfo, miErr := utils.GetMeetingInfo(i, zt.GetAccessToken())
		if miErr != nil {
			log.Println("Err getting meeting info in SyncZoomMeetingIds", miErr)
			errs = append(errs, miErr.Error())
			continue
		}

		innerSuccesses, innerErrs := s.createSessionFromZoomMeetingInfo(meetingInfo, in.GetUserId())
		if len(innerErrs) > 0 {
			errs = append(errs, innerErrs...)
		}
		success += innerSuccesses
	}

	if len(errs) > 0 {
		return &pb.SyncZoomMeetingIdsResponse{
			Errs:      errs,
			Successes: success,
		}, fmt.Errorf("zoom sync errors %+v", errs)
	}
	return &pb.SyncZoomMeetingIdsResponse{
		Errs:      errs,
		Successes: success,
	}, nil
}

func (s *Server) SyncUsersZoomMeeting(ctx context.Context, id *pb.Id) (*pb.Empty, error) {

	token, tErr := s.GetZoomTokenByUserId(ctx, id)

	if tErr != nil {
		return nil, tErr
	}
	lastPage := false
	index := 0

	errs := []string{}
	totalSuccesses := int64(0)

	for !lastPage {
		index += 1

		u := fmt.Sprintf("https://api.zoom.us/v2/users/me/meetings?type=upcoming&page_size=300&page_number=%d", index)

		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			break
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.GetAccessToken()))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errs = append(errs, err.Error())
			log.Println("Err in request to zoom, ", err)
			break
		}
		defer resp.Body.Close()
		body, bErr := ioutil.ReadAll(resp.Body)
		if bErr != nil {
			errs = append(errs, bErr.Error())
			log.Println("Err reading body from zoom, ", bErr)
			break
		}

		if resp.StatusCode != 200 {
			errs = append(errs, fmt.Errorf("get meetings error received status code: %d from zoom, err: %s", resp.StatusCode, string(body)).Error())
			break
		}

		meetingsPaginated := &utils.ZoomMeetingsPaginatedResponse{}

		if err = json.Unmarshal(body, meetingsPaginated); err != nil {
			errs = append(errs, err.Error())
			log.Println("err unmarshalling json: ", err)
			break
		}

		if meetingsPaginated.TotalRecords <= meetingsPaginated.PageCount*meetingsPaginated.PageSize {
			lastPage = true
		}

		meetingIds := []int64{}
		for _, i := range meetingsPaginated.Meetings {
			meetingIds = append(meetingIds, int64(i.ID))
		}
		syncResp, _ := s.SyncZoomMeetingIds(ctx, &pb.SyncZoomMeetingIdsRequest{
			MeetingIds: meetingIds,
			UserId:     id.GetId(),
		})
		totalSuccesses += syncResp.GetSuccesses()
		if len(syncResp.GetErrs()) > 0 {
			errs = append(errs, syncResp.GetErrs()...)
		}

	}
	log.Printf("Errs:%d, Successes: %d ", len(errs), totalSuccesses)
	if len(errs) > 0 {
		log.Println("Errs received: ")
		for _, e := range errs {
			log.Println("err syncing meeting, err: ", e)
		}
		return &pb.Empty{}, fmt.Errorf("errors syncing zoom meetings, pages with errors encountered: %d", len(errs))
	}
	return &pb.Empty{}, nil

}

func (s *Server) getRefreshedZoomToken(refreshToken string) (*models.ZoomToken, error) {

	tries := 0
	var err error
	var z *models.ZoomToken

	for tries < 2 {
		u := fmt.Sprintf("https://zoom.us/oauth/token?grant_type=refresh_token&refresh_token=%s", refreshToken)

		var jsonStr = []byte(``)
		req, er := http.NewRequest("POST", u, bytes.NewBuffer(jsonStr))
		if er != nil {
			tries += 1
			err = er
			continue
		}
		req.SetBasicAuth(s.zoomClientKey, s.zoomClientSecret)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, er := client.Do(req)
		if er != nil {
			tries += 1
			err = er
			continue
		}
		defer resp.Body.Close()
		body, er := ioutil.ReadAll(resp.Body)
		if er != nil {
			tries += 1
			err = er
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode > 300 {
			err = fmt.Errorf("received status code: %d from zoom, err: %s", resp.StatusCode, string(body))
			tries += 1
			continue
		}
		z = &models.ZoomToken{}
		if err = json.Unmarshal(body, z); err != nil {
			tries += 1
			continue
		}
		if z.AccessToken == "" {
			err = fmt.Errorf("zoom token has empty access token, %+v", z)
			tries += 1
			continue
		}
		return z, nil
	}

	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("unexpected end of func")
}

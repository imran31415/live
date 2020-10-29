package http

import (
	pb "admin/protos"
	"admin/server/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var ZoomAppInstallHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	log.Println("Received ZoomAppInstallHandler webhook")

	code := req.URL.Query().Get("code")
	if code == "" {
		log.Println("url parameter: code is required")
		http.Error(w, "url parameter: code is required", http.StatusInternalServerError)
		return
	}
	// state param passed is the user_id by convention
	s := req.URL.Query().Get("state")

	if s == "" {
		log.Println("url parameter: state is required")
		http.Error(w, "url parameter: state is required", http.StatusInternalServerError)
		return
	}
	userId, err := strconv.Atoi(s)

	if err != nil {
		log.Println("Err converting string to int", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, pErr := httpClient.grpcClient.GetUserById(context.Background(), &pb.Id{Id: int64(userId)})
	if pErr != nil {
		log.Println("Err getting user by ID: ", pErr)
		http.Error(w, pErr.Error(), http.StatusInternalServerError)
		return
	}

	saveZoomToken(w, req, code, grantTypeAuthorize, grantAuthorize, int64(userId))
	// asynchronously sync the users meeting in the background
	log.Println("Saved token for install, syncing upcoming meetings in zoom in the background")
	go func() {
		if _, err = httpClient.grpcClient.SyncUsersZoomMeeting(context.Background(), &pb.Id{Id: int64(userId)}); err != nil {
			log.Println("errors encountered in SyncUsersZoomMeeting")
		} else {
			log.Println("success SyncUsersZoomMeeting")
		}
	}()

	redirectZoomSuccess(w, req)

	return
})

var ZoomMeetingWebHook = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	log.Println("Received ZoomMeetingWebHook webhook")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("Unable to read body of request: ", req.Body)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	we := &utils.WebHookEvent{}

	err = json.Unmarshal(body, we)
	// TOdo refactor, only unmarshal event type and to specific unmarshalling within the processWebhookCRUD specific methods.

	if err != nil {
		log.Println("Error ZoomMeetingWebHook Unmarshal: ", err, " Body is: ", string(body))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch we.Event {
	case "meeting.created":
		err = processWebHookMeetingCreated(w, we)
	case "meeting.updated":
		err = processWebHookMeetingUpdated(w, we)
	case "meeting.deleted":
		err = processWebHookMeetingDeleted(w, we)
	default:
		log.Println("Received unhandled webhook event: ", we.Event, " not processing, returning success")
	}

	if err == nil {
		w.WriteHeader(http.StatusNoContent)
	} else {
		log.Println("err processing webhook request from zoom: ", err, "for event: ", we.Event)
		w.WriteHeader(http.StatusInternalServerError)
	}
	return

})

var ZoomDeAuthorizeWebHook = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	log.Println("Received ZoomDeAuthorizeWebHook webhook")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("Unable to read body of request: ", req.Body)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	we := &utils.ZoomAccountDeauthorizedPayload{}

	err = json.Unmarshal(body, we)

	if err != nil {
		log.Println("Error ZoomDeAuthorizeWebHook Unmarshal: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var processErr error
	switch we.Event {
	case "app_deauthorized":
		processErr = processWebHookAppDeAuthorized(w, we)
	default:
		log.Println("Received unhandled account deauthorized webhook event: ", we.Event, " not processing")
	}

	if processErr != nil {
		log.Println("Received error processing event: ", we.Event, "Error: ", processErr)
		http.Error(w, processErr.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return

})

func processWebHookMeetingCreated(w http.ResponseWriter, we *utils.WebHookEvent) error {
	u, uErr := httpClient.grpcClient.GetUserByZoomAccountId(context.Background(), &pb.ZoomAccountId{Id: we.Payload.AccountID})
	if uErr != nil {
		log.Println("Error Getting User from webhook Account Id: ", uErr)
		http.Error(w, uErr.Error(), http.StatusInternalServerError)
		// since we can get a network error here we can retry
		return uErr
	}

	// Errors are reported in logs and since its a webhook we dont want to return it
	_, err := httpClient.grpcClient.SyncZoomMeetingIds(context.Background(), &pb.SyncZoomMeetingIdsRequest{
		MeetingIds: []int64{we.Payload.Object.ID},
		UserId:     u.GetId(),
	})
	if err != nil {

		if strings.Contains(err.Error(), "is not found") {
			// the meeting doesn't exist in zoom, do not return error since no reason to retry
			return nil
		}
		log.Println("Err syncing meetings on create: ", err)
		return err
	}

	return nil
}

func processWebHookMeetingUpdated(w http.ResponseWriter, we *utils.WebHookEvent) error {
	u, uErr := httpClient.grpcClient.GetUserByZoomAccountId(context.Background(), &pb.ZoomAccountId{Id: we.Payload.AccountID})
	if uErr != nil {
		log.Println("Error Getting User from webhook Account Id: ", uErr)
		http.Error(w, uErr.Error(), http.StatusInternalServerError)
		// since we can get a network error here we can retry
		return uErr
	}

	if _, uuErr := httpClient.grpcClient.SyncZoomMeetingIds(context.Background(), &pb.SyncZoomMeetingIdsRequest{
		MeetingIds: []int64{we.Payload.Object.ID},
		UserId:     u.GetId(),
	}); uuErr != nil {
		if strings.Contains(uuErr.Error(), "is not found") {
			// the meeting doesn't exist in zoom, do not return error since no reason to retry
			return nil
		}
		log.Println("Err syncing meetings on update: ", uuErr)
		return uuErr
	}

	log.Println("Successfully synced meeting:  ", we.Payload.Object.ID)
	return nil
}

func processWebHookMeetingDeleted(w http.ResponseWriter, we *utils.WebHookEvent) error {
	switch we.Payload.Object.Type {
	case utils.ZoomMeetingTypeReocurring:
		if we.Payload.Object.Occurrences != nil && len(we.Payload.Object.Occurrences) > 0 {
			errs := []error{}
			for _, o := range we.Payload.Object.Occurrences {
				u, uErr := httpClient.grpcClient.GetUserByZoomAccountId(context.Background(), &pb.ZoomAccountId{Id: we.Payload.AccountID})
				if uErr != nil {
					log.Println("Error Getting User from webhook Account Id: ", uErr)
					errs = append(errs, uErr)
				}

				existingSession, exErr := httpClient.grpcClient.GetSessionByUserIdAndZoomMeetingIdAndOccurrenceId(context.Background(), &pb.GetSessionByUserIdAndZoomMeetingIdAndOccurrenceIdRequest{
					UserId:       u.GetId(),
					ZoomId:       we.Payload.Object.ID,
					OccurrenceId: o.OccurrenceID,
				})

				if exErr != nil {
					if strings.Contains(exErr.Error(), gorm.ErrRecordNotFound.Error()) {
						return nil
					}
					log.Println("err encountered finding existing session for delete in zoom webhook")
					errs = append(errs, exErr)
				}
				_, cErr := httpClient.grpcClient.DeleteSessionById(context.Background(), &pb.GetSessionByIdRequest{Id: existingSession.GetId()})
				if cErr != nil {
					log.Println("Error Deleting meeting from zoom webhook ", cErr)
					errs = append(errs, cErr)
				}

				log.Println("Successfully processWebHookMeetingDeleted for OccurrenceID id: ", o.OccurrenceID, ", meeting id: ", existingSession.GetId(), " for user: ", existingSession.GetUserId())
			}
			if len(errs) > 0 {
				log.Println("Errs encountered deleting occurrences in zoom webhook delete")
				return fmt.Errorf("errs encountered deleting occurrences in zoom webhook delete: %+v", errs)
			}
			return nil

		} else {
			// This is a series of reocurring meetings getting deleted
			u, uErr := httpClient.grpcClient.GetUserByZoomAccountId(context.Background(), &pb.ZoomAccountId{Id: we.Payload.AccountID})
			if uErr != nil {
				log.Println("Error Getting User from webhook Account Id: ", uErr)
				http.Error(w, uErr.Error(), http.StatusInternalServerError)
				return uErr
			}

			_, cErr := httpClient.grpcClient.DeleteSessionsByZoomMeetingId(context.Background(), &pb.DeleteSessionsByZoomMeetingIdRequest{
				ZoomMeetingId: we.Payload.Object.ID,
				UserId:        u.GetId(),
			})
			if cErr != nil {
				log.Println("Error Deleting meeting in zoom webhook ", cErr)
				http.Error(w, cErr.Error(), http.StatusInternalServerError)
				return cErr
			}

			log.Println("Successfully processWebHookMeetingDeleted, zoom meeting id: ", we.Payload.Object.ID, " for user: ", u.GetId())
			return nil
		}
	case utils.ZoomMeetingTypeSingular:
		u, uErr := httpClient.grpcClient.GetUserByZoomAccountId(context.Background(), &pb.ZoomAccountId{Id: we.Payload.AccountID})
		if uErr != nil {
			log.Println("Error Getting User from webhook Account Id: ", uErr)
			http.Error(w, uErr.Error(), http.StatusInternalServerError)
			return uErr
		}

		existingSession, exErr := httpClient.grpcClient.GetSessionByUserIdAndZoomMeetingId(context.Background(), &pb.GetSessionByUserIdAndZoomMeetingIdRequest{
			UserId: u.GetId(),
			ZoomId: we.Payload.Object.ID,
		})

		if exErr != nil {
			if strings.Contains(exErr.Error(), gorm.ErrRecordNotFound.Error()) {
				return nil
			}
			log.Println("err encountered finding existing session for delete in zoom webhook")
			return exErr
		}

		_, cErr := httpClient.grpcClient.DeleteSessionById(context.Background(), &pb.GetSessionByIdRequest{Id: existingSession.GetId()})
		if cErr != nil {
			log.Println("Error Deleting meeting from zoom webhook creation ", cErr)
			http.Error(w, cErr.Error(), http.StatusInternalServerError)
			return cErr
		}

		log.Println("Successfully processWebHookMeetingDeleted, meeting id: ", existingSession.GetId(), " for user: ", existingSession.GetUserId())
		return nil
	default:
		log.Println("skipping zoom webhook deletion of type of ", we.Payload.Object.Type, " received and i")
		return nil

	}

}

func processWebHookAppDeAuthorized(w http.ResponseWriter, event *utils.ZoomAccountDeauthorizedPayload) error {

	u, uErr := httpClient.grpcClient.GetUserByZoomAccountId(context.Background(), &pb.ZoomAccountId{Id: event.Payload.AccountID})
	if uErr != nil {
		log.Println("Error Getting User from webhook deauthorize Account Id: ", uErr)
		http.Error(w, uErr.Error(), http.StatusInternalServerError)
		return uErr
	}

	_, dErr := httpClient.grpcClient.DeleteZoomAccessTokensByUserId(context.Background(), &pb.Id{Id: u.GetId()})
	if dErr != nil {
		log.Println("err DeleteZoomAccessTokensByUserId: ", dErr)
	}

	_, upErr := httpClient.grpcClient.UpdateSessionsNoZoomSyncByUserId(context.Background(), &pb.Id{Id: u.GetId()})
	if upErr != nil {
		if strings.Contains(upErr.Error(), gorm.ErrRecordNotFound.Error()) {
			// not an error since there just were no sessions to update
			upErr = nil
		} else {
			log.Println("err UpdateSessionsNoZoomSyncByUserId", upErr)
		}
	}

	_, upUserErr := httpClient.grpcClient.UpdateUserZoomDeAuthorized(context.Background(), &pb.Id{Id: u.GetId()})
	if upUserErr != nil {
		log.Println("err UpdateUserZoomDeAuthorized", upUserErr)
	}

	if dErr != nil || upErr != nil || upUserErr != nil {
		log.Println("completed de authorization, with server errors")
	} else {
		log.Println("completed de authorization with no errors")
	}

	return nil
}

package http

import (
	pb "admin/protos"
	"context"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"io/ioutil"
	"log"
	"net/http"
)

var GetSessionByIdHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	authZeroSubId := extractAuthZeroSubId(r)
	if authZeroSubId == "" {
		http.Error(w, "Unable to parse auth zero sub id from request", http.StatusBadRequest)
		return
	}

	u, err := httpClient.grpcClient.GetOrCreateUserBySubId(context.Background(), &pb.SubId{Id: authZeroSubId})
	if err != nil {
		log.Println("error in GetOrCreateUserBySubId ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	in := &pb.GetSessionByIdRequest{}

	body, rErr := ioutil.ReadAll(r.Body)
	if rErr != nil {
		log.Println("Unable to read body of request: ", r.Body)
		http.Error(w, rErr.Error(), http.StatusBadRequest)
		return
	}

	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// set user id
	in.UserId = u.GetId()

	session, cErr := httpClient.grpcClient.GetSessionById(context.Background(), in)
	if cErr != nil {
		log.Println("Error GetSessionById: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
	}
	js, mErr := json.Marshal(session)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
	}
	writeJs(js, w, r)
})

var DeleteSessionById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	authZeroSubId := extractAuthZeroSubId(r)
	if authZeroSubId == "" {
		http.Error(w, "Unable to parse auth zero sub id from request", http.StatusBadRequest)
		return
	}

	u, err := httpClient.grpcClient.GetOrCreateUserBySubId(context.Background(), &pb.SubId{Id: authZeroSubId})
	if err != nil {
		log.Println("error in GetOrCreateUserBySubId ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	in := &pb.GetSessionByIdRequest{}

	body, rErr := ioutil.ReadAll(r.Body)
	if rErr != nil {
		log.Println("Unable to read body of request: ", r.Body)
		http.Error(w, rErr.Error(), http.StatusBadRequest)
		return
	}

	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session, cErr := httpClient.grpcClient.GetSessionById(context.Background(), in)
	if cErr != nil {
		log.Println("Error GetSessionById in DeleteSessionByID: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}

	if session.GetUserId() != u.GetId() {
		log.Println("Unauthorized user to create session")
		http.Error(w, "requesting user is not the same user as UserId set to session", http.StatusUnauthorized)
		return

	}

	dResp, dErr := httpClient.grpcClient.DeleteSessionById(context.Background(), in)
	if dErr != nil {
		log.Println("Error DeleteSessionById: ", dErr)
		http.Error(w, dErr.Error(), http.StatusInternalServerError)
		return
	}

	js, mErr := json.Marshal(dResp)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

var UpdateSession = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	authZeroSubId := extractAuthZeroSubId(r)
	if authZeroSubId == "" {
		http.Error(w, "Unable to parse auth zero sub id from request", http.StatusBadRequest)
		return
	}

	u, err := httpClient.grpcClient.GetOrCreateUserBySubId(context.Background(), &pb.SubId{Id: authZeroSubId})
	if err != nil {
		log.Println("error in GetOrCreateUserBySubId ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	in := &pb.Session{}

	body, rErr := ioutil.ReadAll(r.Body)
	if rErr != nil {
		log.Println("Unable to read body of request: ", r.Body)
		http.Error(w, rErr.Error(), http.StatusBadRequest)
		return
	}

	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session, cErr := httpClient.grpcClient.GetSessionById(context.Background(), &pb.GetSessionByIdRequest{
		Id:     in.GetId(),
		UserId: u.GetId(),
	})
	if cErr != nil {
		log.Println("Error GetSessionById in UpdateSession: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
	}

	if session.GetUserId() != u.GetId() {
		log.Println("Unauthorized user to create session")
		http.Error(w, "requesting user is not the same user as UserId set to session", http.StatusUnauthorized)
		return

	}

	// Since this is an update sourced from an HTTP request from via.live we want to sync the meeting to zoom if its enabled for the session
	in.ShouldZoomSyncOnUpdate = true

	updateSession, dErr := httpClient.grpcClient.UpdateSession(context.Background(), in)
	if dErr != nil {
		log.Println("Error UpdateSession: ", dErr)
		http.Error(w, dErr.Error(), http.StatusInternalServerError)
	}

	js, mErr := json.Marshal(updateSession)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
	}
	writeJs(js, w, r)
})

var GetSessionByIdNoAuth = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	in := &pb.GetSessionByIdRequest{}

	body, rErr := ioutil.ReadAll(r.Body)
	if rErr != nil {
		log.Println("Unable to read body of request: ", r.Body)
		http.Error(w, rErr.Error(), http.StatusBadRequest)
		return
	}

	if err := jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session, cErr := httpClient.grpcClient.GetSessionById(context.Background(), in)
	if cErr != nil {
		log.Println("Error GetSessionById: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
	}
	js, mErr := json.Marshal(session)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
	}
	writeJs(js, w, r)
})

var GetSessionsByDate = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	in := &pb.SessionsRequest{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Unable to read request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sessions, cErr := httpClient.grpcClient.GetSessionsByStartDate(context.Background(), in)
	if cErr != nil {
		log.Println("Error GetSessionsByStartDate: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}
	js, mErr := json.Marshal(sessions)
	if mErr != nil {
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

var VerifyZoom = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("811aafe5992d411a9ee743e00d746c77"))
	w.WriteHeader(http.StatusOK)
	return
})

var GetSessionsByDateAndTag = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	in := &pb.SessionsByTagRequest{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Unable to read request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sessions, cErr := httpClient.grpcClient.GetSessionsByStartDateAndTag(context.Background(), in)
	if cErr != nil {
		log.Println("Error GetSessionsByStartDateAndTag: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}
	js, mErr := json.Marshal(sessions)
	if mErr != nil {
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

var GetUpcomingSessionsByUserIdAndDate = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	authZeroSubId := extractAuthZeroSubId(r)
	if authZeroSubId == "" {
		http.Error(w, "Unable to parse auth zero sub id from request", http.StatusBadRequest)
		return
	}
	// First get the user and ensure the user's user ID matches
	u, err := httpClient.grpcClient.GetOrCreateUserBySubId(context.Background(), &pb.SubId{Id: authZeroSubId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	in := &pb.SessionsByIdAndDateRequest{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Unable to read request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if u.GetId() != in.GetUserId() {
		log.Println("unauthorized")
		http.Error(w, "unauthorized", http.StatusBadRequest)
		return
	}

	sessions, cErr := httpClient.grpcClient.GetUpcomingSessionsByUserIdAndDate(context.Background(), in)
	if cErr != nil {
		log.Println("Error GetSessionsByStartDate: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}
	js, mErr := json.Marshal(sessions)
	if mErr != nil {
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

var GetPreviousSessionsByUserIdAndDate = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	authZeroSubId := extractAuthZeroSubId(r)
	if authZeroSubId == "" {
		http.Error(w, "Unable to parse auth zero sub id from request", http.StatusBadRequest)
		return
	}
	// First get the user and ensure the user's user ID matches
	u, err := httpClient.grpcClient.GetOrCreateUserBySubId(context.Background(), &pb.SubId{Id: authZeroSubId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	in := &pb.SessionsByIdAndDateRequest{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Unable to read request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if u.GetId() != in.GetUserId() {
		log.Println("unauthorized")
		http.Error(w, "unauthorized", http.StatusBadRequest)
		return
	}

	sessions, cErr := httpClient.grpcClient.GetPreviousSessionsByUserIdAndDate(context.Background(), in)
	if cErr != nil {
		log.Println("Error GetSessionsByStartDate: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}
	js, mErr := json.Marshal(sessions)
	if mErr != nil {
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

var CreateSession = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	authZeroSubId := extractAuthZeroSubId(r)
	if authZeroSubId == "" {
		http.Error(w, "Unable to parse auth zero sub id from request", http.StatusBadRequest)
		return
	}
	// First get the user and ensure the user's user ID matches
	u, err := httpClient.grpcClient.GetOrCreateUserBySubId(context.Background(), &pb.SubId{Id: authZeroSubId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	in := &pb.Session{}
	body, rErr := ioutil.ReadAll(r.Body)

	if rErr != nil {
		log.Println("Unable to read body of request: ", r.Body)
		http.Error(w, rErr.Error(), http.StatusBadRequest)
		return
	}
	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if u.GetId() != in.GetUserId() {
		log.Println("Unauthorized user to create session")
		http.Error(w, "requesting user is not the same user as UserId set to session", http.StatusUnauthorized)
		return
	}

	c, cErr := httpClient.grpcClient.CreateSession(context.Background(), in)
	if cErr != nil {
		log.Println("Error creating Session, err: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}

	js, mErr := json.Marshal(c)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

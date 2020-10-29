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

var GetUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	authZeroSubId := extractAuthZeroSubId(r)
	if authZeroSubId == "" {
		http.Error(w, "Unable to parse auth zero sub id from request", http.StatusBadRequest)
		return
	}
	user, cErr := httpClient.grpcClient.GetOrCreateUserBySubId(context.Background(), &pb.SubId{Id: authZeroSubId})
	if cErr != nil {
		log.Println("Error GetOrCreateUserBySubId: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}
	js, mErr := json.Marshal(user)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

var GetUserByIdHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	in := &pb.Id{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Unable to read body of request: ", r.Body)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, cErr := httpClient.grpcClient.GetUserById(context.Background(), in)
	if cErr != nil {
		log.Println("Error GetUser: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
	}
	js, mErr := json.Marshal(user)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
	}
	writeJs(js, w, r)
})

var GetUserByIdPublicHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	in := &pb.Id{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Unable to read body of request: ", r.Body)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, cErr := httpClient.grpcClient.GetUserByIdPublic(context.Background(), in)
	if cErr != nil {
		log.Println("Error GetUser: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
	}
	js, mErr := json.Marshal(user)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
	}
	writeJs(js, w, r)
})

var UpdateUserBySubId = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	authZeroSubId := extractAuthZeroSubId(r)
	if authZeroSubId == "" {
		http.Error(w, "Unable to parse auth zero sub id from request", http.StatusBadRequest)
		return
	}
	in := &pb.User{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Unable to read body of request: ", r.Body)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = jsonpb.UnmarshalString(string(body), in); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	in.AuthZeroSubId = authZeroSubId
	user, cErr := httpClient.grpcClient.UpdateUserBySubId(context.Background(), in)
	if cErr != nil {
		log.Println("Error UpdateUserBySubId: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
	}
	js, mErr := json.Marshal(user)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
	}
	writeJs(js, w, r)
})

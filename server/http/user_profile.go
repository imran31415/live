package http

import (
	pb "admin/protos"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

var GetUserProfile = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	profile, pErr := httpClient.grpcClient.GetUserProfile(context.Background(), &pb.Id{Id: user.GetId()})
	if pErr != nil {
		log.Println("Error GetUserProfile: ", pErr)
		http.Error(w, pErr.Error(), http.StatusInternalServerError)
		return
	}

	js, mErr := json.Marshal(profile)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

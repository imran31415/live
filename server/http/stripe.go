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

var GetStripeInstallUrl = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	authZeroSubId := extractAuthZeroSubId(r)
	if authZeroSubId == "" {
		http.Error(w, "Unable to parse auth zero sub id from request", http.StatusBadRequest)
		return
	}

	u, err := httpClient.grpcClient.GetOrCreateUserBySubId(context.Background(), &pb.SubId{Id: authZeroSubId})
	if err != nil {
		log.Println("Error getting user: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	in := &pb.StripeAppInstallUrlRequest{}
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

	if u.GetId() != in.GetUserId() {
		log.Println("user can not get perform this operation on the supplied user")
		http.Error(w, "user can not get perform this operation on the supplied user", http.StatusBadRequest)
		return
	}

	stripeInstallUrl, cErr := httpClient.grpcClient.GetStripeAppInstallUrl(context.Background(), in)
	if cErr != nil {
		log.Println("Error GetStripeAppInstallUrl: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
	}
	js, mErr := json.Marshal(stripeInstallUrl)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

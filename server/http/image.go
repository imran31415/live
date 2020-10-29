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

var GetImagesByUserId = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	images, cErr := httpClient.grpcClient.GetImagesByUserId(context.Background(), &pb.Id{Id: u.GetId()})
	if cErr != nil {
		log.Println("Error GetImagesByUserId: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
	}
	js, mErr := json.Marshal(images)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
	}
	writeJs(js, w, r)
})

var UpdateImageStatus = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	in := &pb.UpdateImageStatusRequest{}
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

	out, cErr := httpClient.grpcClient.UpdateImageStatus(context.Background(), in)
	if cErr != nil {
		log.Println("Error updating image with object_id", in.GetObjectId(), "status by object_id: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}

	js, mErr := json.Marshal(out)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

var GetSignedImageUploadUrl = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	in := &pb.SignedImageUploadUrlRequest{}
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
	signedUrlResponse, sErr := httpClient.grpcClient.GetSignedImageUploadUrl(context.Background(), in)
	if sErr != nil {
		log.Println("Error GetUser: ", sErr)
		http.Error(w, sErr.Error(), http.StatusInternalServerError)
	}
	js, mErr := json.Marshal(signedUrlResponse)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
	}
	writeJs(js, w, r)
})

var CreateImage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	in := &pb.CreateImageRequest{}
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

	in.UserId = u.GetId()
	log.Println("Creating image with ID: ", in.GetObjectId())

	out, cErr := httpClient.grpcClient.CreateImage(context.Background(), in)
	if cErr != nil {
		log.Println("Error creating image: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}
	js, mErr := json.Marshal(out)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Successfully Created Image:", out.GetObjectId())
	writeJs(js, w, r)
})

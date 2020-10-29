package http

import (
	"admin/models"
	pb "admin/protos"
	"admin/server/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"io/ioutil"
	"log"
	"net/http"
)

var GetZoomTokenByUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	in := &pb.Id{Id: u.GetId()}

	zoomToken, zoomTokenErr := httpClient.grpcClient.GetZoomTokenByUserId(context.Background(), in)
	if zoomTokenErr != nil {
		log.Println("Error GetZoomTokenByUserId: ", zoomTokenErr)
		http.Error(w, zoomTokenErr.Error(), http.StatusInternalServerError)
		return
	}
	js, mErr := json.Marshal(zoomToken)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)

})

var GetZoomAppInstallUrl = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	in := &pb.ZoomAppInstallUrlRequest{}
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

	zoomInstallUrl, cErr := httpClient.grpcClient.GetZoomAppInstallUrl(context.Background(), in)
	if cErr != nil {
		log.Println("Error GetZoomAppInstallUrl: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
	}
	js, mErr := json.Marshal(zoomInstallUrl)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

var CreateMeetingInZoom = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	z, zErr := httpClient.grpcClient.GetZoomTokenByUserId(context.Background(), &pb.Id{Id: u.GetId()})
	if zErr != nil {
		log.Println("Error getting GetZoomTokenByUserId: ", zErr)
		http.Error(w, zErr.Error(), http.StatusInternalServerError)
		return
	}

	createZoomMeetingFields := &pb.CreateZoomMeetingFields{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Unable to read body of request: ", r.Body)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = jsonpb.UnmarshalString(string(body), createZoomMeetingFields); err != nil {
		log.Println("Unable to marshal json body into protobuf message, body is: ", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	in := &pb.CreateMeetingInZoomRequest{
		Fields:      createZoomMeetingFields,
		AccessToken: z.GetAccessToken(),
	}

	meeting, meetingErr := httpClient.grpcClient.CreateMeetingInZoom(context.Background(), in)
	if meetingErr != nil {
		log.Println("Unable to create meeting: ", string(body))
		http.Error(w, meetingErr.Error(), http.StatusBadRequest)
		return
	}

	js, mErr := json.Marshal(meeting)
	if mErr != nil {
		log.Println("Error Marshalling JSON: ", mErr)
		http.Error(w, mErr.Error(), http.StatusInternalServerError)
		return
	}
	writeJs(js, w, r)
})

func saveZoomToken(w http.ResponseWriter, req *http.Request, code, grantType, grant string, userId int64) {
	tokenStringData, zErr := getZoomToken(code, grantType, grant)
	if zErr != nil {
		log.Println("Error getZoomToken: ", zErr)
		http.Error(w, zErr.Error(), http.StatusInternalServerError)
		return
	}

	z := &models.ZoomToken{}
	err := json.Unmarshal(tokenStringData, z)
	if err != nil {
		log.Println("Error unmarshalling json: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if z.AccessToken == "" {
		log.Println("access token returned from zoom is empty string")
		http.Error(w, "access token returned from zoom is empty string", http.StatusInternalServerError)
		return
	}
	// set user id for token ownership
	z.UserId = userId
	_, cErr := httpClient.grpcClient.CreateZoomAccessToken(context.Background(), z.ProtoMarshal())
	if cErr != nil {
		log.Println("Error creating access token: ", z.ProtoMarshal().String(), "error: ", cErr)
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}

	zoomUserInfoBytes, zmErr := getUserAccountInfo(z.AccessToken)
	if zmErr != nil {
		log.Println("Error retriving access token", zmErr)
		http.Error(w, zmErr.Error(), http.StatusInternalServerError)
		return
	}

	zoomUserInfo := &utils.ZoomUserInfo{}

	if unmarshalErr := json.Unmarshal(zoomUserInfoBytes, zoomUserInfo); unmarshalErr != nil {
		log.Println("Error unmarshalling user info from zoom", unmarshalErr)
		http.Error(w, unmarshalErr.Error(), http.StatusInternalServerError)
		return
	}

	// Update that we have installed the users zoom info and apply the account id
	if _, uErr := httpClient.grpcClient.UpdateUserById(context.Background(), &pb.User{Id: userId, ZoomAccountId: zoomUserInfo.AccountID, ZoomAppInstalled: true}); uErr != nil {
		log.Println("Error Updating user ZoomAccountId value to", zoomUserInfo.AccountID, "error: ", uErr)
		http.Error(w, uErr.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Success installing zoom app for user ", userId)
	return
}

func redirectZoomSuccess(w http.ResponseWriter, req *http.Request) {
	redirectUri, err := httpClient.grpcClient.GetZoomRedirectSuccessUri(context.Background(), &pb.Empty{})
	if err != nil {
		log.Println("Error getting zoom redirect uri , error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, redirectUri.GetUri(), 302)
}

func getZoomToken(authCode, grantType, grant string) ([]byte, error) {
	log.Println("Auth code is ", authCode)
	url := fmt.Sprintf("https://zoom.us/oauth/token?grant_type=%s&%s=%s", grantType, grant, authCode)

	var jsonStr = []byte(``)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(zoomClientKey, zoomClientSecret)
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
	return body, nil
}

func getUserAccountInfo(authToken string) ([]byte, error) {
	url := "https://api.zoom.us/v2/users/me"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

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
	return body, nil
}

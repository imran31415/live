package grpc

import (
	"admin/models"
	pb "admin/protos"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Gets a order by ID.
func (s *Server) GetImagesByUserId(ctx context.Context, in *pb.Id) (*pb.Images, error) {
	u, err := s.repo.GetImagesByUserId(uint(in.GetId()))
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshal(), nil
}

func (s *Server) CreateImage(ctx context.Context, in *pb.CreateImageRequest) (*pb.Image, error) {

	toCreate := &models.Image{}

	toCreate.ObjectId = in.GetObjectId()
	toCreate.Status = pb.UploadStatus_STARTED.Enum().String()
	toCreate.UserId = in.GetUserId()

	// Validate Fields
	if toCreate.UserId == 0 {
		return nil, errors.New("can not create image without user_id")
	}
	if toCreate.ObjectId == "" {
		return nil, errors.New("objectId can not be empty")
	}

	u, err := s.repo.CreateImage(toCreate)
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshal(), nil
}

func (s *Server) UpdateImageStatus(ctx context.Context, in *pb.UpdateImageStatusRequest) (*pb.Image, error) {

	toUpdate := &models.Image{}

	toUpdate.ObjectId = in.GetObjectId()
	toUpdate.Status = in.GetStatus().Enum().String()
	if toUpdate.Status == pb.UploadStatus_SUCCEEDED.Enum().String() {
		servingUrl, err := getImageServingUrl(toUpdate.ObjectId)
		if err != nil {
			return nil, err
		}
		if servingUrl == "" {
			return nil, errors.New("servingUrl returned by getImageServingUrl is empty")
		}
		toUpdate.ServingUrl = servingUrl
	}

	u, err := s.repo.UpdateImageStatus(toUpdate)
	if err != nil {
		return nil, err
	}
	return u.ProtoMarshal(), nil
}

func (s *Server) GetSignedImageUploadUrl(ctx context.Context, in *pb.SignedImageUploadUrlRequest) (*pb.SignedImageUploadUrlResponse, error) {
	getSignedURL := func(target string, values url.Values) (string, error) {
		resp, err := http.PostForm(target, values)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	var v url.Values
	switch in.GetExt() {
	case pb.SignedImageUploadUrlRequest_JPG:
		v = url.Values{"content_type": {"image/jpg"}, "ext": {"jpg"}}
	case pb.SignedImageUploadUrlRequest_JPEG:
		v = url.Values{"content_type": {"image/jpeg"}, "ext": {"jpeg"}}
	case pb.SignedImageUploadUrlRequest_PNG:
		v = url.Values{"content_type": {"image/png"}, "ext": {"png"}}
	case pb.SignedImageUploadUrlRequest_GIF:
		v = url.Values{"content_type": {"image/gif"}, "ext": {"gif"}}
	default:
		return nil, errors.New("unexpected extension")
	}
	u, err := getSignedURL(signerUrl, v)
	if err != nil {
		log.Println("error getting signed url")
		return nil, err
	}
	return &pb.SignedImageUploadUrlResponse{Url: u}, nil
}

func getImageServingUrl(s string) (string, error) {
	url := fmt.Sprintf("https://image-serve-dot-livehub-277906.uc.r.appspot.com/image-url?bucket=%s&image=%s", bucketName, s)
	method := "GET"
	client := &http.Client{}
	req, rErr := http.NewRequest(method, url, nil)

	if rErr != nil {
		return "", rErr
	}
	res, cErr := client.Do(req)
	defer res.Body.Close()
	if cErr != nil {
		return "", cErr
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		log.Println("Non 200 status code received from getting signed Url, body:", string(body))
		return "", errors.New("non 200 response in getImageServingUrl")
	}

	type Resp struct {
		ImageUrl string `json:"image_url"`
	}

	urlResp := &Resp{}

	err = json.Unmarshal(body, urlResp)
	if err != nil {
		return "", err
	}
	if urlResp.ImageUrl == "" {
		return "", errors.New("serving image_url can not be empty ")
	}
	return urlResp.ImageUrl, nil
}

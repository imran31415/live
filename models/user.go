package models

import (
	pb "admin/protos"
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	FirstName         string
	LastName          string
	Email             string
	Phone             string
	AuthZeroSubId     string
	ProfileImgUrl     string
	ZoomAccessToken   string
	ZoomRefreshToken  string
	Name              string
	Description       string
	InstagramUrl      string
	FacebookUrl       string
	YoutubeChannelUrl string
	TwitterUrl        string
	BannerImgUrl      string
	Tags              string
	ZoomAppInstalled  bool
	ZoomAccountId     string
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) ProtoUnMarshal(user *pb.User) {
	u.Model.ID = uint(user.GetId())
	u.FirstName = user.GetFirstName()
	u.LastName = user.GetLastName()
	u.Email = user.GetEmail()
	u.Phone = user.GetPhone()
	u.AuthZeroSubId = user.GetAuthZeroSubId()
	u.ProfileImgUrl = user.GetProfileImgUrl()
	u.ZoomAccessToken = user.GetZoomAccessToken()
	u.ZoomRefreshToken = user.GetZoomRefreshToken()
	u.Name = user.GetName()
	u.Description = user.GetDescription()
	u.InstagramUrl = user.GetInstagramUrl()
	u.FacebookUrl = user.GetFacebookUrl()
	u.YoutubeChannelUrl = user.GetYoutubeChannelUrl()
	u.TwitterUrl = user.GetTwitterUrl()
	u.BannerImgUrl = user.GetBannerImgUrl()
	u.Tags = user.GetTags()
	u.ZoomAppInstalled = user.GetZoomAppInstalled()
	u.ZoomAccountId = user.GetZoomAccountId()
}

// ProtoMarshalPrivate gets the protobuf representation of the DB
func (u *User) ProtoMarshalPrivate() *pb.User {
	return &pb.User{
		Id:                int64(u.Model.ID),
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		Email:             u.Email,
		Phone:             u.Phone,
		AuthZeroSubId:     u.AuthZeroSubId,
		ProfileImgUrl:     u.ProfileImgUrl,
		ZoomAccessToken:   u.ZoomAccessToken,
		ZoomRefreshToken:  u.ZoomRefreshToken,
		Description:       u.Description,
		Name:              u.Name,
		InstagramUrl:      u.InstagramUrl,
		FacebookUrl:       u.FacebookUrl,
		YoutubeChannelUrl: u.YoutubeChannelUrl,
		TwitterUrl:        u.TwitterUrl,
		BannerImgUrl:      u.BannerImgUrl,
		Tags:              u.Tags,
		ZoomAppInstalled:  u.ZoomAppInstalled,
		ZoomAccountId:     u.ZoomAccountId,
	}
}

// ProtoMarshalPublic gets the public protobuf representation of the DB.
// This method strips some fields that we do not want tos how in the context in which this method is called
func (u *User) ProtoMarshalPublic() *pb.User {
	return &pb.User{
		Id:        int64(u.Model.ID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		// Email:             u.Email,
		//Phone:             u.Phone,
		//AuthZeroSubId:     u.AuthZeroSubId,
		ProfileImgUrl: u.ProfileImgUrl,
		//ZoomAccessToken:   u.ZoomAccessToken,
		//ZoomRefreshToken:  u.ZoomRefreshToken,
		Description:       u.Description,
		Name:              u.Name,
		InstagramUrl:      u.InstagramUrl,
		FacebookUrl:       u.FacebookUrl,
		YoutubeChannelUrl: u.YoutubeChannelUrl,
		TwitterUrl:        u.TwitterUrl,
		BannerImgUrl:      u.BannerImgUrl,
		Tags:              u.Tags,
		// ZoomAppInstalled:  u.ZoomAppInstalled,
	}
}

func (u *User) ProtoMarshalPublicWithEmailAndPhone() *pb.User {
	return &pb.User{
		Id:        int64(u.Model.ID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Phone:     u.Phone,
		//AuthZeroSubId:     u.AuthZeroSubId,
		ProfileImgUrl: u.ProfileImgUrl,
		//ZoomAccessToken:   u.ZoomAccessToken,
		//ZoomRefreshToken:  u.ZoomRefreshToken,
		Description:       u.Description,
		Name:              u.Name,
		InstagramUrl:      u.InstagramUrl,
		FacebookUrl:       u.FacebookUrl,
		YoutubeChannelUrl: u.YoutubeChannelUrl,
		TwitterUrl:        u.TwitterUrl,
		BannerImgUrl:      u.BannerImgUrl,
		Tags:              u.Tags,
		// ZoomAppInstalled:  u.ZoomAppInstalled,
	}
}

type Users struct {
	Users []*User
}

// ProtoMarshalPrivate gets the protobuf representation of the DB
func (c *Users) ProtoMarshal() *pb.Users {
	var users []*pb.User
	for _, i := range c.Users {
		users = append(users, i.ProtoMarshalPublic())
	}
	return &pb.Users{Users: users}
}

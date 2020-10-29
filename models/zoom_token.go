package models

import (
	pb "admin/protos"
	"github.com/jinzhu/gorm"
)

type ZoomToken struct {
	gorm.Model
	AccessToken  string `json:"access_token" gorm:"size:1500" `
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token" gorm:"size:1500"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	UserId       int64  `json:"user_id,omitempty"`
}

func (c *ZoomToken) TableName() string {
	return "zoom_tokens"
}

func (c *ZoomToken) ProtoUnMarshal(i *pb.ZoomToken) {
	c.Model.ID = uint(i.GetId())
	c.AccessToken = i.GetAccessToken()
	c.TokenType = i.GetTokenType()
	c.RefreshToken = i.GetRefreshToken()
	c.ExpiresIn = i.GetExpiresIn()
	c.Scope = i.GetScope()
	c.UserId = i.GetUserId()

}

// ProtoMarshalPrivate gets the protobuf representation of the DB
func (c *ZoomToken) ProtoMarshal() *pb.ZoomToken {
	return &pb.ZoomToken{
		Id:           int64(c.Model.ID),
		AccessToken:  c.AccessToken,
		TokenType:    c.TokenType,
		RefreshToken: c.RefreshToken,
		ExpiresIn:    c.ExpiresIn,
		Scope:        c.Scope,
		UserId:       c.UserId,
	}
}

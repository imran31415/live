package models

import (
	pb "admin/protos"
	"github.com/jinzhu/gorm"
)

type Image struct {
	gorm.Model
	UserId     int64
	Status     string
	ServingUrl string
	ObjectId   string
}

func (c *Image) TableName() string {
	return "images"
}

func (c *Image) ProtoUnMarshal(i *pb.Image) {
	c.Model.ID = uint(i.GetId())
	c.UserId = i.GetUserId()
	c.Status = i.GetStatus().Enum().String()
	c.ServingUrl = i.GetServingUrl()
	c.ObjectId = i.GetObjectId()

}

// ProtoMarshalPrivate gets the protobuf representation of the DB
func (c *Image) ProtoMarshal() *pb.Image {

	return &pb.Image{
		Id:         int64(c.Model.ID),
		UserId:     c.UserId,
		Status:     pb.UploadStatus(pb.UploadStatus_value[c.Status]),
		ServingUrl: c.ServingUrl,
		ObjectId:   c.ObjectId,
	}
}

type Images struct {
	Images []*Image
}

// ProtoMarshalPrivate gets the protobuf representation of the DB
func (c *Images) ProtoMarshal() *pb.Images {
	var images []*pb.Image
	for _, i := range c.Images {
		images = append(images, i.ProtoMarshal())
	}
	return &pb.Images{Images: images}
}

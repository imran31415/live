package models

import (
	pb "admin/protos"

	"github.com/jinzhu/gorm"
)

type Order struct {
	gorm.Model
	CustomerId                int64
	PaymentPlatform           string
	Status                    string
	TransactionType           string
	PaymentMethodId           string
	PaymentPlatformCustomerId string
	PaymentPlatformOrderId    string
	Amount                    int64
	Email                     string
	Tags                      string
	SessionId                 int64
	SessionName               string
	SessionDescription        string `gorm:"size:5000"`
	SessionDate               int64
}

func (c *Order) TableName() string {
	return "orders"
}

func (c *Order) ProtoUnMarshal(o *pb.Order) {
	c.Model.ID = uint(o.GetId())
	c.CustomerId = o.GetCustomerId()
	c.PaymentPlatform = o.GetPaymentPlatform().Enum().String()
	c.PaymentPlatformCustomerId = o.GetPaymentPlatformCustomerId()
	c.PaymentPlatformOrderId = o.GetPaymentPlatformOrderId()
	c.PaymentMethodId = o.GetPaymentMethodId()
	c.Status = o.GetStatus().Enum().String()
	c.Amount = o.GetAmount()
	c.Email = o.GetEmail()
	c.SessionId = o.GetSessionId()
	c.Tags = o.GetTags()
	c.TransactionType = o.GetTransactionType().Enum().String()
	c.SessionName = o.GetSessionName()
	c.SessionDescription = o.GetSessionDescription()
	c.SessionDate = o.GetSessionDate()

}

// ProtoMarshalPrivate gets the protobuf representation of the DB
func (o *Order) ProtoMarshal() *pb.Order {
	return &pb.Order{
		Id:                        int64(o.Model.ID),
		CustomerId:                o.CustomerId,
		PaymentPlatform:           pb.PaymentPlatforms(pb.PaymentPlatforms_value[o.PaymentPlatform]),
		PaymentPlatformCustomerId: o.PaymentPlatformCustomerId,
		PaymentPlatformOrderId:    o.PaymentPlatformOrderId,
		Status:                    pb.Order_PaymentPlatformStatus(pb.Order_PaymentPlatformStatus_value[o.Status]),
		PaymentMethodId:           o.PaymentMethodId,
		Amount:                    o.Amount,
		Email:                     o.Email,
		SessionId:                 o.SessionId,
		Tags:                      o.Tags,
		TransactionType:           pb.TransactionType(pb.TransactionType_value[o.TransactionType]),
		SessionName:               o.SessionName,
		SessionDescription:        o.SessionDescription,
		SessionDate:               o.SessionDate,
	}
}

type Orders struct {
	Orders []*Order
}

// ProtoMarshalPrivate gets the protobuf representation of the DB
func (c *Orders) ProtoMarshal() *pb.Orders {
	var s []*pb.Order
	for _, i := range c.Orders {
		s = append(s, i.ProtoMarshal())
	}
	return &pb.Orders{Orders: s}
}

package models

import (
	pb "admin/protos"
	"github.com/jinzhu/gorm"
)

type Customer struct {
	gorm.Model
	UserId                    int64
	PaymentPlatform           string
	PaymentPlatformCustomerId string
	Email                     string
}

func (c *Customer) TableName() string {
	return "customers"
}

func (c *Customer) ProtoUnMarshal(customer *pb.Customer) {
	c.Model.ID = uint(customer.GetId())
	c.UserId = customer.GetUserId()
	c.PaymentPlatform = customer.GetPlatform().Enum().String()
	c.PaymentPlatformCustomerId = customer.GetPaymentPlatformCustomerId()
	c.Email = customer.GetEmail()
}

// ProtoMarshalPrivate gets the protobuf representation of the DB
func (c *Customer) ProtoMarshal() *pb.Customer {

	return &pb.Customer{
		Id:                        int64(c.Model.ID),
		UserId:                    c.UserId,
		Platform:                  pb.PaymentPlatforms(pb.PaymentPlatforms_value[c.PaymentPlatform]),
		PaymentPlatformCustomerId: c.PaymentPlatformCustomerId,
		Email:                     c.Email,
	}
}

type Customers struct {
	Customers []*Customer
}

// ProtoMarshalPrivate gets the protobuf representation of the DB
func (c *Customers) ProtoMarshal() *pb.Customers {
	var customers []*pb.Customer
	for _, i := range c.Customers {
		customers = append(customers, i.ProtoMarshal())
	}
	return &pb.Customers{Customers: customers}
}

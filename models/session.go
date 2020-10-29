package models

import (
	pb "admin/protos"
	"github.com/jinzhu/gorm"
	"google.golang.org/protobuf/proto"
)

type Session struct {
	gorm.Model
	UserId      int64
	Name        string
	Description string `gorm:"size:5000"`
	Date        int64
	MaxSize     int32
	//CostCents unit is cents, i.e $123.56 = 123456
	CostCents     int64
	Duration      int64
	MeetingUrl    *string
	ProfileImgUrl string
	BannerImgUrl  string
	Tags          string
	ZoomPassword  *string
	// Bool must be a pointer so we can update from true to false,
	// since false is interpreted by gorm as unset so it does not apply the update
	IsDraft             *bool
	PaywallType         string
	TimeZone            string
	ZoomMeetingId       *int64
	ZoomMeetingJoinUrl  *string `gorm:"size:1500"`
	ZoomMeetingStartUrl *string `gorm:"size:1500"`
	Source              string
	ZoomSyncEnabled     *bool
	ZoomOccurrenceId    string
	ZoomMeetingType     int64
}

func (c *Session) TableName() string {
	return "sessions"
}

func (c *Session) ProtoUnMarshal(s *pb.Session) {
	c.Model.ID = uint(s.GetId())
	c.UserId = s.GetUserId()
	c.Name = s.GetName()
	c.Date = s.GetStartTime()
	c.Description = s.GetDescription()
	c.Description = s.GetDescription()
	c.MaxSize = s.GetMaxSessionSize()
	c.CostCents = s.GetCost()
	c.Duration = s.GetDuration()

	c.ProfileImgUrl = s.GetProfileImgUrl()
	c.Tags = s.GetTags()

	c.PaywallType = s.GetPaywallType().Enum().String()
	c.TimeZone = s.GetTimeZone().Enum().String()

	c.Source = s.GetSource().Enum().String()
	c.ZoomOccurrenceId = s.GetZoomOccurrenceId()
	c.ZoomMeetingType = s.GetZoomMeetingType()

	if s.GetZoomMeetingId() != nil {
		c.ZoomMeetingId = proto.Int64(s.GetZoomMeetingId().GetValue())
	}
	if s.GetMeetingUrl() != nil {
		c.MeetingUrl = proto.String(s.GetMeetingUrl().GetValue())
	}
	if s.GetZoomPassword() != nil {
		c.ZoomPassword = proto.String(s.GetZoomPassword().GetValue())
	}
	if s.GetZoomSyncEnabled() != nil {
		c.ZoomSyncEnabled = proto.Bool(s.GetZoomSyncEnabled().GetValue())
	}
	if s.GetZoomMeetingJoinUrl() != nil {
		c.ZoomMeetingJoinUrl = proto.String(s.GetZoomMeetingJoinUrl().GetValue())
	}
	if s.GetZoomMeetingStartUrl() != nil {
		c.ZoomMeetingStartUrl = proto.String(s.GetZoomMeetingStartUrl().GetValue())
	}
	if s.GetIsDraft() != nil {
		c.IsDraft = proto.Bool(s.GetIsDraft().GetValue())
	}
}
func (c *Session) ProtoMarshalPrivate() *pb.Session {
	return &pb.Session{
		Id:             int64(c.Model.ID),
		UserId:         c.UserId,
		Name:           c.Name,
		Description:    c.Description,
		StartTime:      c.Date,
		MaxSessionSize: c.MaxSize,
		Cost:           c.CostCents,
		Duration:       c.Duration,
		MeetingUrl:     &pb.OptionalString{Value: stringFromStringPointer(c.MeetingUrl)},
		ProfileImgUrl:  c.ProfileImgUrl,
		BannerImgUrl:   c.BannerImgUrl,
		Tags:           c.Tags,
		ZoomPassword:   &pb.OptionalString{Value: stringFromStringPointer(c.ZoomPassword)},
		IsDraft:        &pb.OptionalBool{Value: boolFromBoolPointer(c.IsDraft)},
		PaywallType:    pb.Session_PaywallType(pb.Session_PaywallType_value[c.PaywallType]),
		TimeZone:       pb.TimeZone(pb.TimeZone_value[c.TimeZone]),
		// TODO (we dont want to expose these to users other than the user requesting: create another endpoint
		ZoomMeetingId:       &pb.OptionalInt64{Value: int64FromInt64Pointer(c.ZoomMeetingId)},
		ZoomMeetingJoinUrl:  &pb.OptionalString{Value: stringFromStringPointer(c.ZoomMeetingJoinUrl)},
		ZoomMeetingStartUrl: &pb.OptionalString{Value: stringFromStringPointer(c.ZoomMeetingStartUrl)},
		Source:              pb.Session_SourceType(pb.Session_SourceType_value[c.Source]),
		ZoomSyncEnabled:     &pb.OptionalBool{Value: boolFromBoolPointer(c.ZoomSyncEnabled)},
		ZoomOccurrenceId:    c.ZoomOccurrenceId,
		ZoomMeetingType:     c.ZoomMeetingType,
	}
}

func (c *Session) ProtoMarshalPublic() *pb.Session {
	return &pb.Session{
		Id:             int64(c.Model.ID),
		UserId:         c.UserId,
		Name:           c.Name,
		Description:    c.Description,
		StartTime:      c.Date,
		MaxSessionSize: c.MaxSize,
		Cost:           c.CostCents,
		Duration:       c.Duration,
		MeetingUrl:     &pb.OptionalString{Value: stringFromStringPointer(c.MeetingUrl)},
		ProfileImgUrl:  c.ProfileImgUrl,
		BannerImgUrl:   c.BannerImgUrl,
		Tags:           c.Tags,
		// ZoomPassword:   &pb.OptionalString{Value: stringFromStringPointer(c.ZoomPassword)},
		IsDraft:     &pb.OptionalBool{Value: boolFromBoolPointer(c.IsDraft)},
		PaywallType: pb.Session_PaywallType(pb.Session_PaywallType_value[c.PaywallType]),
		TimeZone:    pb.TimeZone(pb.TimeZone_value[c.TimeZone]),
		// TODO (we dont want to expose these to users other than the user requesting: create another endpoint
		// ZoomMeetingId:         &pb.OptionalInt64{Value: int64FromInt64Pointer(c.ZoomMeetingId)},
		ZoomMeetingJoinUrl: &pb.OptionalString{Value: stringFromStringPointer(c.ZoomMeetingJoinUrl)},
		// ZoomMeetingStartUrl:   &pb.OptionalString{Value: stringFromStringPointer(c.ZoomMeetingStartUrl)},
		// Source:                pb.Session_SourceType(pb.Session_SourceType_value[c.Source]),
		// ZoomSyncEnabled:       &pb.OptionalBool{Value: boolFromBoolPointer(c.ZoomSyncEnabled)},
		// ZoomOccurrenceId:      c.ZoomOccurrenceId,
		// ZoomMeetingType:       c.ZoomMeetingType,
		// ZoomRawMeetingPayload: c.ZoomRawMeetingPayload,
	}
}

type Sessions struct {
	Sessions []*Session
}

// ProtoMarshalPrivate gets the protobuf representation of the DB
func (c *Sessions) ProtoMarshalPrivate() *pb.Sessions {
	var s []*pb.Session
	for _, i := range c.Sessions {
		s = append(s, i.ProtoMarshalPrivate())
	}
	return &pb.Sessions{Sessions: s}
}

// ProtoMarshalPrivate gets the protobuf representation of the DB
func (c *Sessions) ProtoMarshalPublic() *pb.Sessions {
	var s []*pb.Session
	for _, i := range c.Sessions {
		s = append(s, i.ProtoMarshalPublic())
	}
	return &pb.Sessions{Sessions: s}
}

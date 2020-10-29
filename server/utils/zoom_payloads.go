package utils

import (
	pb "admin/protos"
	"time"
)

const (
	// These are defaulted to empty as our grpc.CreateSession logic should fill these if empty with unsplash
	defaultBannerImage     = ""
	defaultProfileUrl      = ""
	defaultMaxParticipants = 100

	ZoomMeetingTypeSingular   = 2
	ZoomMeetingTypeReocurring = 8
)

type MeetingInfo struct {
	Agenda      string    `json:"agenda"`
	CreatedAt   time.Time `json:"created_at"`
	Duration    int       `json:"duration"`
	HostID      string    `json:"host_id"`
	ID          int       `json:"id"`
	JoinURL     string    `json:"join_url"`
	Occurrences []struct {
		OccurrenceID string    `json:"occurrence_id"`
		StartTime    time.Time `json:"start_time"`
		Duration     int       `json:"duration"`
		Status       string    `json:"status"`
	} `json:"occurrences"`
	Settings struct {
		AlternativeHosts      string   `json:"alternative_hosts"`
		ApprovalType          int      `json:"approval_type"`
		Audio                 string   `json:"audio"`
		AutoRecording         string   `json:"auto_recording"`
		CloseRegistration     bool     `json:"close_registration"`
		CnMeeting             bool     `json:"cn_meeting"`
		EnforceLogin          bool     `json:"enforce_login"`
		EnforceLoginDomains   string   `json:"enforce_login_domains"`
		GlobalDialInCountries []string `json:"global_dial_in_countries"`
		GlobalDialInNumbers   []struct {
			City        string `json:"city"`
			Country     string `json:"country"`
			CountryName string `json:"country_name"`
			Number      string `json:"number"`
			Type        string `json:"type"`
		} `json:"global_dial_in_numbers"`
		HostVideo                    bool `json:"host_video"`
		InMeeting                    bool `json:"in_meeting"`
		JoinBeforeHost               bool `json:"join_before_host"`
		MuteUponEntry                bool `json:"mute_upon_entry"`
		ParticipantVideo             bool `json:"participant_video"`
		RegistrantsConfirmationEmail bool `json:"registrants_confirmation_email"`
		UsePmi                       bool `json:"use_pmi"`
		WaitingRoom                  bool `json:"waiting_room"`
		Watermark                    bool `json:"watermark"`
		RegistrantsEmailNotification bool `json:"registrants_email_notification"`
	} `json:"settings"`
	StartTime time.Time `json:"start_time"`
	StartURL  string    `json:"start_url"`
	Status    string    `json:"status"`
	Timezone  string    `json:"timezone"`
	Topic     string    `json:"topic"`
	Type      int       `json:"type"`
	UUID      string    `json:"uuid"`
}

type WebHookEvent struct {
	Event   string `json:"event"`
	Payload struct {
		AccountID  string `json:"account_id"`
		Operator   string `json:"operator"`
		OperatorID string `json:"operator_id"`
		Object     struct {
			Occurrences []struct {
				OccurrenceID string    `json:"occurrence_id"`
				StartTime    time.Time `json:"start_time"`
				Duration     int       `json:"duration"`
			} `json:"occurrences"`
			UUID   string `json:"uuid"`
			ID     int64  `json:"id"`
			HostID string `json:"host_id"`
			Topic  string `json:"topic"`
			Type   int    `json:"type"`
			// this could be time.Time, but sometimes its nil and we dont really use it since we pull the meetinginfo from zoom
			StartTime string `json:"start_time"`
			Duration  int    `json:"duration"`
			Timezone  string `json:"timezone"`
			Password  string `json:"password"`
		} `json:"object"`
	} `json:"payload"`
}

type ZoomUserInfo struct {
	ID                 string        `json:"id"`
	FirstName          string        `json:"first_name"`
	LastName           string        `json:"last_name"`
	Email              string        `json:"email"`
	Type               int           `json:"type"`
	RoleName           string        `json:"role_name"`
	Pmi                int           `json:"pmi"`
	UsePmi             bool          `json:"use_pmi"`
	VanityURL          string        `json:"vanity_url"`
	PersonalMeetingURL string        `json:"personal_meeting_url"`
	Timezone           string        `json:"timezone"`
	Verified           int           `json:"verified"`
	Dept               string        `json:"dept"`
	CreatedAt          time.Time     `json:"created_at"`
	LastLoginTime      time.Time     `json:"last_login_time"`
	LastClientVersion  string        `json:"last_client_version"`
	PicURL             string        `json:"pic_url"`
	HostKey            string        `json:"host_key"`
	Jid                string        `json:"jid"`
	GroupIds           []interface{} `json:"group_ids"`
	ImGroupIds         []string      `json:"im_group_ids"`
	AccountID          string        `json:"account_id"`
	Language           string        `json:"language"`
	PhoneCountry       string        `json:"phone_country"`
	PhoneNumber        string        `json:"phone_number"`
	Status             string        `json:"status"`
}

//getTopicAndAgenda returns agenda if its not nil or defaults to the topic
func getNameAndDescription(info *MeetingInfo) (n string, d string) {
	n = info.Topic
	d = info.Topic
	if info.Agenda != "" {
		d = info.Agenda
	}
	return n, d

}

func SessionFromZoomMeetingWebHookCreated(password string, meetingInfo *MeetingInfo, userId int64) *pb.Session {

	n, d := getNameAndDescription(meetingInfo)
	return &pb.Session{
		UserId:         userId,
		Name:           n,
		Description:    d,
		StartTime:      meetingInfo.StartTime.Unix(),
		MeetingUrl:     &pb.OptionalString{Value: meetingInfo.JoinURL},
		MaxSessionSize: defaultMaxParticipants,
		BannerImgUrl:   defaultBannerImage,
		// TODO: do we want all meetings to automatically be free?
		// Probably better to let the user decide if their default is free or a certain amount
		// @nelson
		Cost:          500,
		Duration:      int64(meetingInfo.Duration),
		ProfileImgUrl: defaultProfileUrl,
		Tags:          "",
		ZoomPassword:  &pb.OptionalString{Value: password},
		// TODO review
		// @nelson we probably want whether sessions are created in draft or not to also be a user setting
		IsDraft:             &pb.OptionalBool{Value: true},
		ZoomMeetingId:       &pb.OptionalInt64{Value: int64(meetingInfo.ID)},
		ZoomMeetingJoinUrl:  &pb.OptionalString{Value: meetingInfo.JoinURL},
		ZoomMeetingStartUrl: &pb.OptionalString{Value: meetingInfo.StartURL},
		Source:              pb.Session_ZOOM_WEBHOOK,
		CreateMeetingInZoom: false,
		ZoomSyncEnabled:     &pb.OptionalBool{Value: true},
		ZoomMeetingType:     int64(meetingInfo.Type),
	}
}

func SessionFromZoomMeetingWebHookUpdate(meetingInfo *MeetingInfo, duration, startTime int64) *pb.Session {
	n, d := getNameAndDescription(meetingInfo)
	return &pb.Session{
		Name:                n,
		Description:         d,
		StartTime:           startTime,
		MeetingUrl:          &pb.OptionalString{Value: meetingInfo.JoinURL},
		Duration:            duration,
		ZoomMeetingId:       &pb.OptionalInt64{Value: int64(meetingInfo.ID)},
		ZoomMeetingJoinUrl:  &pb.OptionalString{Value: meetingInfo.JoinURL},
		ZoomMeetingStartUrl: &pb.OptionalString{Value: meetingInfo.StartURL},
	}
}

func SessionUpdateFieldsForComparison(s *pb.Session) *pb.Session {
	return &pb.Session{
		Id:                  s.GetId(),
		Name:                s.GetName(),
		Description:         s.GetDescription(),
		StartTime:           s.GetStartTime(),
		MeetingUrl:          s.GetMeetingUrl(),
		Duration:            s.GetDuration(),
		ZoomMeetingId:       s.GetZoomMeetingId(),
		ZoomMeetingJoinUrl:  s.GetZoomMeetingJoinUrl(),
		ZoomMeetingStartUrl: s.GetZoomMeetingStartUrl(),
	}
}

type CreateZoomMeetingApiResponse struct {
	UUID              string    `json:"uuid"`
	ID                int64     `json:"id"`
	HostID            string    `json:"host_id"`
	Topic             string    `json:"topic"`
	Type              int       `json:"type"`
	Status            string    `json:"status"`
	StartTime         time.Time `json:"start_time"`
	Duration          int       `json:"duration"`
	Timezone          string    `json:"timezone"`
	Agenda            string    `json:"agenda"`
	CreatedAt         time.Time `json:"created_at"`
	StartURL          string    `json:"start_url"`
	JoinURL           string    `json:"join_url"`
	Password          string    `json:"password"`
	H323Password      string    `json:"h323_password"`
	PstnPassword      string    `json:"pstn_password"`
	EncryptedPassword string    `json:"encrypted_password"`
	Settings          struct {
		HostVideo                    bool     `json:"host_video"`
		ParticipantVideo             bool     `json:"participant_video"`
		CnMeeting                    bool     `json:"cn_meeting"`
		InMeeting                    bool     `json:"in_meeting"`
		JoinBeforeHost               bool     `json:"join_before_host"`
		MuteUponEntry                bool     `json:"mute_upon_entry"`
		Watermark                    bool     `json:"watermark"`
		UsePmi                       bool     `json:"use_pmi"`
		ApprovalType                 int      `json:"approval_type"`
		Audio                        string   `json:"audio"`
		AutoRecording                string   `json:"auto_recording"`
		EnforceLogin                 bool     `json:"enforce_login"`
		EnforceLoginDomains          string   `json:"enforce_login_domains"`
		AlternativeHosts             string   `json:"alternative_hosts"`
		CloseRegistration            bool     `json:"close_registration"`
		RegistrantsConfirmationEmail bool     `json:"registrants_confirmation_email"`
		WaitingRoom                  bool     `json:"waiting_room"`
		GlobalDialInCountries        []string `json:"global_dial_in_countries"`
		GlobalDialInNumbers          []struct {
			CountryName string `json:"country_name"`
			City        string `json:"city"`
			Number      string `json:"number"`
			Type        string `json:"type"`
			Country     string `json:"country"`
		} `json:"global_dial_in_numbers"`
		RegistrantsEmailNotification bool `json:"registrants_email_notification"`
		MeetingAuthentication        bool `json:"meeting_authentication"`
	} `json:"settings"`
}

type ZoomAccountDeauthorizedPayload struct {
	Event   string `json:"event"`
	Payload struct {
		UserDataRetention   string    `json:"user_data_retention"`
		AccountID           string    `json:"account_id"`
		UserID              string    `json:"user_id"`
		Signature           string    `json:"signature"`
		DeauthorizationTime time.Time `json:"deauthorization_time"`
		ClientID            string    `json:"client_id"`
	} `json:"payload"`
}

type ZoomMeetingsPaginatedResponse struct {
	PageCount    int `json:"page_count"`
	PageNumber   int `json:"page_number"`
	PageSize     int `json:"page_size"`
	TotalRecords int `json:"total_records"`
	Meetings     []struct {
		UUID      string    `json:"uuid"`
		ID        int       `json:"id"`
		HostID    string    `json:"host_id"`
		Topic     string    `json:"topic"`
		Type      int       `json:"type"`
		StartTime time.Time `json:"start_time"`
		Duration  int       `json:"duration"`
		Timezone  string    `json:"timezone"`
		CreatedAt time.Time `json:"created_at"`
		JoinURL   string    `json:"join_url"`
		Agenda    string    `json:"agenda,omitempty"`
	} `json:"meetings"`
}

// UpdateMeetingInZoomPayloadFromSession is used when we update a class in our system to generate the fields we want to also update in zoom
func UpdateMeetingInZoomPayloadFromSession(s *pb.Session) *pb.UpdateZoomMeetingFields {
	return &pb.UpdateZoomMeetingFields{
		Topic: s.GetName(),
		// Based on zoom time documentation
		StartTime: time.Unix(s.GetStartTime(), 0).Format(time.RFC3339),
		Duration:  s.GetDuration(),
	}
}

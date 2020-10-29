package repo

import (
	"admin/models"
	pb "admin/protos"
	"github.com/golang/protobuf/proto"
	"golang.org/x/sync/errgroup"
	"strings"
	"time"
)

const (
	previousSessionLimit = 500
	upcomingSessionLimit = 500
	ordersLimit          = 500
)

func (m *SqlRepo) IsSessionIsPurchasedByUser(userId, SessionId uint) (bool, error) {

	c := &models.Customer{}

	if err := m.DB.Where("user_id = ?", userId).First(&c).Error; err != nil {
		// No customer exists so no order can exist
		return false, err
	}

	o := &models.Order{}
	if err := m.DB.Where("customer_id = ? AND status = ? AND session_id = ?", c.ID, pb.Order_SUCCEEDED.Enum().String(), SessionId).First(&o).Error; err != nil {
		return false, err
	}
	if o.ID != 0 {
		return true, nil
	}
	return false, nil
}

func (m *SqlRepo) UpdateSession(c *models.Session) (*models.Session, error) {
	if err := m.DB.Where("id = ?", c.ID).Find(&models.Session{}).Updates(c).Error; err != nil {
		return nil, err
	}
	out, err := m.GetSessionById(c.ID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *SqlRepo) UpdateSessionsNoZoomSyncByUserId(userId int64) error {
	if err := m.DB.Where("user_id = ? AND zoom_sync_enabled = ?", userId, true).Find(&models.Session{}).Updates(&models.Session{ZoomSyncEnabled: proto.Bool(false)}).Error; err != nil {
		return err
	}
	return nil
}

func (m *SqlRepo) CreateSession(c *models.Session) (*models.Session, error) {
	if err := m.DB.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (m *SqlRepo) GetSessionById(id uint) (*models.Session, error) {
	c := &models.Session{}
	if err := m.DB.Where("id = ?", id).First(&c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (m *SqlRepo) GetSessionByUserIdAndZoomMeetingId(userId, zoomMeetingId int64) (*models.Session, error) {
	c := &models.Session{}
	if err := m.DB.Where("user_id = ? AND zoom_meeting_id = ?", userId, zoomMeetingId).First(&c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (m *SqlRepo) GetSessionByUserIdAndZoomMeetingIdAndRawPayloadId(userId, zoomMeetingId int64, zoomRawPayload string) (*models.Session, error) {
	c := &models.Session{}
	if err := m.DB.Where("user_id = ? AND zoom_meeting_id = ? AND zoom_raw_meeting_payload = ?", userId, zoomMeetingId, zoomRawPayload).First(&c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (m *SqlRepo) GetSessionByUserIdAndZoomMeetingIdAndOccurrenceId(userId, zoomMeetingId int64, occurrenceId string) (*models.Session, error) {
	c := &models.Session{}
	if err := m.DB.Where("user_id = ? AND zoom_meeting_id = ? and zoom_occurrence_id = ?", userId, zoomMeetingId, occurrenceId).First(&c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (m *SqlRepo) GetSessionByUserIdAndNameAndDateAndDuration(userId, date, duration int64, name string) (*models.Session, error) {
	c := &models.Session{}
	if err := m.DB.Where("user_id = ? AND date = ? AND name = ? AND duration = ?", userId, date, name, duration).First(&c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (m *SqlRepo) GetSessionByUserIdAndMeetingId(userId, meetingId int64) (*models.Session, error) {
	c := &models.Session{}
	if err := m.DB.Where("user_id = ? AND zoom_meeting_id = ?", userId, meetingId).First(&c).Error; err != nil {
		return nil, err
	}
	return c, nil
}
func (m *SqlRepo) DeleteSessionById(id uint) error {
	c := &models.Session{}
	if err := m.DB.Where("id = ?", id).Delete(&c).Error; err != nil {
		return err
	}
	return nil
}

func (m *SqlRepo) DeleteSessionsByUserIdAndZoomMeetingId(userId, zoomMeetingId uint) error {
	c := &models.Session{}
	if err := m.DB.Where("user_id = ? AND zoom_meeting_id = ?", userId, zoomMeetingId).Delete(&c).Error; err != nil {
		return err
	}
	return nil
}

func (m *SqlRepo) DeleteSessionsByUserIdAndZoomMeetingIdAndNotOccurrenceIds(userId, zoomMeetingId uint, notOccurrenceIds []string) error {
	c := &models.Session{}
	if err := m.DB.Where("user_id = ? AND zoom_meeting_id = ? AND zoom_occurrence_id NOT IN (?)", userId, zoomMeetingId, notOccurrenceIds).Delete(&c).Error; err != nil {
		return err
	}
	return nil
}

func (m *SqlRepo) GetSessionsByStartDate(date int64, limit int) (models.Sessions, error) {
	var out []*models.Session
	if err := m.DB.Where("date >= ?", date).Not("is_draft", true).Order("date").Limit(limit).Find(&out).Error; err != nil {
		return models.Sessions{Sessions: nil}, err
	}
	return models.Sessions{Sessions: out}, nil
}

func (m *SqlRepo) GetSessionsByStartDateAndTag(date int64, limit int, tag string) (models.Sessions, error) {
	var out []*models.Session
	if err := m.DB.Where("date >= ? AND (tags LIKE ?", date, fmtLikeStatement(tag)).Not("is_draft", true).Order("date").Limit(limit).Find(&out).Error; err != nil {
		return models.Sessions{Sessions: nil}, err
	}
	return models.Sessions{Sessions: out}, nil
}

func (m *SqlRepo) GetUpcomingSessionsByUserIdAndDate(id uint, date int64) (models.Sessions, error) {
	if date < 0 {
		date = time.Now().Unix()
	}
	var out []*models.Session
	if err := m.DB.Where("user_id = ? AND date >= ?", id, date).Order("date").Limit(upcomingSessionLimit).Find(&out).Error; err != nil {
		return models.Sessions{}, err
	}
	return models.Sessions{Sessions: out}, nil
}

func (m *SqlRepo) GetPreviousSessionsByUserIdAndDate(id uint, date int64) (models.Sessions, error) {
	if date < 0 {
		date = time.Now().Unix()
	}
	var out []*models.Session
	if err := m.DB.Where("user_id = ? AND date < ?", id, date).Order("date desc").Limit(previousSessionLimit).Find(&out).Error; err != nil {
		return models.Sessions{}, err
	}
	return models.Sessions{Sessions: out}, nil
}

func (m *SqlRepo) GetUpcomingSessionsByIds(ids []uint) (models.Sessions, error) {
	var out []*models.Session
	if err := m.DB.Where("id in (?) AND date >= ?", ids, time.Now().Unix()).Order("date").Limit(upcomingSessionLimit).Find(&out).Error; err != nil {
		return models.Sessions{}, err
	}
	return models.Sessions{Sessions: out}, nil
}

func (m *SqlRepo) GetPreviousSessionsByIds(ids []uint) (models.Sessions, error) {
	var out []*models.Session
	if err := m.DB.Where("id in (?) AND date < ?", ids, time.Now().Unix()).Order("date").Limit(upcomingSessionLimit).Find(&out).Error; err != nil {
		return models.Sessions{}, err
	}
	return models.Sessions{Sessions: out}, nil
}

func (m *SqlRepo) GetSessionsByUserId(id uint) (models.Sessions, error) {
	var previous, upcoming models.Sessions
	var err error
	var g errgroup.Group
	g.Go(func() error {
		if previous, err = m.GetPreviousSessionsByUserIdAndDate(id, time.Now().Unix()); err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		if upcoming, err = m.GetUpcomingSessionsByUserIdAndDate(id, time.Now().Unix()); err != nil {
			return err
		}
		return nil
	})

	// Wait for all fetches to complete.
	if err = g.Wait(); err != nil {
		return models.Sessions{Sessions: nil}, err
	}
	out := append(previous.Sessions, upcoming.Sessions...)

	return models.Sessions{Sessions: out}, nil
}

func fmtLikeStatement(tag string) string {
	return "%" + strings.ToLower(tag) + "%"
}
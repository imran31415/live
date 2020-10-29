package repo

import "admin/models"

func (m *SqlRepo) CreateZoomAccessToken(z *models.ZoomToken) (*models.ZoomToken, error) {
	c := &models.ZoomToken{}
	if err := m.DB.Where("user_id = ?", z.UserId).Delete(&c).Error; err != nil {
		return nil, err
	}
	err := m.DB.Create(z).Error
	if err != nil {
		return nil, err
	}

	out := &models.ZoomToken{}

	if err = m.DB.Where("access_token = ?", z.AccessToken).First(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (m *SqlRepo) DeleteZoomAccessTokensByUserId(userId int64) error {
	if err := m.DB.Where("user_id = ?", userId).Delete(&models.ZoomToken{}).Error; err != nil {
		return err
	}
	return nil
}

func (m *SqlRepo) GetZoomTokenById(id uint) (*models.ZoomToken, error) {
	z := &models.ZoomToken{}
	if err := m.DB.Where("id = ?", id).First(&z).Error; err != nil {
		return nil, err
	}
	return z, nil
}

func (m *SqlRepo) GetZoomTokenByUserId(id uint) (*models.ZoomToken, error) {
	z := &models.ZoomToken{}
	if err := m.DB.Where("user_id = ?", id).First(&z).Error; err != nil {
		return nil, err
	}
	return z, nil
}

func (m *SqlRepo) UpdateZoomTokenById(u *models.ZoomToken) (*models.ZoomToken, error) {

	if err := m.DB.Where("id = ?", u.ID).Find(&models.ZoomToken{}).Updates(u).Error; err != nil {
		return nil, err
	}
	out, err := m.GetZoomTokenById(u.ID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

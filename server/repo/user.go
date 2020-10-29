package repo

import (
	"admin/models"
	"github.com/jinzhu/gorm"
)

func (m *SqlRepo) GetUserById(id uint) (*models.User, error) {
	u := &models.User{}
	if err := m.DB.Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (m *SqlRepo) GetUserByEmail(email string) (*models.User, error) {
	u := &models.User{}
	if err := m.DB.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (m *SqlRepo) GetUserByZoomAccountId(id string) (*models.User, error) {
	u := &models.User{}
	if err := m.DB.Where("zoom_account_id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (m *SqlRepo) GetUsersById(ids []uint) (models.Users, error) {
	var out []*models.User
	if err := m.DB.Where("id in (?)", ids).Find(&out).Error; err != nil {
		return models.Users{}, err
	}
	return models.Users{Users: out}, nil
}

func (m *SqlRepo) UpdateUserBySubId(u *models.User) (*models.User, error) {

	if err := m.DB.Where("auth_zero_sub_id = ?", u.AuthZeroSubId).Find(&models.User{}).Updates(u).Error; err != nil {
		return nil, err
	}
	out, err := m.GetOrCreateUserBySubId(u.AuthZeroSubId)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *SqlRepo) UpdateUserZoomDeAuthorized(id uint) error {

	if err := m.DB.Where("id = ?", id).Find(&models.User{}).Updates(map[string]interface{}{"zoom_app_installed": false, "zoom_account_id": ""}).Error; err != nil {
		return err
	}
	return nil
}

func (m *SqlRepo) UpdateUserById(u *models.User) (*models.User, error) {

	if err := m.DB.Where("id = ?", u.ID).Find(&models.User{}).Updates(u).Error; err != nil {
		return nil, err
	}
	out, err := m.GetUserById(u.ID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *SqlRepo) GetOrCreateUserBySubId(id string) (*models.User, error) {
	found := &models.User{}
	err := m.DB.Where("auth_zero_sub_id = ?", id).First(&found).Error

	switch err {
	case gorm.ErrRecordNotFound:
		c := &models.User{AuthZeroSubId: id}
		err = m.DB.Create(c).Error
		if err != nil {
			return nil, err
		}
		return c, nil

	case nil:
		return found, nil
	default:
		return nil, err
	}
}

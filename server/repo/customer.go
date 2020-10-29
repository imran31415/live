package repo

import (
	"admin/models"
	"github.com/jinzhu/gorm"
)

func (m *SqlRepo) GetCustomerById(id uint) (*models.Customer, error) {
	c := &models.Customer{}
	if err := m.DB.Where("id = ?", id).First(&c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (m *SqlRepo) GetCustomersByIds(ids []uint) (models.Customers, error) {
	var out []*models.Customer
	if err := m.DB.Where("id in (?)", ids).Find(&out).Error; err != nil {
		return models.Customers{}, err
	}
	return models.Customers{Customers: out}, nil
}

func (m *SqlRepo) UpdateCustomer(c *models.Customer) (*models.Customer, error) {
	if err := m.DB.Where("id = ?", c.ID).Find(&models.Customer{}).Updates(c).Error; err != nil {
		return nil, err
	}
	out, err := m.GetCustomerById(c.ID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *SqlRepo) GetOrCreateCustomerByUserId(id uint) (*models.Customer, error) {
	found := &models.Customer{}
	err := m.DB.Where("user_id = ?", id).First(&found).Error

	switch err {
	case gorm.ErrRecordNotFound:
		c := &models.Customer{UserId: int64(id)}
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

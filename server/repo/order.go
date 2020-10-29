package repo

import (
	"admin/models"
	pb "admin/protos"
)

func (m *SqlRepo) GetOrderByPaymentPlatformOrderId(id string) (*models.Order, error) {
	o := &models.Order{}
	if err := m.DB.Where("payment_platform_order_id = ?", id).First(&o).Error; err != nil {
		return nil, err
	}
	return o, nil
}

func (m *SqlRepo) GetOrderById(id uint) (*models.Order, error) {
	o := &models.Order{}
	if err := m.DB.Where("id = ?", id).First(&o).Error; err != nil {
		return nil, err
	}
	return o, nil
}

func (m *SqlRepo) UpdateOrderStatusByOrderId(o *models.Order) (*models.Order, error) {
	if err := m.DB.Where("id = ?", o.ID).Find(&models.Order{}).Updates(&models.Order{Status: o.Status}).Error; err != nil {
		return nil, err
	}
	out, err := m.GetOrderById(o.ID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m SqlRepo) CreateOrder(o *models.Order) (*models.Order, error) {
	err := m.DB.Create(o).Error
	if err != nil {
		return nil, err
	}

	out := &models.Order{}

	if err = m.DB.Where("payment_platform_order_id = ?", o.PaymentPlatformOrderId).First(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (m *SqlRepo) UpdateOrderPaymentMethodId(orderId uint, paymentMethodId string) (*models.Order, error) {
	if err := m.DB.Where("id = ?", orderId).Find(&models.Order{}).Updates(&models.Order{PaymentMethodId: paymentMethodId}).Error; err != nil {
		return nil, err
	}
	out, err := m.GetOrderById(orderId)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *SqlRepo) GetSucceededOrdersByCustomerId(id uint) (models.Orders, error) {
	var out []*models.Order
	if err := m.DB.Where("customer_id = ? AND status = ?", id, pb.Order_SUCCEEDED.String()).Limit(ordersLimit).Find(&out).Error; err != nil {
		return models.Orders{}, err
	}
	return models.Orders{Orders: out}, nil
}

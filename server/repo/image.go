package repo

import (
	"admin/models"
	pb "admin/protos"
)

func (m *SqlRepo) UpdateImageStatus(c *models.Image) (*models.Image, error) {
	if err := m.DB.Where("object_id = ?", c.ObjectId).Find(&models.Image{}).Updates(c).Error; err != nil {
		return nil, err
	}
	out, err := m.GetImageByObjectId(c.ObjectId)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *SqlRepo) CreateImage(i *models.Image) (*models.Image, error) {
	if err := m.DB.Create(i).Error; err != nil {
		return nil, err
	}
	return i, nil
}

func (m *SqlRepo) GetImageByObjectId(objectId string) (*models.Image, error) {
	c := &models.Image{}
	if err := m.DB.Where("object_id = ?", objectId).First(&c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (m *SqlRepo) GetImagesByUserId(userId uint) (models.Images, error) {
	var out []*models.Image
	if err := m.DB.Where("user_id = ? AND status = ?", userId, pb.UploadStatus_SUCCEEDED.Enum().String()).Order("created_at desc").Limit(50).Find(&out).Error; err != nil {
		return models.Images{Images: nil}, err
	}
	return models.Images{Images: out}, nil
}

package repository

import (
	"FCH/internal/models"

	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(msg *models.Message) error
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(msg *models.Message) error {
	return r.db.Create(msg).Error
}

package repository

import (
	"FCH/internal/models"

	"gorm.io/gorm"
)

type MessageRepository interface {
	FindByChatID(chatID uint) ([]models.Message, error)
	Create(msg *models.Message) error
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) FindByChatID(chatID uint) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Where("chat_id = ?", chatID).Order("created_at ASC").Find(&messages).Error
	return messages, err
}

func (r *messageRepository) Create(msg *models.Message) error {
	return r.db.Create(msg).Error
}

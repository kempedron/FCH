package repository

import (
	"FCH/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ParticipantRepository interface {
	Create(participant *models.ChatParticipants) error
	IsUserInChat(chatID, userID uint) (bool, error)
	FindByChatID(chatID uint) ([]models.ChatParticipants, error)
	AddUserToGroup(chatID, userID uint) error
}

type participantRepository struct {
	db *gorm.DB
}

func NewParticipantRepository(db *gorm.DB) ParticipantRepository {
	return &participantRepository{db: db}
}

func (r *participantRepository) Create(p *models.ChatParticipants) error {
	return r.db.Create(p).Error
}

func (r *participantRepository) IsUserInChat(chatID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.ChatParticipants{}).
		Where("chat_id = ? AND user_id = ? AND deleted_at IS NULL", chatID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *participantRepository) FindByChatID(chatID uint) ([]models.ChatParticipants, error) {
	var participants []models.ChatParticipants
	err := r.db.Where("chat_id = ? AND deleted_at IS NULL", chatID).Find(&participants).Error
	return participants, err
}

func (r *participantRepository) AddUserToGroup(chatID, userID uint) error {
	participant := models.ChatParticipants{
		ChatID: chatID,
		UserID: userID,
		Role:   "member",
	}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&participant).Error
}

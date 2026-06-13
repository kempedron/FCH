package repository

import (
	"FCH/internal/models"

	"gorm.io/gorm"
)

type ChatRepository interface {
	FindByID(id uint) (*models.Chat, error)
	FindByIDWithMessages(id, userID uint) (*models.Chat, error)
	FindPersonalChat(userID1, userID2 uint) (*models.Chat, error)
	FindUserChats(userID uint) ([]models.Chat, error)
	Create(chat *models.Chat) error
	FindByInviteCode(code string) (*models.Chat, error)
	ExistsPersonalChat(userID1, userID2 uint) (bool, *models.Chat)
	FindByIDAndInviteCode(id uint, code string) (*models.Chat, error)
	CreateGroupWithTransaction(name string, creatorID uint, inviteCode string) (uint, error)
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) FindByID(id uint) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.Preload("Participants.User").First(&chat, id).Error
	return &chat, err
}

func (r *chatRepository) FindByIDWithMessages(id, userID uint) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.Preload("Participants.User").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("id IN (?)", r.db.Table("chat_participants").
			Select("chat_id").
			Where("user_id = ? AND deleted_at IS NULL", userID),
		).
		First(&chat, id).Error
	return &chat, err
}

func (r *chatRepository) FindPersonalChat(userID1, userID2 uint) (*models.Chat, error) {
	var chat models.Chat
	havingCount := 2
	if userID1 == userID2 {
		havingCount = 1
	}
	subquery := r.db.Model(&models.ChatParticipants{}).
		Select("chat_id").
		Joins("JOIN chats ON chats.id = chat_participants.chat_id").
		Where("chat_participants.user_id IN ? AND chats.is_group = ? AND chat_participants.deleted_at IS NULL",
			[]uint{userID1, userID2}, false).
		Group("chat_id").
		Having("COUNT(DISTINCT chat_participants.user_id) = ?", havingCount)

	err := r.db.
		Preload("Participants.User").
		Preload("Messages").
		Where("id IN (?)", subquery).
		First(&chat).Error
	return &chat, err
}

func (r *chatRepository) FindUserChats(userID uint) ([]models.Chat, error) {
	var chats []models.Chat
	err := r.db.
		Preload("Participants.User").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("messages.created_at DESC")
		}).
		Joins("JOIN chat_participants ON chat_participants.chat_id = chats.id").
		Where("chat_participants.user_id = ?", userID).
		Where("chat_participants.deleted_at IS NULL").
		Distinct("chats.*").
		Order("chats.updated_at DESC").
		Find(&chats).Error
	return chats, err
}

func (r *chatRepository) Create(chat *models.Chat) error {
	return r.db.Create(chat).Error
}

func (r *chatRepository) FindByInviteCode(code string) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.
		Where("invite_code = ? AND is_group = ? AND deleted_at IS NULL", code, true).
		First(&chat).Error
	return &chat, err
}

func (r *chatRepository) ExistsPersonalChat(userID1, userID2 uint) (bool, *models.Chat) {
	var chat models.Chat
	err := r.db.Table("chats").
		Select("chats.*").
		Joins("JOIN chat_participants p1 ON p1.chat_id = chats.id").
		Joins("JOIN chat_participants p2 ON p2.chat_id = chats.id").
		Where("((p1.user_id = ? AND p2.user_id = ?) OR (p1.user_id = ? AND p2.user_id = ?))",
			userID1, userID2, userID2, userID1).
		Where("chats.is_group = ?", false).
		Where("p1.deleted_at IS NULL AND p2.deleted_at IS NULL AND chats.deleted_at IS NULL").
		First(&chat).Error
	return err == nil, &chat
}

func (r *chatRepository) FindByIDAndInviteCode(id uint, code string) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.Where("id = ? AND invite_code = ? AND is_group = ? AND deleted_at IS NULL", id, code, true).First(&chat).Error
	return &chat, err
}

// Переносим транзакцию БД на уровень репозитория, где ей и место
func (r *chatRepository) CreateGroupWithTransaction(name string, creatorID uint, inviteCode string) (uint, error) {
	var groupID uint
	err := r.db.Transaction(func(tx *gorm.DB) error {
		newChat := models.Chat{
			Name:       name,
			IsGroup:    true,
			CreatorID:  creatorID,
			InviteCode: inviteCode,
		}
		if err := tx.Create(&newChat).Error; err != nil {
			return err
		}

		participant := models.ChatParticipants{
			ChatID: newChat.ID,
			UserID: creatorID,
			Role:   "admin",
		}
		if err := tx.Create(&participant).Error; err != nil {
			return err
		}
		groupID = newChat.ID
		return nil
	})
	return groupID, err
}

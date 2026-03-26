package service

import (
	"FCH/internal/database"
	"FCH/internal/middleware"
	"FCH/internal/models"
	"errors"
	"log"

	"gorm.io/gorm"
)

func GetChatInfo(username string, myID uint) (*models.Chat, error) {
	var chat models.Chat
	log.Printf("username: %s myID: %d", username, myID)
	err := database.DB.Preload("Participants").
		Preload("Messages").
		Preload("Participants.User").
		Where("name = ?", username).
		First(&chat).Error
	if err != nil {
		log.Printf("error: %s", err)
		if err == gorm.ErrRecordNotFound {
			newChat, err := CreateChat(middleware.GetIDByUsername(username), myID)
			if err != nil {
				return &models.Chat{}, errors.New("failed to create chat")
			}
			return newChat, nil
		}
		return &models.Chat{}, errors.New("chat not found")
	}
	return &chat, nil
}

func CreateChat(userID uint, myID uint) (*models.Chat, error) {
	username := middleware.GetUsernameByID(userID)
	log.Printf("username: %s", username)
	if username == "" {
		return nil, errors.New("user not found")
	}
	chat := &models.Chat{
		Name:         username,
		Participants: []models.ChatParticipants{{UserID: userID, IsGroup: false}, {UserID: myID, IsGroup: false}},
		Messages:     []models.Message{},
	}
	err := database.DB.Create(chat).Error
	return chat, err
}

func GetOrCreatePersonalChat(myID, opponentID uint) (*models.Chat, error) {
	var chat models.Chat

	err := database.DB.Table("chats").
		Select("chats.*").
		Joins("JOIN chat_participants p1 ON p1.chat_id = chats.id").
		Joins("JOIN chat_participants p2 ON p2.chat_id = chats.id").
		Where("p1.user_id = ? AND p2.user_id = ?", myID, opponentID).
		Where("p1.is_group = ? AND p2.is_group = ?", false, false).
		Where("p1.deleted_at IS NULL AND p2.deleted_at IS NULL AND chats.deleted_at IS NULL").
		First(&chat).Error

	if err == nil {
		return &chat, nil
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			newChat := models.Chat{Name: "Personal Chat"}
			if err := database.DB.Create(&newChat).Error; err != nil {
				return nil, err
			}

			p1 := models.ChatParticipants{ChatID: newChat.ID, UserID: myID, IsGroup: false}
			p2 := models.ChatParticipants{ChatID: newChat.ID, UserID: opponentID, IsGroup: false}
			database.DB.Create(&p1)
			database.DB.Create(&p2)

			return &newChat, nil
		}
		return nil, err
	}

	return &chat, nil
}

func GetMyChats(myID uint) ([]models.Chat, error) {
	var chats []models.Chat
	err := database.DB.
		// Предзагружаем участников и их данные
		Preload("Participants.User").
		// Предзагружаем сообщения с сортировкой
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("messages.created_at DESC")
		}).
		// Используем Joins с условием, чтобы GORM понимал связь
		Joins("JOIN chat_participants ON chat_participants.chat_id = chats.id").
		// Явно указываем условия
		Where("chat_participants.user_id = ?", myID).
		Where("chat_participants.deleted_at IS NULL").
		// Группируем или используем Distinct, чтобы избежать дублей
		Distinct("chats.*").
		Order("chats.updated_at DESC").
		Find(&chats).Error
	return chats, err
}

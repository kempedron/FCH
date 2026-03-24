package service

import (
	"FCH/internal/database"
	"FCH/internal/middleware"
	"FCH/internal/models"
	"errors"

	"gorm.io/gorm"
)

func GetChatInfo(username string, myID uint) (models.Chat, error) {
	var chat models.Chat
	err := database.DB.Preload("Participants").
		Preload("Messages").
		Preload("Participants.User").
		Where("name = ?", username).
		First(&chat).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err := CreateChat(middleware.GetIDByUsername(username), myID)
			if err != nil {
				return models.Chat{}, err
			}
			return models.Chat{}, nil
		}
		return models.Chat{}, errors.New("chat not found")
	}
	return chat, nil
}

func CreateChat(userID uint, myID uint) error {
	username := middleware.GetUsernameByID(userID)
	if username == "" {
		return errors.New("user not found")
	}
	chat := &models.Chat{
		Name:         username,
		Participants: []models.ChatParticiant{{UserID: userID, IsGroup: false}, {UserID: myID, IsGroup: false}},
		Messages:     []models.Message{},
	}
	err := database.DB.Create(chat).Error
	return err
}

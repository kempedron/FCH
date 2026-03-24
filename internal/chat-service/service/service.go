package service

import (
	"FCH/internal/database"
	"FCH/internal/middleware"
	"FCH/internal/models"
	"errors"
	"log"

	"gorm.io/gorm"
)

func GetChatInfo(username string, myID uint) (models.Chat, error) {
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
			err := CreateChat(middleware.GetIDByUsername(username), myID)
			if err != nil {
				return models.Chat{}, errors.New("failed to create chat")
			}
			return models.Chat{}, nil
		}
		return models.Chat{}, errors.New("chat not found")
	}
	return chat, nil
}

func CreateChat(userID uint, myID uint) error {
	username := middleware.GetUsernameByID(userID)
	log.Printf("username: %s", username)
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

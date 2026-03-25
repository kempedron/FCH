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
		Participants: []models.ChatParticiant{{UserID: userID, IsGroup: false}, {UserID: myID, IsGroup: false}},
		Messages:     []models.Message{},
	}
	err := database.DB.Create(chat).Error
	return chat, err
}

func SaveMessageToDB()

func GetOrCreatePersonalChat(myID, opponentID uint) (*models.Chat, error) {
	var chat models.Chat

	err := database.DB.Table("chats").
		Joins("JOIN chat_particiants p1 ON p1.chat_id = chats.id").
		Joins("JOIN chat_particiants p2 ON p2.chat_id = chats.id").
		Where("p1.user_id = ? AND p2.user_id = ? AND p1.is_group = ?", myID, opponentID, false).
		First(&chat).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			newChat := models.Chat{Name: "Personal Chat"}
			if err := database.DB.Create(&newChat).Error; err != nil {
				return nil, err
			}

			p1 := models.ChatParticiant{ChatID: newChat.ID, UserID: myID, IsGroup: false}
			p2 := models.ChatParticiant{ChatID: newChat.ID, UserID: opponentID, IsGroup: false}
			database.DB.Create(&p1)
			database.DB.Create(&p2)

			return &newChat, nil
		}
		return nil, err
	}

	return &chat, nil
}

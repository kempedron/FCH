package service

import (
	"FCH/internal/database"
	"FCH/internal/middleware"
	"FCH/internal/models"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		Participants: []models.ChatParticipants{{UserID: userID}, {UserID: myID}},
		Messages:     []models.Message{},
	}
	err := database.DB.Create(chat).Error
	return chat, err
}

func GetOrCreatePersonalChat(myID, opponentID uint) (*models.Chat, error) {
	var chat models.Chat

	havingCount := 2
	if myID == opponentID {
		havingCount = 1
	}

	subquery := database.DB.Model(&models.ChatParticipants{}).
		Select("chat_participants.chat_id").
		Joins("JOIN chats ON chats.id = chat_participants.chat_id").
		Where("chat_participants.user_id IN ? AND chats.is_group = ? AND chat_participants.deleted_at IS NULL",
			[]uint{myID, opponentID}, false).
		Group("chat_participants.chat_id").
		Having("COUNT(DISTINCT chat_participants.user_id) = ?", havingCount)

	err := database.DB.
		Preload("Participants.User").
		Preload("Messages").
		Where("id IN (?)", subquery).
		First(&chat).Error
	if err == nil {
		return &chat, nil
	}
	if err == gorm.ErrRecordNotFound {
		newChat := models.Chat{
			Name:       "Personal Chat",
			InviteCode: GenerateInviteCode(),
		}
		if err := database.DB.Create(&newChat).Error; err != nil {
			return nil, err
		}
		p1 := models.ChatParticipants{ChatID: newChat.ID, UserID: myID}
		p2 := models.ChatParticipants{ChatID: newChat.ID, UserID: opponentID}

		if err := database.DB.Create(&p1).Error; err != nil {
			return nil, err
		}
		if err := database.DB.Create(&p2).Error; err != nil {
			return nil, err
		}
		return &newChat, nil
	}

	return nil, err
}

func GetMyChats(myID uint) ([]models.Chat, error) {
	var chats []models.Chat
	err := database.DB.
		Preload("Participants.User").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("messages.created_at DESC")
		}).
		Joins("JOIN chat_participants ON chat_participants.chat_id = chats.id").
		Where("chat_participants.user_id = ?", myID).
		Where("chat_participants.deleted_at IS NULL").
		Distinct("chats.*").
		Order("chats.updated_at DESC").
		Find(&chats).Error
	return chats, err
}

func IsChatExist(myID, opponentID uint) (bool, models.Chat) {
	var chat models.Chat
	err := database.DB.Table("chats").
		Select("chats.*").
		Joins("JOIN chat_participants p1 ON p1.chat_id = chats.id").
		Joins("JOIN chat_participants p2 ON p2.chat_id = chats.id").
		Where("((p1.user_id = ? AND p2.user_id = ?) OR (p1.user_id = ? AND p2.user_id = ?))",
			myID, opponentID, opponentID, myID).
		Where("chats.is_group = ?", false).
		Where("p1.deleted_at IS NULL AND p2.deleted_at IS NULL AND chats.deleted_at IS NULL").
		First(&chat).Error
	return err == nil, chat
}

func AddUserToGroupChat(userID uint, groupChatID uint) error {
	var chat models.Chat

	if err := database.DB.First(&chat, groupChatID).Error; err != nil {
		return err
	}

	if !chat.IsGroup {
		return errors.New("нельзя добавить пользователя в личный чат напрямую")
	}

	if IsUserInGroup(userID, chat.Participants) {
		return errors.New("пользователь уже находится в этом чате")
	}

	newParticipant := models.ChatParticipants{
		ChatID: groupChatID,
		UserID: userID,
		Role:   "member",
	}

	return database.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&newParticipant).Error
}

func IsUserInGroup(userID uint, groupChat []models.ChatParticipants) bool {
	for _, val := range groupChat {
		if val.UserID == userID {
			return true
		}
	}
	return false
}

func GenerateInviteCode() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

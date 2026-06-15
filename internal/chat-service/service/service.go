package service

import (
	"FCH/internal/models"
	"FCH/internal/repository"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"gorm.io/gorm"
)

type ChatService interface {
	GetOrCreatePersonalChat(myID, opponentID uint) (*models.Chat, error)
	GetMyChats(myID uint) ([]models.Chat, error)
	IsChatExist(myID, opponentID uint) (bool, *models.Chat)
	GetChatByID(chatID, userID uint) (*models.Chat, error)
	CreateMessage(chatID, authorID uint, content string) (*models.Message, error)
	GetParticipantIDs(chatID uint) ([]uint, error)
	CreateGroupChat(name string, creatorID uint) (uint, error)
	JoinToGroupChat(chatID uint, inviteCode string, userID uint) error
	IsUserInChat(chatID, userID uint) (bool, error)
}

type service struct {
	chatRepo repository.ChatRepository
	partRepo repository.ParticipantRepository
	msgRepo  repository.MessageRepository
}

func NewService(
	chatRepo repository.ChatRepository,
	partRepo repository.ParticipantRepository,
	msgRepo repository.MessageRepository) *service {
	return &service{chatRepo: chatRepo, partRepo: partRepo, msgRepo: msgRepo}
}

func (s *service) GetOrCreatePersonalChat(myID, opponentID uint) (*models.Chat, error) {
	chat, err := s.chatRepo.FindPersonalChat(myID, opponentID)
	if err == nil {
		return chat, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	newChat := models.Chat{
		Name:       "Personal Chat",
		InviteCode: GenerateInviteCode(),
	}
	if err := s.chatRepo.Create(&newChat); err != nil {
		return nil, err
	}

	p1 := models.ChatParticipants{ChatID: newChat.ID, UserID: myID}
	p2 := models.ChatParticipants{ChatID: newChat.ID, UserID: opponentID}
	if err := s.partRepo.Create(&p1); err != nil {
		return nil, err
	}
	if err := s.partRepo.Create(&p2); err != nil {
		return nil, err
	}

	return &newChat, nil
}

func (s *service) GetMyChats(myID uint) ([]models.Chat, error) {
	return s.chatRepo.FindUserChats(myID)
}

func (s *service) IsChatExist(myID, opponentID uint) (bool, *models.Chat) {
	return s.chatRepo.ExistsPersonalChat(myID, opponentID)
}

func (s *service) IsUserInChat(chatID, userID uint) (bool, error) {
	return s.partRepo.IsUserInChat(chatID, userID)
}

func (s *service) GetChatByID(chatID, userID uint) (*models.Chat, error) {
	return s.chatRepo.FindByIDWithMessages(chatID, userID)
}

func (s *service) CreateMessage(chatID, authorID uint, content string) (*models.Message, error) {
	ok, err := s.partRepo.IsUserInChat(chatID, authorID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("access denied")
	}

	msg := &models.Message{
		ChatID:   chatID,
		AuthorID: authorID,
		Content:  content,
	}
	if err := s.msgRepo.Create(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *service) GetParticipantIDs(chatID uint) ([]uint, error) {
	participants, err := s.partRepo.FindByChatID(chatID)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, len(participants))
	for i, p := range participants {
		ids[i] = p.UserID
	}
	return ids, nil
}

func (s *service) CreateGroupChat(name string, creatorID uint) (uint, error) {
	inviteCode := GenerateInviteCode()
	return s.chatRepo.CreateGroupWithTransaction(name, creatorID, inviteCode)
}

func (s *service) JoinToGroupChat(chatID uint, inviteCode string, userID uint) error {
	chat, err := s.chatRepo.FindByIDAndInviteCode(chatID, inviteCode)
	if err != nil {
		return errors.New("invalid or expired invite link")
	}
	if !chat.IsGroup {
		return errors.New("not a group chat")
	}
	ok, err := s.partRepo.IsUserInChat(chatID, userID)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return s.partRepo.AddUserToGroup(chatID, userID)
}

func GenerateInviteCode() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

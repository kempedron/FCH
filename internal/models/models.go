package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(255);unique" json:"username"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`

	Chats    []ChatParticipants `gorm:"foreignKey:UserID"`
	Messages []Message          `gorm:"foreignKey:AuthorID"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

type Chat struct {
	gorm.Model
	Name         string             `gorm:"type:varchar(255)" json:"name"`
	Participants []ChatParticipants `gorm:"foreignKey:ChatID"`
	Messages     []Message          `gorm:"foreignKey:ChatID"`
	IsGroup      bool               `gorm:"type:boolean;not null;default:false"`
	CreatorID    uint               `json:"creator_id"`
}

func (Chat) TableName() string {
	return "chats"
}

type ChatParticipants struct {
	gorm.Model
	ChatID     uint   `gorm:"type:index;not null"`
	UserID     uint   `gorm:"type:index;not null"`
	Role       string `gorm:"type:varchar(255);not null;default:'member'"`
	LastReadAt time.Time
	Chat       Chat `gorm:"foreignKey:ChatID"`
	User       User `gorm:"foreignKey:UserID"`
}

func (ChatParticipants) TableName() string {
	return "chat_participants"
}

type Message struct {
	gorm.Model
	ChatID      uint         `gorm:"type:index;not null"`
	AuthorID    uint         `gorm:"type:index;not null"`
	Content     string       `gorm:"type:text;not null"`
	Attachments []Attachment `gorm:"foreignKey:MessageID"`
	Chat        Chat         `gorm:"foreignKey:ChatID"`
	Author      User         `gorm:"foreignKey:AuthorID;"`
}

func (Message) TableName() string {
	return "messages"
}

type Attachment struct {
	gorm.Model
	MessageID   uint   `gorm:"index;not null"`
	FileName    string `gorm:"type:varchar(255);not null"`
	FileSize    int64
	MimeType    string  `gorm:"type:varchar(100)"`
	StoragePath string  `gorm:"type:varchar(512);not null"`
	Message     Message `gorm:"foreignKey:MessageID"`
}

func (Attachment) TableName() string {
	return "attachments"
}

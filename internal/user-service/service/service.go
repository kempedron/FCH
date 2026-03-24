package service

import (
	"FCH/internal/database"
	"FCH/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func Login(username, password string) *models.User {
	var user models.User
	database.DB.Where("username = ?", username).First(&user)
	if err := user.CheckPassword(password); err != nil {
		return nil
	}

	return &user
}

func Register(username, password string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err == nil {
		return nil, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}
	if err := database.DB.Create(&newUser).Error; err != nil {
		return nil, err
	}

	return &newUser, nil
}

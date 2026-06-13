package service

import (
	"FCH/internal/database"
	"FCH/internal/models"
	"FCH/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var db, err = database.InitDB()

type Service struct {
	userRepo repository.UserRepository
}

func NewService(repo repository.UserRepository) *Service {
	return &Service{userRepo: repo}
}

func (s *Service) Login(username, password string) *models.User {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil
	}
	if err := user.CheckPassword(password); err != nil {
		return nil
	}
	return user
}

func (s *Service) Register(username, password string) (*models.User, error) {
	existing, _ := s.userRepo.FindByUsername(username)
	if existing != nil {
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
	if err := s.userRepo.Create(&newUser); err != nil {
		return nil, err
	}

	return &newUser, nil
}

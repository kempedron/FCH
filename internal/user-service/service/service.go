package service

import (
	"FCH/internal/models"
	"FCH/internal/repository"
)

type UserService interface {
	Login(string, string) *models.User
	Register(string, string) (*models.User, error)
	SearchByUsername(string) (*models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewService(repo repository.UserRepository) UserService {
	return &userService{userRepo: repo}
}

func (s *userService) Login(username, password string) *models.User {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil
	}
	if err := user.CheckPassword(password); err != nil {
		return nil
	}
	return user
}

func (s *userService) Register(username, password string) (*models.User, error) {
	existing, _ := s.userRepo.FindByUsername(username)
	if existing != nil {
		return nil, nil
	}

	newUser := models.User{Username: username}
	if err := newUser.HashPassword(password); err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(&newUser); err != nil {
		return nil, err
	}

	return &newUser, nil
}

func (s *userService) SearchByUsername(username string) (*models.User, error) {
	return s.userRepo.FindByUsername(username)
}

package repository

import (
	"FCH/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(id uint) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	Create(user *models.User) error
}

type userRepo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepo) GetUsernameByID(userID uint) string {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return ""
	}
	return user.Username
}

func (r *userRepo) GetIDByUsername(username string) uint {
	var user models.User
	if err := r.db.First(&user, "username = ?", username).Error; err != nil {
		return 0
	}
	return user.ID
}

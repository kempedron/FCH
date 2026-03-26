package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"FCH/internal/models"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	fmt.Println("dsn: ", dsn)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("не удалось подключиться к БД:", err)
	}
	err = DB.AutoMigrate(
		&models.User{},
		&models.Chat{},
		&models.ChatParticipants{},
		&models.Message{},
		&models.Attachment{},
	)

	if err != nil {
		log.Printf("error migrate DB: %s", err)
		panic(err)
	}

	fmt.Println("SUCCESSFYLLU INIT DATABASE!")

	return nil

}

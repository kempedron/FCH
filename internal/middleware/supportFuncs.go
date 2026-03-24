package middleware

import (
	"FCH/internal/database"
	"FCH/internal/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func GetParamByUrl(name string, r *http.Request) int {
	vars := mux.Vars(r)
	Id := vars[name]
	IdInt, err := strconv.Atoi(Id)
	if err != nil {
		log.Printf("error convert id to int: %s", err)
	}
	return IdInt
}

func GetUsernameByID(userID uint) string {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return ""
	}
	return user.Username
}

func GetIDByUsername(username string) uint {
	var user models.User
	if err := database.DB.First(&user, "username = ?", username).Error; err != nil {
		return 0
	}
	return user.ID
}

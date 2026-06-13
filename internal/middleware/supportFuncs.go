package middleware

import (
	"FCH/internal/database"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var db, err = database.InitDB()

func GetParamByUrl(name string, r *http.Request) int {
	vars := mux.Vars(r)
	Id := vars[name]
	IdInt, err := strconv.Atoi(Id)
	if err != nil {
		log.Printf("error convert id to int: %s", err)
	}
	return IdInt
}

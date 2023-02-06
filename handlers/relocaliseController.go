package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func MakeRelocaliseControllerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sceneName := vars["name"]
		log.Print("receive user file request: ", sceneName)
		defer r.Body.Close()
	}
}

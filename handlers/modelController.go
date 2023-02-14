package handlers

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func MakeModelControllerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sceneName := vars["name"]
		log.Print("[MakeModelControllerHandler] model controller request: ", sceneName)
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		log.Print("[MakeModelControllerHandler] this is body: ", string(body))
		if string(body) != "OK" {
			log.Fatal(body)
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			// Add scene to candidate list, the scene name must be unique
			ScenesListLock.Lock()
			ProcessingScenesIndex[sceneName] = len(ProcessingScenesList)
			ProcessingScenesList = append(ProcessingScenesList, sceneName)
			ScenesListLock.Unlock()
		}

		w.WriteHeader(http.StatusOK)
	}
}

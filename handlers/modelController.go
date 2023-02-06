package handlers

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func MakeModelControllerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sceneName := vars["name"]
		log.Print("model controller request: ", sceneName)
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		if string(body) != "OK" {
			log.Fatal(body)
		} else {
			// Add scene to candidate list
			ScenesListLock.Lock()
			PrepareScenesList = append(PrepareScenesList, sceneName)
			sceneIndex++
			ScenesListLock.Unlock()

			// If we have more than 1 scene to relocalise,
			// send request to relocalise controller to start relocalisation
			length := len(PrepareScenesList)
			if length > 1 {
				// randomly choose scene to relocalise
				rand.Seed(time.Now().Unix())
				var index1, index2 int
				index1 = rand.Intn(length)
				if index1 == length-1 {
					index1, index2 = index1-1, index1
				} else {
					index2 = index1 + 1
				}
				name1, name2 := PrepareScenesList[index1], PrepareScenesList[index2]

				// delete candidates scene from PrepareScenesList and add them to ProcessingSceneList
				ScenesListLock.Lock()
				ProcessingSceneList = append(ProcessingSceneList, name1)
				ProcessingSceneList = append(ProcessingSceneList, name2)
				PrepareScenesList = append(PrepareScenesList[:index1], PrepareScenesList[index2+1:]...)
				ScenesListLock.Unlock()

				clientNO := ClientScenes[name1]
				clientIP := ClientAddrs[clientNO]
				url := clientIP + "/?name1=" + name1 + "&name2=" + name2

			}
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

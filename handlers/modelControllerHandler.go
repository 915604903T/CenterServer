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

		if string(body) == "Failed" {
			w.WriteHeader(http.StatusOK)
			log.Print("Render ", sceneName, "failed")
			return
		}

		TimeOutMapLock.RLock()
		timeout, ok := TimeOutMap[sceneName]
		TimeOutMapLock.RUnlock()
		// if not the rt scene, add it to the normal scene list
		if !ok {
			// Add scene to candidate list, the scene name must be unique
			ScenesListLock.Lock()
			ProcessingScenesIndex[sceneName] = len(ProcessingScenesList)
			ProcessingScenesList = append(ProcessingScenesList, sceneName)
			ScenesListLock.Unlock()
		} else {
			if !timeout.IsFinished {
				rtScene := RtScene{
					Name:       sceneName,
					ExpireTime: timeout.ExpireTime,
				}
				RtScenesListLock.Lock()
				RtProcessingScenesList = append(RtProcessingScenesList, rtScene)
				RtScenesListLock.Unlock()
			} else {
				log.Println("[MakeModelControllerHandler]", sceneName, "is TimeOut!!!!!!!!!")
			}

		}

		addr := body
		clientNO := ClientIpsMap[string(addr)]
		// fmt.Println("finishRendering!!!!!!!!!!!!!!!!!!name:", sceneName, "addr:", addr, "clientNO:", clientNO)

		ClientScenesLock.Lock()
		ClientScenes[sceneName] = map[int]bool{clientNO: true}
		ClientScenesLock.Unlock()

		// init a graph node
		sceneGraphLock.Lock()
		sceneGraph[sceneName] = make(map[string]Pose)
		sceneGraphLock.Unlock()

		// init a union element
		sceneUnionLock.Lock()
		sceneUnion.initInsert(sceneName)
		sceneUnionLock.Unlock()

		w.WriteHeader(http.StatusOK)
	}
}

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func processMeshInfo(info MeshInfo) {
	// cover origin scene info match
	originMeshes := make(map[*MeshInfo]bool)
	// find the match user for the scene
	var user *User
	var userName string
	for scene := range info.Scenes {
		SceneUserMapLock.RLock()
		userName = SceneUserMap[scene]
		SceneUserMapLock.RUnlock()
		break
	}
	UsersLock.RLock()
	user = Users[userName]
	UsersLock.RUnlock()

	user.SceneMeshLock.Lock()
	for scene := range info.Scenes {
		user.SceneMesh[scene] = &info
		originMeshes[&info] = true
	}
	user.SceneMeshLock.Unlock()

	RunningMeshesLock.Lock()
	for mesh := range originMeshes {
		delete(RunningMeshes, mesh)
	}
	RunningMeshesLock.Unlock()
}

func MakeMeshInfoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)
		log.Print("[MakeMeshInfoHandler] receive client merge finish request:\n", string(body))
		meshInfo := MeshInfo{}
		err := json.Unmarshal(body, &meshInfo)
		if err != nil {
			log.Println("unmarshal mesh info err: ", err)
			panic(err)
		}
		processMeshInfo(meshInfo)
		w.WriteHeader(http.StatusOK)
	}
}

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
	sceneMeshLock.Lock()
	for scene := range info.Scenes {
		sceneMesh[scene] = &info
		originMeshes[&info] = true
	}
	sceneMeshLock.Unlock()

	RunningMeshesLock.Lock()
	for mesh := range originMeshes {
		delete(RunningMeshes, mesh)
	}
	RunningMeshesLock.Unlock()
}

func MakeMeshInfoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("[MakeMeshInfoHandler] receive client merge finish request")
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)
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

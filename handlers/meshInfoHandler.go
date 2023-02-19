package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func processMeshInfo(info MeshInfo) {
	// cover origin scene info match
	sceneMeshLock.Lock()
	for scene := range info.Scenes {
		sceneMesh[scene] = info
	}
	sceneMeshLock.Unlock()
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

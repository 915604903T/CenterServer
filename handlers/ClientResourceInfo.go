package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func MakeClientResourceInfoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("receive client resource")
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		id, _ := strconv.Atoi(mux.Vars(r)["id"])
		resourceInfo := ResourceInfo{}
		err := json.Unmarshal(body, &resourceInfo)
		if err != nil {
			panic(err)
		}
		resourceInfoLock.Lock()
		ClientResourceStats[id] = resourceInfo
		resourceInfoLock.Unlock()
		w.WriteHeader(http.StatusOK)
	}
}

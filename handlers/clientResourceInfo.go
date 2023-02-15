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
		log.Print("[MakeClientResourceInfoHandler] receive client resource")
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		id, _ := strconv.Atoi(mux.Vars(r)["id"])
		// log.Print("[MakeClientResourceInfoHandler]", id, "client resource: ", string(body))
		resourceInfo := ResourceInfo{}
		err := json.Unmarshal(body, &resourceInfo)
		if err != nil {
			panic(err)
		}
		resourceInfoLock.Lock()
		ClientResourceStats[id-1] = resourceInfo
		resourceInfoLock.Unlock()
		w.WriteHeader(http.StatusOK)
	}
}

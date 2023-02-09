package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func MakeRelocaliseControllerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("relocalise global pose request")
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		poseInfo := globalPose{}
		err := json.Unmarshal(body, &poseInfo)
		if err != nil {
			log.Fatal("error de-serializing request body: ", body)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		isExit := false
		pair := [2]string{poseInfo.Scene1Name, poseInfo.Scene2Name}
		globalPoseLock.RLock()
		if _, ok := globalPoses[pair]; ok {
			isExit = true
		}
		globalPoseLock.RUnlock()
		if !isExit {
			globalPoseLock.Lock()
			globalPoses[pair] = [2]pose{poseInfo.Scene1Pose, poseInfo.Scene2Pose}
			globalPoseLock.Unlock()
		}
		log.Println("this is ", poseInfo.Scene1Name, poseInfo.Scene1Name, "global pose")
		log.Println(poseInfo)
		w.WriteHeader(http.StatusOK)
	}
}

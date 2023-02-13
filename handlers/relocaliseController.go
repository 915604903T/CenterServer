package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func MakeRelocaliseControllerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("relocalise global pose request")
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		bodyStr := string(body)
		log.Println("receive globalpose: ", bodyStr)
		if strings.Contains(bodyStr, "failed") {
			content := strings.Fields(bodyStr)
			scene1, scene2 := content[0], content[1]
			if FailedSceneList[scene1] == nil {
				FailedSceneList[scene1] = make(map[string]int)
				FailedSceneList[scene1][scene2] = 1
			} else {
				FailedSceneList[scene1][scene2]++
			}
			if FailedSceneList[scene2] == nil {
				FailedSceneList[scene2] = make(map[string]int)
				FailedSceneList[scene2][scene1] = 1
			} else {
				FailedSceneList[scene2][scene1]++
			}
			log.Println("add ", scene1, scene2, "to failedList")
			PrepareScenesList = append(PrepareScenesList, scene1)
			PrepareScenesList = append(PrepareScenesList, scene2)
			w.WriteHeader(http.StatusOK)
			return
		}

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
		log.Println("this is globalpose struct:\n", poseInfo)
		w.WriteHeader(http.StatusOK)
	}
}

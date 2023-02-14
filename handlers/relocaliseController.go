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
		log.Print("[MakeRelocaliseControllerHandler] relocalise global pose request!!!!!!!!!!!!!!!")
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
			log.Println("[MakeRelocaliseControllerHandler] add ", scene1, scene2, "to failedList")

			w.WriteHeader(http.StatusOK)
			return
		}

		poseInfo := globalPose{}
		err := json.Unmarshal(body, &poseInfo)
		if err != nil {
			log.Fatal("[MakeRelocaliseControllerHandler] error de-serializing request body: ", body)
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
		log.Println("[MakeRelocaliseControllerHandler] this is globalpose struct:\n", poseInfo)

		// delete processed scene from processing list and move to succeedScene
		ScenesListLock.Lock()

		SucceedSceneList = append(SucceedSceneList, poseInfo.Scene1Name)
		SucceedSceneList = append(SucceedSceneList, poseInfo.Scene2Name)

		index1, index2 := -1, -1
		if val, ok := ProcessingScenesIndex[poseInfo.Scene1Name]; ok {
			index1 = val
			delete(ProcessingScenesIndex, poseInfo.Scene1Name)
		}
		if val, ok := ProcessingScenesIndex[poseInfo.Scene2Name]; ok {
			index2 = val
			delete(ProcessingScenesIndex, poseInfo.Scene2Name)
		}
		if index1 != -1 {
			tmp := 0
			for i, v := range ProcessingScenesList {
				if i != index1 {
					ProcessingScenesList[tmp] = v
					tmp++
				}
			}
			ProcessingScenesList = ProcessingScenesList[:tmp]
		}
		if index2 != -1 {
			tmp := 0
			for i, v := range ProcessingScenesList {
				if i != index2 {
					ProcessingScenesList[tmp] = v
					tmp++
				}
			}
			ProcessingScenesList = ProcessingScenesList[:tmp]
		}
		if index1 != -1 || index2 != -1 {
			for i, v := range ProcessingScenesList {
				ProcessingScenesIndex[v] = i
			}
		}
		ScenesListLock.Unlock()

		w.WriteHeader(http.StatusOK)
	}
}

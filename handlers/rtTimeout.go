package handlers

import (
	"log"
	"time"
)

func DealRealTimeSceneTimeout() {
	for ; ; time.Sleep(time.Second) {
		RtScenesListLock.RLock()
		for i, rtScene := range RtProcessingScenesList {
			if time.Now().After(rtScene.ExpireTime) {
				RtScenesListLock.RUnlock()
				log.Println("[DealRealTimeSceneTimeout] ", rtScene.Name, "is timeout!!! Remove it from RtProcessingScenesList")
				RtScenesListLock.Lock()
				// remove timeout scene
				RtProcessingScenesList = append(RtProcessingScenesList[:i], RtProcessingScenesList[i+1:]...)
				RtScenesListLock.Unlock()

				TimeOutMapLock.Lock()
				sceneTimeout := TimeOutMap[rtScene.Name]
				sceneTimeout.IsFinished = true
				log.Println("[DealRealTimeSceneTimeout] timemap: ", TimeOutMap)
				TimeOutMapLock.Unlock()

				RtScenesListLock.RLock()
			}
		}
		RtScenesListLock.RUnlock()
	}
}

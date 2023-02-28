package handlers

import "time"

func DealRealTimeSceneTimeout() {
	for ; ; time.Sleep(time.Second) {
		RtScenesListLock.RLock()
		for i, rtScene := range RtProcessingScenesList {
			if time.Now().After(rtScene.ExpireTime) {
				RtScenesListLock.RUnlock()
				RtScenesListLock.Lock()
				// remove timeout scene
				RtProcessingScenesList = append(RtProcessingScenesList[:i], RtProcessingScenesList[i+1:]...)
				RtScenesListLock.Unlock()
				RtScenesListLock.RLock()
			}
		}
		RtScenesListLock.RUnlock()
	}
}

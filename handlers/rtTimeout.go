package handlers

import (
	"log"
	"time"
)

func DealRealTimeSceneTimeout() {
	for ; ; time.Sleep(time.Second) {
		UsersLock.RLock()
		for name, user := range Users {
			user.ExpireTimeLock.RLock()
			isTimeout := time.Now().After(user.ExpireTime)
			user.ExpireTimeLock.RUnlock()
			if isTimeout {
				UsersLock.RUnlock()
				UsersLock.Lock()
				delete(Users, name)
				UsersLock.Unlock()
				UsersLock.RLock()
				log.Println("[DealRealTimeSceneTimeout] Delete ", name)
				log.Println("[DealRealTimeSceneTimeout] This is Users: ", Users)
			}
		}
		UsersLock.RUnlock()
	}
}

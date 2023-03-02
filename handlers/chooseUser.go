package handlers

import (
	"math"
	"time"
)

func scoreUser(user *User) float64 {
	score := 0.0
	// less time left, higher priority
	now := time.Now()
	left := user.ExpireTime.Sub(now)
	score += left.Seconds()
	// more relocalisecnt, less priority
	user.RelocaliseCntLock.RLock()
	cnt := user.RelocaliseCnt
	user.RelocaliseCntLock.RUnlock()
	score -= float64(cnt * 10)
	return score
}
func findMaxScore() *User {
	maxScore := math.Inf(-1)
	var user *User
	for _, u := range Users {
		score := scoreUser(u)
		if score > maxScore {
			maxScore = score
			user = u
		}
	}
	return user
}

func chooseUser() *User {
	return findMaxScore()
}

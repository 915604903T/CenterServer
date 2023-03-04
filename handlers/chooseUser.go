package handlers

import (
	"log"
	"time"

	"github.com/emirpasic/gods/maps/treemap"
)

type UserScore struct {
	User  *User
	Score float64
}

func byID(a, b interface{}) int {

	// Type assertion, program will panic if this is not respected
	c1 := a.(UserScore)
	c2 := b.(UserScore)

	switch {
	case c1.Score > c2.Score:
		return -1
	case c1.Score < c2.Score:
		return 1
	default:
		return 0
	}
}

func scoreUser(user *User) float64 {
	score := 0.0
	// less time left, higher priority
	now := time.Now()
	left := user.ExpireTime.Sub(now)
	score += 1 / left.Seconds() * 1e4
	log.Println("[scoreUser] user:", user.Name, " time score: ", score)
	// more relocalisecnt, less priority
	user.RelocaliseCntLock.RLock()
	cnt := user.RelocaliseCnt
	user.RelocaliseCntLock.RUnlock()
	score -= float64(cnt * 5)
	log.Println("[scoreUser] user:", user.Name, "cnt score: ", score)
	return score
}
func findMaxScore() *treemap.Map {
	userScore := treemap.NewWith(byID)

	for _, u := range Users {
		score := scoreUser(u)
		userScore.Put(UserScore{u, score}, true)
	}
	return userScore
}

func chooseUser() *treemap.Map {
	return findMaxScore()
}

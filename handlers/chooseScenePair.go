package handlers

import (
	"log"
	"math/rand"
	"time"
)

func genRandomCandidates() (string, string) {
	var index1, index2 int
	length := len(ProcessingScenesList)
	index1 = rand.Intn(length)
	for index2 == index1 {
		index2 = rand.Intn(length)
	}
	// Always keep index1<index2
	if index2 < index1 {
		index1, index2 = index2, index1
	}
	return ProcessingScenesList[index1], ProcessingScenesList[index2]
}

func scoreCandidate(c1, c2 string) float64 {
	sceneUnionLock.Lock()
	p1, p2 := sceneUnion.find(c1), sceneUnion.find(c2)
	sceneUnionLock.Unlock()
	// if they are already in the same mesh component, do not choose them
	if p1 == p2 {
		return -100
	}
	score := 0.0
	if FailedSceneList[c1] != nil {
		if count, ok := FailedSceneList[c1][c2]; ok {
			if count > 3 { // if failed over 3 times, do not choose these pair
				return -400.0
			}
			score -= float64(count * 100)
		}
	}
	sceneLengthLock.RLock()
	length1 := sceneLength[c1]
	length2 := sceneLength[c2]
	sceneLengthLock.RUnlock()
	score += float64(length1+length2) / 10.0
	return score
}

//choose candidate from normal processing list
func genWeightedCandidate() (string, string) {
	length := len(ProcessingScenesList)
	index1 := rand.Intn(length)
	index2 := rand.Intn(length)
	for index2 == index1 {
		index2 = rand.Intn(length)
	}
	if index2 < index1 {
		index1, index2 = index2, index1
	}
	return ProcessingScenesList[index1], ProcessingScenesList[index2]
}
func genWeightedCandidates() (string, string) {
	maxf := 0.0
	maxIndex := [2]string{"", ""}
	for i := 0; i < candidateNum; i++ {
		c1, c2 := genWeightedCandidate()
		score := scoreCandidate(c1, c2)
		if score == 2.0 { // has never failed before
			return c1, c2
		}
		// if candidate failed over 3 times, it will never go into the following branch
		if maxf < score {
			maxf = score
			maxIndex[0] = c1
			maxIndex[1] = c2
		}
	}
	return maxIndex[0], maxIndex[1]
}

//choose candidate from real time processing list
func scoreRTCandidate(c1, c2 string) float64 {
	sceneUnionLock.Lock()
	p1, p2 := sceneUnion.find(c1), sceneUnion.find(c2)
	sceneUnionLock.Unlock()
	// if they are already in the same mesh component, do not choose them
	if p1 == p2 {
		return -100
	}
	score := 0.0

	if FailedSceneList[c1] != nil {
		if count, ok := FailedSceneList[c1][c2]; ok {
			if count > 3 { // if failed over 3 times, do not choose these pair
				return -400.0
			}
			score -= float64(count * 100)
		}
	}
	log.Println("[scoreRTCandidate] score failedList: ", score)

	// Prioritize those with less time
	now := time.Now()
	TimeOutMapLock.RLock()
	expireTime1 := TimeOutMap[c1]
	expireTime2 := TimeOutMap[c2]
	TimeOutMapLock.RUnlock()
	left1 := expireTime1.ExpireTime.Sub(now)
	left2 := expireTime2.ExpireTime.Sub(now)
	score += 1 / left1.Seconds() * 1e4
	score += 1 / left2.Seconds() * 1e4
	log.Println("[scoreRTCandidate] score leftTime: ", score)

	// longer scene has higher priority
	sceneLengthLock.RLock()
	length1 := sceneLength[c1]
	length2 := sceneLength[c2]
	sceneLengthLock.RUnlock()
	score += float64(length1+length2) / 10.0
	log.Println("[scoreRTCandidate] score length: ", score)
	return score
}
func genRealTimeCandidate() (string, string) {
	length := len(RtProcessingScenesList)
	index1 := rand.Intn(length)
	index2 := rand.Intn(length)
	for index2 == index1 {
		index2 = rand.Intn(length)
	}
	if index2 < index1 {
		index1, index2 = index2, index1
	}
	return RtProcessingScenesList[index1].Name, RtProcessingScenesList[index2].Name
}
func genRealTimeCandidates() (string, string) {
	maxf := 0.0
	maxIndex := [2]string{"", ""}
	for i := 0; i < candidateNum; i++ {
		c1, c2 := genRealTimeCandidate()
		score := scoreRTCandidate(c1, c2)
		if score == 2.0 { // has never failed before
			return c1, c2
		}
		// if candidate failed over 3 times, it will never go into the following branch
		if maxf < score {
			maxf = score
			maxIndex[0] = c1
			maxIndex[1] = c2
		}
	}
	return maxIndex[0], maxIndex[1]
}

func genCandidates(method string) (string, string) {
	switch method {
	case "random":
		return genRandomCandidates()
	case "weighted":
		return genWeightedCandidates()
	case "realTime":
		return genRealTimeCandidates()
	default:
		log.Println("invalid generate candidate ")
		return "", ""
	}
}

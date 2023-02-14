package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
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
	if FailedSceneList[c1] != nil {
		if count, ok := FailedSceneList[c1][c2]; ok {
			return 1.0 / float64(count)
		}
	}
	return 2.0
}

func genWeightedCandidate() (string, string) {
	length := len(ProcessingScenesList)
	if length >= 2 {
		index1 := rand.Intn(length)
		index2 := rand.Intn(length)
		for index2 == index1 {
			index2 = rand.Intn(length)
		}
		if index2 < index1 {
			index1, index2 = index2, index1
		}
		return ProcessingScenesList[index1], ProcessingScenesList[index2]
	} else {
		lengthSucc := len(SucceedSceneList)
		index2 := rand.Intn(lengthSucc)
		return ProcessingScenesList[0], SucceedSceneList[index2]
	}
}

func genWeightedCandidates() (string, string) {
	maxf := 0.0
	maxIndex := [2]string{}
	for i := 0; i < candidateNum; i++ {
		c1, c2 := genWeightedCandidate()
		score := scoreCandidate(c1, c2)
		if score == 2.0 { // has never failed before
			return c1, c2
		}
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
	default:
		log.Println("invalid generate candidate ")
		return "", ""
	}
}

func RunReloclise() {
	for {
		log.Println("[runRelocalise] preparedScene: ", ProcessingScenesList)
		if len(ProcessingScenesList) > 0 && len(ProcessingScenesList)+len(SucceedSceneList) >= 2 {
			// randomly choose scene to relocalise
			ScenesListLock.RLock()
			name1, name2 := genCandidates("weighted")
			ScenesListLock.RUnlock()

			clientNO1 := ClientScenes[name1]
			clientNO2 := ClientScenes[name2]
			score1, score2 := scoreClient(clientNO1), scoreClient(clientNO2)
			if score1 < score2 {
				name1, name2 = name2, name1
				clientNO1, clientNO2 = clientNO2, clientNO1
			}

			// do delete candidates scene from PrepareScenesList
			/*
				ScenesListLock.Lock()
				PrepareScenesList = append(PrepareScenesList[:index1], PrepareScenesList[index2+1:]...)
				ScenesListLock.Unlock()
			*/

			// send scene info to client
			clientIP := ClientAddrs[clientNO1]
			url := clientIP + "/relocalise/info"
			info := relocaliseInfo{
				name1,
				ClientAddrs[clientNO1],
				name2,
				ClientAddrs[clientNO2],
			}
			log.Println("[runRelocalise] this is relocalise info: \n", info)
			log.Println("[runRelocalise] send to url: ", url)
			infoStr, err := json.Marshal(info)
			if err != nil {
				log.Fatal(err)
				return
			}
			buf := bytes.NewBuffer([]byte(infoStr))
			request, err := http.NewRequest("GET", url, buf)
			if err != nil {
				log.Fatal(err)
				return
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Fatal(err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				resp_body, _ := ioutil.ReadAll(resp.Body)
				log.Fatal("[runRelocalise] receive error from relocalise: ", resp_body)
				return
			}
		}
		time.Sleep(time.Second * 5)
	}
}

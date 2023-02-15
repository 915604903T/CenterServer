package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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
			if count > 3 { // if failed over 3 times, do not choose these pair
				return -100.0
			}
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
	for ; ; time.Sleep(time.Second * 5) {
		log.Println("[runRelocalise] PrepareScene: ", ProcessingScenesList)
		if len(ProcessingScenesList) > 0 && len(ProcessingScenesList)+len(SucceedSceneList) >= 2 {
			// randomly choose scene to relocalise
			ScenesListLock.RLock()
			name1, name2 := genCandidates("weighted")
			ScenesListLock.RUnlock()
			fmt.Println("runReloclise!!!!!!!! scene pair: ", name1, name2)
			// cannot choose a suitable candidate, then continue
			if name1 == "" && name2 == "" {
				continue
			}
			RunningScenePairsLock.RLock()
			fmt.Println("runReloclise!!!!!!!! RunningScenePairs: ", RunningScenePairs)
			_, ok := RunningScenePairs[scenePair{name1, name2}]
			if ok {
				RunningScenePairsLock.RUnlock()
				continue
			}
			_, ok = RunningScenePairs[scenePair{name2, name1}]
			if ok {
				RunningScenePairsLock.RUnlock()
				continue
			}
			RunningScenePairsLock.RUnlock()

			maxScore1, maxScore2 := -200.0, -200.0
			clientNO1, clientNO2 := -1, -2
			// if no available client is ready; wait and continue to choose
			log.Println("[runRelocalise] ProcessingScene: ", ProcessingScenesList)
			ClientScenesLock.RLock()
			clients4scene1 := ClientScenes[name1]
			clients4scene2 := ClientScenes[name2]
			ClientScenesLock.RUnlock()
			fmt.Println("runReloclise!!!!!!!!scene1:", name1, " clients4scene1:", clients4scene1)
			fmt.Println("runReloclise!!!!!!!!scene2:", name2, " clients4scene2:", clients4scene2)
			maxScore1, maxScore2 = -200.0, -200.0
			//choose client 1
			for k, _ := range clients4scene1 {
				if _, ok := clients4scene2[k]; ok {
					clientNO1, clientNO2 = k, k
					break
				}
				score := scoreRelocClient(k)
				fmt.Println("runReloclise!!!!!!!!scene1:", name1, " clientNO1:", k, " score:", score)
				if score > maxScore1 {
					maxScore1 = score
					clientNO1 = k
				}
			}
			// choose client 2
			if clientNO1 != clientNO2 {
				for k, _ := range clients4scene2 {
					score := scoreRelocClient(k)
					fmt.Println("runReloclise!!!!!!!!scene2:", name2, " clientNO2:", k, " score:", score)
					if score > maxScore2 {
						maxScore2 = score
						clientNO2 = k
					}
				}
			} else {
				score := scoreRelocClient(clientNO1)
				fmt.Println(name1, name2, "runReloclise!!!!!!!!on the same server", clientNO1, " score:", score)
				if score > maxScore1 {
					maxScore1, maxScore2 = score, score
				}
				if score < 0 {
					continue
				}
			}
			fmt.Println("runReloclise!!!!!!!!maxScore1:", maxScore1, "clientNO1:", clientNO1, "maxScore2:", maxScore2, "clientNO2:", clientNO2)
			if maxScore1 < 0 && maxScore2 < 0 { // no available client can do relocalise
				continue
			}
			// always send to high score client to do relocalise
			if maxScore1 < maxScore2 {
				name1, name2 = name2, name1
				clientNO1, clientNO2 = clientNO2, clientNO1
			}

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
			RunningScenePairsLock.Lock()
			RunningScenePairs[scenePair{name1, name2}] = true
			RunningScenePairs[scenePair{name2, name1}] = true
			RunningScenePairsLock.Unlock()
		}
	}
}

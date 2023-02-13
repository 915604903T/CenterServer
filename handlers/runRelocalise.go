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

func genRandomCandidate() (int, int) {
	var index1, index2 int
	length := len(PrepareScenesList)
	index1 = rand.Intn(length)
	for index2 == index1 {
		index2 = rand.Intn(length)
	}
	// Always keep index1<index2
	if index2 < index1 {
		index1, index2 = index2, index1
	}
	return index1, index2
}

func scoreCandidate(c1, c2 string) float64 {
	if FailedSceneList[c1] != nil {
		if count, ok := FailedSceneList[c1][c2]; ok {
			return 1.0 / float64(count)
		}
	}
	return 2.0
}
func genWeightedCandidate() (int, int) {
	length := len(PrepareScenesList)
	maxf := 0.0
	maxIndex := [2]int{}
	for i := 0; i < candidateNum; i++ {
		i1 := rand.Intn(length)
		i2 := rand.Intn(length)
		for i2 == i1 {
			i2 = rand.Intn(length)
		}
		if i2 < i1 {
			i1, i2 = i2, i1
		}
		score := scoreCandidate(PrepareScenesList[i1], PrepareScenesList[i2])
		if score == 2.0 {
			return i1, i2
		}
		if maxf < score {
			maxf = score
			maxIndex[0] = i1
			maxIndex[1] = i2
		}
	}
	return maxIndex[0], maxIndex[1]
}

func genCandidate(method string) (int, int) {
	switch method {
	case "random":
		return genRandomCandidate()
	case "weighted":
		return genWeightedCandidate()
	default:
		log.Println("invalid generate candidate ")
		return -1, -1
	}
}

func RunReloclise() {
	for {
		if len(PrepareScenesList) >= 2 {
			// randomly choose scene to relocalise
			index1, index2 := genCandidate("weighted")
			name1, name2 := PrepareScenesList[index1], PrepareScenesList[index2]

			// delete candidates scene from PrepareScenesList
			ScenesListLock.Lock()
			PrepareScenesList = append(PrepareScenesList[:index1], PrepareScenesList[index2+1:]...)
			ScenesListLock.Unlock()

			// send scene info to client
			// always use the first scene client to run relocalise
			clientNO := ClientScenes[name1]
			clientIP := ClientAddrs[clientNO]
			url := clientIP + "/relocalise/info"
			info := relocaliseInfo{
				name1,
				ClientAddrs[ClientScenes[name1]],
				name2,
				ClientAddrs[ClientScenes[name2]],
			}
			log.Println("this is relocalise info: \n", info)
			log.Println("send to url: ", url)
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
				log.Fatal("receive error from relocalise: ", resp_body)
				return
			}
		}
		time.Sleep(time.Second)
	}

}

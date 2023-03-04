package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"
)

func RunReloclise() {
	var sleepDuration time.Duration

	for ; ; time.Sleep(sleepDuration) {
		// get user candidates
		UsersLock.RLock()
		usersScore := chooseUser().Keys()
		UsersLock.RUnlock()

		// choose user and get scene pair candidate
		var user *User
		var name1, name2 string
		for _, tmpUserScore := range usersScore {
			tmpUser := tmpUserScore.(UserScore).User
			log.Println("[runRelocalise] User:", tmpUser.Name, "|| ProcessingScene: ", tmpUser.ProcessingScenes)
			if len(tmpUser.ProcessingScenes) >= 2 {
				tmpUser.ProcessingScenesLock.RLock()
				name1, name2 = tmpUser.GenCandidates("weighted")
				tmpUser.ProcessingScenesLock.RUnlock()
			} else {
				continue
			}
			log.Println("[runRelocalise] User:", tmpUser.Name, "|| scene pair: ", name1, name2)
			// cannot choose a suitable candidate, then continue
			if name1 == "" && name2 == "" {
				log.Println("[runRelocalise] User:", tmpUser.Name, "invalid candidate")
				continue
			} else {
				user = tmpUser
				break
			}
		}
		if user == nil {
			log.Println("[runRelocalise] not available user's scene to do relocalise")
			sleepDuration = time.Second * time.Duration(NormalRelocaliseInterval)
			continue
		}
		// if scene pair is already running, choose another, wait shorter
		RunningScenePairsLock.RLock()
		_, ok := RunningScenePairs[scenePair{user.Name + "-" + name1, user.Name + "-" + name2}]
		if ok {
			RunningScenePairsLock.RUnlock()
			log.Println("[runRelocalise] ", user.Name+"-"+name1, user.Name+"-"+name2, "scene pair is running")
			sleepDuration = time.Second * time.Duration(ShortRelocaliseInterval)
			continue
		}
		_, ok = RunningScenePairs[scenePair{user.Name + "-" + name2, user.Name + "-" + name1}]
		if ok {
			RunningScenePairsLock.RUnlock()
			log.Println("[runRelocalise] ", user.Name+"-"+name1, user.Name+"-"+name2, "scene pair is running")
			sleepDuration = time.Second * time.Duration(ShortRelocaliseInterval)
			continue
		}
		RunningScenePairsLock.RUnlock()

		maxScore1, maxScore2 := math.Inf(-1), math.Inf(-1)
		clientNO1, clientNO2 := -1, -2
		// if no available client is ready; wait and continue to choose
		user.ClientScenesLock.RLock()
		clients4scene1 := user.ClientScenes[name1]
		clients4scene2 := user.ClientScenes[name2]
		user.ClientScenesLock.RUnlock()
		// log.Println("[runRelocalise] scene1:", name1, " clients4scene1:", clients4scene1)
		// log.Println("[runRelocalise] scene2:", name2, " clients4scene2:", clients4scene2)

		//choose client 1
		for k := range clients4scene1 {
			if _, ok := clients4scene2[k]; ok {
				clientNO1, clientNO2 = k, k
				break
			}
			score := scoreRelocClient(k)
			// fmt.Println("runReloclise!!!!!!!!scene1:", name1, " clientNO1:", k, " score:", score)
			if score > maxScore1 {
				maxScore1 = score
				clientNO1 = k
			}
		}
		// choose client 2
		if clientNO1 != clientNO2 {
			for k := range clients4scene2 {
				score := scoreRelocClient(k)
				// fmt.Println("runReloclise!!!!!!!!scene2:", name2, " clientNO2:", k, " score:", score)
				if score > maxScore2 {
					maxScore2 = score
					clientNO2 = k
				}
			}
		} else {
			score := scoreRelocClient(clientNO1)
			log.Println("[runReloclise]", name1, name2, "runReloclise!!!!!!!!on the same server", clientNO1, " score:", score)
			if score > maxScore1 {
				maxScore1, maxScore2 = score, score
			}
			if score < 0 {
				sleepDuration = time.Second * time.Duration(NormalRelocaliseInterval)
				continue
			}
		}
		log.Println("[runReloclise] maxScore1:", maxScore1, "clientNO1:", clientNO1, "maxScore2:", maxScore2, "clientNO2:", clientNO2)
		// no available client can do relocalise
		if maxScore1 < 0 && maxScore2 < 0 {
			log.Println("[runReloclise] no available client to run")
			sleepDuration = time.Second * time.Duration(NormalRelocaliseInterval)
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
		log.Println("[runRelocalise] !!!!send to url: ", url)
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

		RunningScenePairsLock.Lock()
		RunningScenePairs[scenePair{user.Name + "-" + name1, user.Name + "-" + name2}] = true
		RunningScenePairs[scenePair{user.Name + "-" + name2, user.Name + "-" + name1}] = true
		RunningScenePairsLock.Unlock()

		user.RelocaliseCntLock.Lock()
		user.RelocaliseCnt++
		user.RelocaliseCntLock.Unlock()

		log.Println("[runRelocalise] add scene pair to running list:", name1, name2)
		if resp.StatusCode != http.StatusOK {
			resp_body, _ := ioutil.ReadAll(resp.Body)
			log.Fatal("[runRelocalise] receive error from relocalise: ", resp_body)
			return
		}
		sleepDuration = time.Second * time.Duration(NormalRelocaliseInterval)
	}
}

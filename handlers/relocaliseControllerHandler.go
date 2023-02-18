package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func bfsFindPath(scene1, scene2 string) []string {
	pre := map[string]string{}
	q := []string{scene1}
	for len(q) > 0 {
		now := q[0]
		if now == scene2 {
			break
		}
		for node, _ := range sceneGraph[now] {
			pre[node] = now
			q = append(q, node)
		}
	}
	path := []string{scene2}
	idx := scene2
	for {
		if preNode, ok := pre[idx]; ok {
			path = append(path, preNode)
		} else {
			break
		}
	}
	return path
}

func findPose(scene1, scene2 string) [4][4]float64 {
	sceneGraphLock.RLock()
	path := bfsFindPath(scene1, scene2)
	poseMat := [4][4]float64{
		{1.0, 0.0, 0.0, 0.0},
		{0.0, 1.0, 0.0, 0.0},
		{0.0, 0.0, 1.0, 0.0},
		{0.0, 0.0, 0.0, 1.0},
	}
	for i := 0; i < len(path)-1; i++ {
		posei_i1 := sceneGraph[path[i]][path[i+1]]
		poseM := posei_i1.GetM()
		poseMat = Mul(poseM, poseMat)
	}
	sceneGraphLock.RUnlock()
	return poseMat
}

func doMeshRequest(scene1, scene2 string) {
	sceneUnionLock.Lock()
	pScene1, pScene2 := sceneUnion.find(scene1), sceneUnion.find(scene2)
	sceneUnionLock.Unlock()
	// if in the same union, then do not combine the mesh
	if pScene1 == pScene2 {
		return
	}
	// if in different union, find the path
	poseM := findPose(scene1, scene2)
	sceneMeshLock.RLock()
	mesh1, mesh2 := sceneMesh[pScene1], sceneMesh[pScene2]
	sceneMeshLock.RUnlock()
	mergeMeshInfo := &MergeMeshInfo{
		File1Name:  mesh1.fileName,
		File1Ip:    mesh1.client,
		File2Name:  mesh2.fileName,
		File2Ip:    mesh2.client,
		PoseMatrix: poseM,
	}
	content, err := json.Marshal(mergeMeshInfo)
	if err != nil {
		log.Println("marshal merge mesh info err: ", err)
		panic(err)
	}
	client1, client2 := mesh1.client, mesh2.client
	url := client1 + "/mesh"
	if client1 != client2 {
		score1, score2 := scoreMeshClient(ClientIpsMap[client1]), scoreMeshClient(ClientIpsMap[client2])
		if score2 > score1 {
			url = client2 + "/mesh"
		}
	}
	buf := bytes.NewBuffer([]byte(content))
	request, err := http.NewRequest("GET", url, buf)
	if err != nil {
		log.Println("generate request err: ", err)
		panic(err)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println("get mesh response err: ", err)
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		resp_body, _ := ioutil.ReadAll(resp.Body)
		log.Fatal("[doMeshRequest] receive error from mesh: ", resp_body)
		return
	}
}

func addGraphEdge(poseInfo globalPose) {
	scene1, scene2 := poseInfo.Scene1Name, poseInfo.Scene2Name
	pose21 := poseInfo.Transform
	pose12Mat := Inverse(pose21.GetM())
	pose12 := Pose{
		Matrix: pose12Mat,
		HasM:   true,
	}
	sceneGraphLock.Lock()
	sceneGraph[scene1][scene2] = pose12
	sceneGraph[scene2][scene1] = pose21
	sceneGraphLock.Unlock()

	sceneUnionLock.Lock()
	sceneUnion.union(scene1, scene2)
	sceneUnionLock.Unlock()
}

func addMeshInfo(poseInfo globalPose) {
	scene1, scene2 := poseInfo.Scene1Name, poseInfo.Scene2Name
	meshInfo := MeshInfo{
		scenes:     [2]string{scene1, scene2},
		worldScene: scene1,
		fileName:   scene1 + "-" + scene2 + ".ply",
		client:     poseInfo.Scene1Ip,
	}
	sceneMeshLock.Lock()
	sceneMesh[scene1] = meshInfo
	sceneMesh[scene2] = meshInfo
	sceneMeshLock.Unlock()
}

func MakeRelocaliseControllerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("[MakeRelocaliseControllerHandler] relocalise global pose request!!!!!!!!!!!!!!!")
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		bodyStr := string(body)
		log.Println("receive globalpose: ", bodyStr)

		if strings.Contains(bodyStr, "failed") {
			content := strings.Fields(bodyStr)
			scene1, scene2 := content[0], content[1]

			RunningScenePairsLock.Lock()
			delete(RunningScenePairs, scenePair{scene1, scene2})
			delete(RunningScenePairs, scenePair{scene2, scene1})
			// not deadlock??
			if FailedSceneList[scene1] == nil {
				FailedSceneList[scene1] = make(map[string]int)
				FailedSceneList[scene1][scene2] = 1
			} else {
				FailedSceneList[scene1][scene2]++
			}
			if FailedSceneList[scene2] == nil {
				FailedSceneList[scene2] = make(map[string]int)
				FailedSceneList[scene2][scene1] = 1
			} else {
				FailedSceneList[scene2][scene1]++
			}
			RunningScenePairsLock.Unlock()

			log.Println("[MakeRelocaliseControllerHandler] add ", scene1, scene2, "to failedList")
			w.WriteHeader(http.StatusOK)
			return
		}

		// save global pose for two scenes
		poseInfo := globalPose{}
		err := json.Unmarshal(body, &poseInfo)
		if err != nil {
			log.Fatal("[MakeRelocaliseControllerHandler] error de-serializing request body: ", body)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// add edge between two scene
		addGraphEdge(poseInfo)

		// save mesh info for two scene
		addMeshInfo(poseInfo)

		// request merge mesh if need
		scene1, scene2 := poseInfo.Scene1Name, poseInfo.Scene2Name
		doMeshRequest(scene1, scene2)

		RunningScenePairsLock.Lock()
		delete(RunningScenePairs, scenePair{poseInfo.Scene1Name, poseInfo.Scene2Name})
		delete(RunningScenePairs, scenePair{poseInfo.Scene2Name, poseInfo.Scene1Name})
		RunningScenePairsLock.Unlock()

		log.Println("[MakeRelocaliseControllerHandler] this is globalpose struct:\n", poseInfo)

		// if ip1 and ip2 are not the same one, add it to client scenes map
		ip1, ip2 := poseInfo.Scene1Ip, poseInfo.Scene2Ip
		if ip1 != ip2 {
			clientNO1 := ClientIpsMap[ip1]
			ClientScenesLock.Lock()
			ClientScenes[scene2][clientNO1] = true
			ClientScenesLock.Unlock()
		}

		w.WriteHeader(http.StatusOK)
	}
}

package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func bfsFindPath(scene1, scene2 string) []string {
	log.Println("[bfsFindPath] ", scene1, scene2)
	pre := map[string]string{}
	q := []string{scene1}
	vis := map[string]bool{scene1: true}
	for len(q) > 0 {
		now := q[0]
		q = q[1:]
		if now == scene2 {
			break
		}
		for node := range sceneGraph[now] {
			if vis[node] {
				continue
			}
			vis[node] = true
			pre[node] = now
			q = append(q, node)
		}
	}
	log.Println("[bfsFindPath] after find path", scene1, scene2)
	path := []string{scene2}
	idx := scene2
	log.Println("\n[bfsFindPath] preNodes: ", pre)
	for {
		if preNode, ok := pre[idx]; ok {
			path = append(path, preNode)
			idx = preNode
		} else {
			break
		}
	}
	log.Println("[bfsFindPath] return path", scene1, scene2)
	return path
}

func findPose(scene1, scene2 string) [4][4]float64 {
	log.Println("[findPose] begin find pose:", scene1, scene2)
	sceneGraphLock.RLock()
	log.Println("[findPose] get sceneGraph Lock:", scene1, scene2)
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
	log.Println("[findPose] after find pose", scene1, scene2)
	return poseMat
}

func doMeshRequest(scene1, scene2 string) {
	log.Println("[doMeshRequest] merge", scene1, scene2, " meshes")
	var size1, size2 int
	var pScene1, pScene2 string
	var mesh1, mesh2 *MeshInfo
	// check whether the required meshes are occupied
	for ; ; time.Sleep(time.Second) {
		sceneUnionLock.Lock()
		pScene1, pScene2 = sceneUnion.find(scene1), sceneUnion.find(scene2)
		size1, size2 = sceneUnion.Size[pScene1], sceneUnion.Size[pScene2]
		sceneUnionLock.Unlock()

		sceneMeshLock.RLock()
		mesh1, mesh2 = sceneMesh[pScene1], sceneMesh[pScene2]
		sceneMeshLock.RUnlock()
		RunningMeshesLock.RLock()
		occupied := RunningMeshes[mesh1] || RunningMeshes[mesh2]
		RunningMeshesLock.RUnlock()
		// if both scenes are not running, do following things
		log.Println("[doMeshRequest] verify if mesh is occupied: ", mesh1.FileName, mesh2.FileName, " status:", occupied)
		if !occupied {
			break
		}
	}
	// if in the same union then do not combine the mesh
	log.Println("[doMeshRequest] scene1:", scene1, "p1:", pScene1, " scene2:", scene2, "p2:", pScene2)
	if pScene1 == pScene2 {
		log.Println("[doMeshRequest] parent equal in the same union: ", scene1, scene2, " then return")
		return
	}
	log.Println("[doMeshRequest] this is sceneUnion size1: ", size1, "size2: ", size2)
	// if there is just one scene, it is the relocalise result
	if size1 == 1 && size2 == 1 {
		sceneUnionLock.Lock()
		sceneUnion.union(scene1, scene2)
		log.Println("[doMeshRequest] in the same size union: ", sceneUnion)
		sceneUnionLock.Unlock()
		return
	}
	// add to running mesh list
	RunningMeshesLock.Lock()
	RunningMeshes[mesh1] = true
	RunningMeshes[mesh2] = true
	RunningMeshesLock.Unlock()
	log.Println("[doMeshRequest] add to mesh to running meshes: ", mesh1.WorldScene, mesh2.WorldScene)
	// if in different union, find the path
	poseM := findPose(mesh1.WorldScene, mesh2.WorldScene)
	mergeMeshInfo := &MergeMeshInfo{
		Mesh1:      *mesh1,
		Mesh2:      *mesh2,
		PoseMatrix: poseM,
	}
	log.Println("[doMeshRequest] !!!!!!!!!!!!!mesh 1111111111 info: ", mesh1)
	log.Println("[doMeshRequest] !!!!!!!!!!!!!mesh  2222222222222 info: ", mesh2)
	content, err := json.Marshal(mergeMeshInfo)
	if err != nil {
		log.Println("[doMeshRequest] marshal merge mesh info err: ", err)
		panic(err)
	}
	log.Println("[doMeshRequest] this is send content: ", string(content))
	client1, client2 := mesh1.Client, mesh2.Client
	url := client1 + "/mesh"
	if client1 != client2 {
		score1, score2 := scoreMeshClient(ClientIpsMap[client1]), scoreMeshClient(ClientIpsMap[client2])
		if score2 > score1 {
			url = client2 + "/mesh"
		}
	}
	log.Println("[doMeshRequest] request merge mesh url: ", url)
	buf := bytes.NewBuffer([]byte(content))
	request, err := http.NewRequest("GET", url, buf)
	if err != nil {
		log.Println("[doMeshRequest] generate request err: ", err)
		panic(err)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println("[doMeshRequest] get mesh response err: ", err)
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		resp_body, _ := ioutil.ReadAll(resp.Body)
		log.Fatal("[doMeshRequest] receive error from mesh: ", resp_body)
		return
	}
	// add to the same union
	sceneUnionLock.Lock()
	sceneUnion.union(scene1, scene2)
	log.Println("[doMeshRequest] after send union: ", sceneUnion)
	sceneUnionLock.Unlock()
}

func addGraphEdge(poseInfo globalPose) {
	scene1, scene2 := poseInfo.Scene1Name, poseInfo.Scene2Name
	pose12 := poseInfo.Transform
	pose21Mat := Inverse(pose12.GetM())
	pose21 := NewPoseMatrix(pose21Mat)
	log.Println("[addGraphEdge] add scene graph edge: ", scene1, scene2)
	sceneGraphLock.Lock()
	sceneGraph[scene1][scene2] = pose12
	sceneGraph[scene2][scene1] = pose21
	sceneGraphLock.Unlock()
	log.Println("[addGraphEdge] after add scene graph edge: ", scene1, scene2)
}

func addMeshInfo(poseInfo globalPose) {
	scene1, scene2 := poseInfo.Scene1Name, poseInfo.Scene2Name
	scenes := make(map[string]bool)
	scenes[scene1] = true
	scenes[scene2] = true
	// scene1 is the default world scene
	meshInfo := MeshInfo{
		Scenes:     scenes,
		WorldScene: scene1,
		FileName:   scene1 + "-" + scene2 + ".ply",
		Client:     poseInfo.Scene1Ip,
	}
	log.Println("!!!!!!!!!!!!!add mesh info: ", meshInfo)
	sceneMeshLock.Lock()
	if _, ok := sceneMesh[scene1]; !ok {
		sceneMesh[scene1] = &meshInfo
	}
	if _, ok := sceneMesh[scene2]; !ok {
		sceneMesh[scene2] = &meshInfo
	}
	sceneMeshLock.Unlock()

}

func MakeRelocaliseControllerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("[MakeRelocaliseControllerHandler] relocalise global pose request!!!!!!!!!!!!!!!")
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		bodyStr := string(body)
		log.Println("[MakeRelocaliseControllerHandler] receive globalpose:\n\n", bodyStr)

		if bodyStr == "not successfully run" {
			log.Println("[MakeRelocaliseControllerHandler] error run relocalise program")
			w.WriteHeader(http.StatusOK)
			return
		}

		RelocaliseCntLock.Lock()
		RelocaliseCnt++
		log.Println("[runRelocalise] Relocalise Times: ", RelocaliseCnt)
		RelocaliseCntLock.Unlock()

		// do not relocalise
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

			log.Println("[MakeRelocaliseControllerHandler] add ", scene1, scene2, "to failedList and remove from running list")
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
		log.Println("[MakeRelocaliseControllerHandler] remove from running list: ", poseInfo.Scene1Name, poseInfo.Scene2Name)
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

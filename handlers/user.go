package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type User struct {
	Name              string
	ExpireTime        time.Time
	ExpireTimeLock    sync.RWMutex
	RelocaliseCnt     int
	RelocaliseCntLock sync.RWMutex
	// Scene prepare to relocalise
	ProcessingScenes     []string
	ProcessingScenesLock sync.RWMutex
	// sceneUnion shows which scenes are in the same scenario
	SceneUnion     *UnionSet
	SceneUnionLock sync.Mutex
	// save the pose between different scenes
	SceneGraph     map[string]map[string]Pose
	SceneGraphLock sync.RWMutex
	// find the corresponding mesh of a specific scene
	SceneMesh     map[string]*MeshInfo
	SceneMeshLock sync.RWMutex
	// record the length of the video (picture count)
	SceneLength     map[string]int
	SceneLengthLock sync.RWMutex
	// give info about which clients a specific scene locates (one scene can locate multiple client)
	ClientScenes     map[string]map[int]bool
	ClientScenesLock sync.RWMutex
}

func NewUser(name string) *User {
	return &User{
		Name:             name,
		RelocaliseCnt:    0,
		ProcessingScenes: []string{},
		SceneUnion:       NewUnionSet(),
		SceneGraph:       map[string]map[string]Pose{},
		SceneMesh:        map[string]*MeshInfo{},
		SceneLength:      map[string]int{},
		ClientScenes:     map[string]map[int]bool{},
	}
}

// relocalise controller helper: add graph info, add mesh info, add
func (user *User) bfsFindPath(scene1, scene2 string) []string {
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
		for node := range user.SceneGraph[now] {
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

func (user *User) findPose(scene1, scene2 string) [4][4]float64 {
	log.Println("[findPose] begin find pose:", scene1, scene2)
	user.SceneGraphLock.RLock()
	log.Println("[findPose] get sceneGraph Lock:", scene1, scene2)
	path := user.bfsFindPath(scene1, scene2)
	poseMat := [4][4]float64{
		{1.0, 0.0, 0.0, 0.0},
		{0.0, 1.0, 0.0, 0.0},
		{0.0, 0.0, 1.0, 0.0},
		{0.0, 0.0, 0.0, 1.0},
	}
	for i := 0; i < len(path)-1; i++ {
		posei_i1 := user.SceneGraph[path[i]][path[i+1]]
		poseM := posei_i1.GetM()
		poseMat = Mul(poseM, poseMat)
	}
	user.SceneGraphLock.RUnlock()
	log.Println("[findPose] after find pose", scene1, scene2)
	return poseMat
}

func (user *User) DoMeshRequest(scene1, scene2 string) {
	log.Println("[doMeshRequest] merge", scene1, scene2, " meshes")
	var size1, size2 int
	var pScene1, pScene2 string
	var mesh1, mesh2 *MeshInfo
	// check whether the required meshes are occupied
	for ; ; time.Sleep(time.Second * 2) {
		user.SceneUnionLock.Lock()
		pScene1, pScene2 = user.SceneUnion.find(scene1), user.SceneUnion.find(scene2)
		size1, size2 = user.SceneUnion.Size[pScene1], user.SceneUnion.Size[pScene2]
		user.SceneUnionLock.Unlock()

		user.SceneMeshLock.RLock()
		mesh1, mesh2 = user.SceneMesh[pScene1], user.SceneMesh[pScene2]
		user.SceneMeshLock.RUnlock()
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
		user.SceneUnionLock.Lock()
		user.SceneUnion.union(scene1, scene2)
		log.Println("[doMeshRequest] in the same size union: ", user.SceneUnion)
		user.SceneUnionLock.Unlock()
		return
	}
	// add to running mesh list
	RunningMeshesLock.Lock()
	RunningMeshes[mesh1] = true
	RunningMeshes[mesh2] = true
	RunningMeshesLock.Unlock()
	log.Println("[doMeshRequest] add to mesh to running meshes: ", mesh1.WorldScene, mesh2.WorldScene)
	// if in different union, find the path
	poseM := user.findPose(mesh1.WorldScene, mesh2.WorldScene)
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
	user.SceneUnionLock.Lock()
	user.SceneUnion.union(scene1, scene2)
	log.Println("[doMeshRequest] after send union: ", user.SceneUnion)
	user.SceneUnionLock.Unlock()
}

func (user *User) AddGraphEdge(poseInfo globalPose) {
	scene1, scene2 := poseInfo.Scene1Name, poseInfo.Scene2Name
	pose12 := poseInfo.Transform
	pose21Mat := Inverse(pose12.GetM())
	pose21 := NewPoseMatrix(pose21Mat)
	log.Println("[addGraphEdge] add scene graph edge: ", scene1, scene2)
	user.SceneGraphLock.Lock()
	log.Println("[addGraphEdge] get scenegraph lock!!")
	user.SceneGraph[scene1][scene2] = pose12
	log.Println("[addGraphEdge] 1111111111")
	user.SceneGraph[scene2][scene1] = pose21
	user.SceneGraphLock.Unlock()
	log.Println("[addGraphEdge] after add scene graph edge: ", scene1, scene2)
}

func (user *User) AddMeshInfo(poseInfo globalPose) {
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
	user.SceneMeshLock.Lock()
	if _, ok := user.SceneMesh[scene1]; !ok {
		user.SceneMesh[scene1] = &meshInfo
	}
	if _, ok := user.SceneMesh[scene2]; !ok {
		user.SceneMesh[scene2] = &meshInfo
	}
	user.SceneMeshLock.Unlock()

}

// choose scene pair helper
func (user *User) genRandomCandidates() (string, string) {
	var index1, index2 int
	length := len(user.ProcessingScenes)
	index1 = rand.Intn(length)
	for index2 == index1 {
		index2 = rand.Intn(length)
	}
	// Always keep index1<index2
	if index2 < index1 {
		index1, index2 = index2, index1
	}
	return user.ProcessingScenes[index1], user.ProcessingScenes[index2]
}

func (user *User) scoreCandidate(c1, c2 string) float64 {
	user.SceneUnionLock.Lock()
	p1, p2 := user.SceneUnion.find(c1), user.SceneUnion.find(c2)
	user.SceneUnionLock.Unlock()
	// if they are already in the same mesh component, do not choose them
	if p1 == p2 {
		return math.Inf(-1)
	}
	score := 0.0
	// if the pair already failed, give them penalty
	if FailedSceneList[c1] != nil {
		if count, ok := FailedSceneList[c1][c2]; ok {
			if count > 3 { // if failed over 3 times, do not choose these pair
				return -400.0
			}
			score -= float64(count * 100)
		}
	}
	// log.Println("[scoreCandidate] after evaluate failedlist score: ", score)
	// let longer scene first, because it will probabaly contain more scenes
	user.SceneLengthLock.RLock()
	length1 := user.SceneLength[c1]
	length2 := user.SceneLength[c2]
	user.SceneLengthLock.RUnlock()
	score += float64(length1+length2) / 10.0
	// log.Println("[scoreCandidate] after evaluate sceneLength score: ", score)
	return score
}

//choose candidate from normal processing list
func (user *User) genWeightedCandidate() (string, string) {
	length := len(user.ProcessingScenes)
	index1 := rand.Intn(length)
	index2 := rand.Intn(length)
	for index2 == index1 {
		index2 = rand.Intn(length)
	}
	if index2 < index1 {
		index1, index2 = index2, index1
	}
	return user.ProcessingScenes[index1], user.ProcessingScenes[index2]
}

func (user *User) genWeightedCandidates() (string, string) {
	maxf := 0.0
	maxIndex := [2]string{"", ""}
	for i := 0; i < candidateNum; i++ {
		c1, c2 := user.genWeightedCandidate()
		score := user.scoreCandidate(c1, c2)
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

func (user *User) GenCandidates(method string) (string, string) {
	switch method {
	case "random":
		return user.genRandomCandidates()
	case "weighted":
		return user.genWeightedCandidates()
	default:
		log.Println("invalid generate candidate ")
		return "", ""
	}
}

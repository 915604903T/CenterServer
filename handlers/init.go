package handlers

import (
	"math/rand"
	"time"
)

func init() {
	Client1Ip = "210.28.134.72"
	Client1Port = "23334"
	Client2Ip = "210.28.134.72"
	Client2Port = "23335"

	ClientAddrs = []string{}
	ClientIpsMap = make(map[string]int)
	count := 0
	ClientAddrs = append(ClientAddrs, "http://"+Client1Ip+":"+Client1Port)
	ClientIpsMap["http://"+Client1Ip+":"+Client1Port] = count
	count++
	ClientAddrs = append(ClientAddrs, "http://"+Client2Ip+":"+Client2Port)
	ClientIpsMap["http://"+Client2Ip+":"+Client2Port] = count
	count++

	Users = make(map[string]*User)
	Users[DefaultUserName] = NewUser(DefaultUserName)
	SceneUserMap = make(map[string]string)
	RunningScenePairs = make(map[scenePair]bool)
	FailedSceneList = make(map[string]map[string]int)
	RunningMeshes = make(map[*MeshInfo]bool)

	/*
		ClientScenes = make(map[string]map[int]bool) // save where the scene locate
		sceneLength = make(map[string]int)
		ProcessingScenesList = []string{}
		ProcessingScenesIndex = make(map[string]int)
		RtProcessingScenesList = []RtScene{}
		TimeOutMap = make(map[string]*SceneTimeout)
		sceneUnion = NewUnionSet()
		sceneGraph = make(map[string]map[string]Pose)
		sceneMesh = make(map[string]*MeshInfo)
	*/

	rand.Seed(time.Now().Unix())
}

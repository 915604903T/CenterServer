package handlers

import "sync"

var Client1Ip, Client2Ip string
var Client1Port, Client2Port string

var ClientAddrs []string
var ClientScenes map[string]int
var globalPoses map[[2]string][2]pose

var PrepareScenesList []string
var ProcessingSceneList []string
var SuccessScenesList map[string]bool

var ScenesListLock sync.RWMutex
var globalPoseLock sync.RWMutex

var clientCnt int = 2
var nowClient int32 = -1
var sceneIndex int = 0

type pose [4][2]float64

type globalPose struct {
	Scene1Name string `json:"scene1name"`
	Scene1Pose pose   `json:"scene1pose"`
	Scene2Name string `json:"scene2name"`
	Scene2Pose pose   `json:"scene2pose"`
}

type relocaliseInfo struct {
	Scene1Name string `json:"scene1name"`
	Scene1IP   string `json:"scene1ip"`
	Scene2Name string `json:"scene2name"`
	Scene2IP   string `json:"scene2ip"`
}

func init() {
	Client1Ip = "127.0.0.1"
	Client1Port = "23334"
	Client2Ip = "127.0.0.1"
	Client2Port = "23335"

	ClientAddrs = []string{}
	ClientAddrs = append(ClientAddrs, "http://"+Client1Ip+":"+Client1Port)
	ClientAddrs = append(ClientAddrs, "http://"+Client2Ip+":"+Client2Port)

	ClientScenes = make(map[string]int)

	PrepareScenesList = []string{}
	ProcessingSceneList = []string{}
	SuccessScenesList = make(map[string]bool)
	globalPoses = make(map[[2]string][2]pose)
}

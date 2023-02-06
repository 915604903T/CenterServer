package handlers

import "sync"

var Client1Ip, Client2Ip string
var Client1Port, Client2Port string

var ClientAddrs []string
var ClientScenes map[string]int
var globalPose map[[2]string][4][4]float64

var PrepareScenesList []string
var ProcessingSceneList []string
var SuccessScenesList map[string]bool

var ScenesListLock sync.RWMutex

var clientCnt int = 2
var nowClient int32 = -1
var sceneIndex int = 0

func init() {
	Client1Ip = "127.0.0.1"
	Client1Port = "23334"
	Client2Ip = "127.0.0.1"
	Client2Port = "23335"

	ClientAddrs = []string{}
	ClientAddrs = append(ClientAddrs, Client1Ip+":"+Client1Port)
	ClientAddrs = append(ClientAddrs, Client2Ip+":"+Client2Port)

	ClientScenes = make(map[string]int)

	PrepareScenesList = []string{}
	ProcessingSceneList = []string{}
	SuccessScenesList = make(map[string]bool)
	globalPose = make(map[[2]string][4][4]float64)
}

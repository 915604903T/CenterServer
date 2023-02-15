package handlers

import (
	"math/rand"
	"sync"
	"time"
)

const clientCnt int = 2

var Client1Ip, Client2Ip string
var Client1Port, Client2Port string

var ClientAddrs []string
var ClientIpsMap map[string]int
var ClientResourceStats [clientCnt]ResourceInfo

var ClientScenes map[string]map[int]bool

var globalPoses map[scenePair][2]pose

var ProcessingScenesList []string
var ProcessingScenesIndex map[string]int

type scenePair struct {
	scene1, scene2 string
}

var RunningScenePairs map[scenePair]bool

var SucceedSceneList []string

var FailedSceneList map[string]map[string]int

var ScenesListLock sync.RWMutex
var globalPoseLock sync.RWMutex
var resourceInfoLock sync.RWMutex
var ClientScenesLock sync.RWMutex
var RunningScenePairsLock sync.RWMutex

var nowClient int32 = -1

const candidateNum int = 10

type pose [4][2]float64

type globalPose struct {
	Scene1Name string `json:"scene1name"`
	Scene1Ip   string `json:"scene1ip"`
	Scene1Pose pose   `json:"scene1pose"`
	Scene2Name string `json:"scene2name"`
	Scene2Ip   string `json:"scene2ip"`
	Scene2Pose pose   `json:"scene2pose"`
}

type relocaliseInfo struct {
	Scene1Name string `json:"scene1name"`
	Scene1IP   string `json:"scene1ip"`
	Scene2Name string `json:"scene2name"`
	Scene2IP   string `json:"scene2ip"`
}

type ResourceInfo struct {
	GPUMemoryFree uint64    `json:"gpumemoryfree"`
	MemoryFree    uint64    `json:"memoryfree"`
	CpuUsage      []float64 `json:"cpuusage"`
}

func init() {
	Client1Ip = "127.0.0.1"
	Client1Port = "23334"
	Client2Ip = "127.0.0.1"
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

	ClientScenes = make(map[string]map[int]bool) // save where the scene locate

	RunningScenePairs = make(map[scenePair]bool)

	ProcessingScenesList = []string{}
	ProcessingScenesIndex = make(map[string]int)

	SucceedSceneList = []string{}

	FailedSceneList = make(map[string]map[string]int)
	globalPoses = make(map[scenePair][2]pose)

	rand.Seed(time.Now().Unix())

}

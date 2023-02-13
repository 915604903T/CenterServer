package handlers

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

const clientCnt int = 2

var Client1Ip, Client2Ip string
var Client1Port, Client2Port string

var ClientAddrs []string
var ClientResourceStats [clientCnt]ResourceInfo

var ClientScenes map[string]int
var globalPoses map[[2]string][2]pose

var PrepareScenesList []string
var SuccessScenesList map[string]bool
var FailedSceneList map[string]map[string]int

var ScenesListLock sync.RWMutex
var globalPoseLock sync.RWMutex

var resourceInfoLock sync.RWMutex

var nowClient int32 = -1
var sceneIndex int = 0

const candidateNum int = 10

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
	ClientAddrs = append(ClientAddrs, "http://"+Client1Ip+":"+Client1Port)
	ClientAddrs = append(ClientAddrs, "http://"+Client2Ip+":"+Client2Port)

	ClientScenes = make(map[string]int)

	PrepareScenesList = []string{}
	SuccessScenesList = make(map[string]bool)
	FailedSceneList = make(map[string]map[string]int)
	globalPoses = make(map[[2]string][2]pose)

	rand.Seed(time.Now().Unix())

	v, _ := mem.VirtualMemory()
	log.Println("v: ", v)
}

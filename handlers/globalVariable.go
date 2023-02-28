package handlers

import (
	"sync"
	"time"
)

const clientCnt int = 2

// client network info
var Client1Ip, Client2Ip string
var Client1Port, Client2Port string
var ClientAddrs []string
var ClientIpsMap map[string]int

// client resource info
var ClientResourceStats [clientCnt]ResourceInfo
var resourceInfoLock sync.RWMutex

// give info about which clients a specific scene locates (one scene can locate multiple client)
var ClientScenes map[string]map[int]bool
var ClientScenesLock sync.RWMutex

// record the length of the video (picture count)
var sceneLength map[string]int
var sceneLengthLock sync.RWMutex

// record failed scene pair in order to generate more possible candidate
var FailedSceneList map[string]map[string]int

// Scene prepare to relocalise
var ProcessingScenesList []string
var ProcessingScenesIndex map[string]int // help ProcessingScenesList to delete scene
var ScenesListLock sync.RWMutex

// Real time Scene prepare to relocalise
var RtProcessingScenesList []RtScene
var RtScenesListLock sync.RWMutex

// Real time scene timeout time
var TimeOutMap map[string]time.Time
var TimeOutMapLock sync.RWMutex

// prevent running the same scene pair at the same time
var RunningScenePairs map[scenePair]bool
var RunningScenePairsLock sync.RWMutex

// sceneUnion shows which scenes are in the same scenario
var sceneUnion UnionSet
var sceneUnionLock sync.Mutex

// save the pose between different scenes
var sceneGraph map[string]map[string]Pose // does not have circle in graph
var sceneGraphLock sync.RWMutex

// find the corresponding mesh of a specific scene
var sceneMesh map[string]*MeshInfo
var sceneMeshLock sync.RWMutex

var RunningMeshes map[*MeshInfo]bool
var RunningMeshesLock sync.RWMutex

// for random choose client usage
var nowClient int32 = -1

// generate scene pair candidates number
const candidateNum int = 5

var RelocaliseCntLock sync.Mutex
var RelocaliseCnt = 0

package handlers

import "time"

type scenePair struct {
	scene1, scene2 string
}

// always regard the world scene as scene1
type globalPose struct {
	Scene1Name string `json:"scene1name"`
	Scene1Ip   string `json:"scene1ip"`
	Scene2Name string `json:"scene2name"`
	Scene2Ip   string `json:"scene2ip"`
	Transform  Pose   `json:"transform"`
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

type MeshInfo struct {
	Scenes     map[string]bool `json:"scenes"`
	WorldScene string          `json:"worldscene"`
	FileName   string          `json:"filename"`
	Client     string          `json:"client"`
}

type MergeMeshInfo struct {
	Mesh1      MeshInfo      `json:"mesh1"`
	Mesh2      MeshInfo      `json:"mesh2"`
	PoseMatrix [4][4]float64 `json:"posematrix"` //scene2 to scene1 transform matrix
}

type RtScene struct {
	Name       string
	ExpireTime time.Time
}

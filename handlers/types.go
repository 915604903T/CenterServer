package handlers

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
	scenes     [2]string
	worldScene string
	fileName   string
	client     string
}

type MergeMeshInfo struct {
	File1Name  string        `json:"file1name"`
	File1Ip    string        `json:"file1ip"`
	File2Name  string        `json:"file2name"`
	File2Ip    string        `json:"file2ip"`
	PoseMatrix [4][4]float64 `json:"posematrix"`
}

package handlers

import (
	"log"
	"sync/atomic"
)

func chooseSequentialClient() int {
	atomic.AddInt32(&nowClient, 1)
	return int(atomic.LoadInt32(&nowClient)) % clientCnt
}

func scoreMeshClient(id int) float64 {
	score := 0.0
	resourceInfoLock.RLock()
	resourceInfo := ClientResourceStats[id]
	resourceInfoLock.RUnlock()
	score += float64(resourceInfo.MemoryFree) / 1e9
	score += float64(resourceInfo.GPUMemoryFree) / 1e7
	for _, cpu := range resourceInfo.CpuUsage {
		score += 1 - cpu
	}
	return score
}

func scoreRelocClient(id int) float64 {
	score := 0.0
	resourceInfoLock.RLock()
	resourceInfo := ClientResourceStats[id]
	resourceInfoLock.RUnlock()
	if resourceInfo.GPUMemoryFree/1e6 < 5000.0 {
		return -100.0
	}
	score += float64(resourceInfo.MemoryFree) / 1e9
	score += float64(resourceInfo.GPUMemoryFree) / 1e7
	for _, cpu := range resourceInfo.CpuUsage {
		score += 1 - cpu
	}
	return score
}

func scoreRenderClient(id int) float64 {
	score := 0.0
	resourceInfoLock.RLock()
	resourceInfo := ClientResourceStats[id]
	resourceInfoLock.RUnlock()
	if resourceInfo.GPUMemoryFree/1e6 < 2700.0 {
		return -100.0
	}
	score += float64(resourceInfo.MemoryFree) / 1e9
	score += float64(resourceInfo.GPUMemoryFree) / 1e7
	for _, cpu := range resourceInfo.CpuUsage {
		score += 1 - cpu
	}
	return score
}

func chooseResourceClient() int {
	maxScore := 0.0
	maxIndex := -1
	for i := 0; i < clientCnt; i++ {
		score := scoreRenderClient(i)
		log.Println("[chooseResourceClient] this is ", i, " client score: ", score)
		if score > maxScore {
			maxIndex = i
			maxScore = score
		}
	}
	return maxIndex
}

func chooseClient(method string) int {
	switch method {
	case "sequential":
		return chooseSequentialClient()
	case "weighted":
		return chooseResourceClient()
	default:
		log.Println("[chooseClient] invalid client choose methods")
		return -1
	}
}


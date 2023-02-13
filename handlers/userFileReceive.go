package handlers

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync/atomic"

	"github.com/gorilla/mux"
)

func chooseSequentialClient() int {
	atomic.AddInt32(&nowClient, 1)
	return int(atomic.LoadInt32(&nowClient)) % clientCnt
}

func scoreClient(id int) float64 {
	score := 0.0
	resourceInfoLock.RLock()
	resourceInfo := ClientResourceStats[id]
	resourceInfoLock.RUnlock()
	score += float64(resourceInfo.MemoryFree) / 1e9
	score += float64(resourceInfo.GPUMemoryFree) / 1e9
	for _, cpu := range resourceInfo.CpuUsage {
		score += 1 - cpu
	}
	return score
}

func chooseResourceClient() int {
	maxScore := 0.0
	maxIndex := -1
	for i := 0; i < clientCnt; i++ {
		score := scoreClient(i)
		log.Println("this is ", i, " client score: ", score)
		if score > maxScore {
			maxIndex = i
			score = maxScore
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
		log.Println("invalid client choose methods")
		return -1
	}
}

func MakeUserFileReceiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sceneName := vars["name"]
		log.Print("receive user file request: ", sceneName)
		defer r.Body.Close()

		// Create directory to save images, poses, calib.txt

		// os.Mkdir(sceneName, 0644)
		// read multiple files
		reader, err := r.MultipartReader()
		bodyBuffer := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuffer)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			fmt.Printf("FileName=[%s], FormName=[%s]\n", part.FileName(), part.FormName())
			if part.FileName() == "" { // this is FormData
				data, _ := ioutil.ReadAll(part)
				fmt.Printf("FormData=[%s]\n", string(data))
			} else { // This is FileData
				//Filename contains the directory
				name := filepath.Join(sceneName, part.FileName())
				fileWriter, _ := bodyWriter.CreateFormFile("files", name)
				io.Copy(fileWriter, part)
			}
		}
		contentType := bodyWriter.FormDataContentType()
		bodyWriter.Close()

		clientNO := chooseClient("weighted")
		sendAddr := ClientAddrs[clientNO]
		url := sendAddr + "/render/scene/" + sceneName
		log.Print("forward the request to client server: ", url)
		resp, err := http.Post(url, contentType, bodyBuffer)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			resp_body, _ := ioutil.ReadAll(resp.Body)
			log.Fatal("receive error from model controller: ", string(resp_body))
		}
		ClientScenes[sceneName] = clientNO

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("save file success!"))
		log.Println("finish UserFileReceiveHandler")
	}
}

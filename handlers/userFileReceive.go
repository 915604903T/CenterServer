package handlers

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/mux"
)

func chooseClient() int {
	atomic.AddInt32(&nowClient, 1)
	return int(atomic.LoadInt32(&nowClient)) % clientCnt
}
func MakeUserFileReceiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sceneName := vars["name"]
		log.Print("receive user file request: ", sceneName)
		defer r.Body.Close()

		// Create directory to save images, poses, calib.txt
		/*
			os.Mkdir(sceneName, 0644)
			// read multiple files
			reader, err := r.MultipartReader()
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
					dst, _ := os.Create(part.FileName())
					defer dst.Close()
					io.Copy(dst, part)
				}
			}
		*/

		// send file to a client to process voxel scene and relocaliser model
		/*
			files, err := os.ReadDir(sceneName)
			bodyBuffer := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuffer)
			if err != nil {
				log.Fatal(err)
			}
			for _, entry := range files {
				name := filepath.Join("./" + sceneName + entry.Name())
				file, err := os.Open(name)
				if err != nil {
					log.Fatal(err)
				}
				fileWriter, _ := bodyWriter.CreateFormFile("files", name)
				io.Copy(fileWriter, file)
				file.Close()
			}
			contentType := bodyWriter.FormDataContentType()
			bodyWriter.Close()
		*/

		log.Print("forward the request to client server")
		clientNO := chooseClient()
		sendAddr := ClientAddrs[clientNO]
		contentType := "multipart/form-data"
		url := sendAddr + "/render/scene/" + sceneName
		resp, err := http.Post(url, contentType, r.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			resp_body, _ := ioutil.ReadAll(resp.Body)
			log.Fatal("receive error from model controller: ", resp_body)
		}
		ClientScenes[sceneName] = clientNO

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("save file success!"))
		log.Println("finish UserFileReceiveHandler")
	}
}

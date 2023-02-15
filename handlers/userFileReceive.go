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
	"time"

	"github.com/gorilla/mux"
)

func MakeUserFileReceiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sceneName := vars["name"]
		log.Print("[MakeUserFileReceiveHandler] receive user file request: ", sceneName)
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
			// fmt.Printf("FileName=[%s], FormName=[%s]\n", part.FileName(), part.FormName())
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

		// if client does not have enough gpu resource; wait to choose
		clientNO := -1
		for ; clientNO == -1; time.Sleep(time.Second) {
			clientNO = chooseClient("weighted")
		}

		log.Println("[MakeUserFileReceiveHandler] this is client", clientNO, "choose to render ", sceneName)
		sendAddr := ClientAddrs[clientNO]
		url := sendAddr + "/render/scene/" + sceneName
		log.Print("[MakeUserFileReceiveHandler] forward the request to client server: ", url)
		resp, err := http.Post(url, contentType, bodyBuffer)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			resp_body, _ := ioutil.ReadAll(resp.Body)
			log.Fatal("[MakeUserFileReceiveHandler] receive error from model controller: ", string(resp_body))
		}
		ClientScenesLock.Lock()
		ClientScenes[sceneName] = map[int]bool{clientNO: true}
		fmt.Println("this is clientScenes: ", ClientScenes)
		ClientScenesLock.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("save file success!"))
		log.Println("[MakeUserFileReceiveHandler] finish")
	}
}

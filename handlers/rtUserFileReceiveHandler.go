package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

func MakeRTUserFileReceiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sceneName := vars["name"]
		log.Print("[MakeUserFileReceiveHandler] receive user file request: ", sceneName)
		defer r.Body.Close()

		// Create directory to save images, poses, calib.txt
		// read multiple files
		timeout := 0
		picsLength := 0
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
				timeout, _ = strconv.Atoi(string(data))
				fmt.Printf("FormData=[timeout: %d s]\n", timeout)
			} else { // This is FileData
				//Filename contains the directory
				name := sceneName + "/" + part.FileName()
				fileWriter, _ := bodyWriter.CreateFormFile("files", name)
				var data []byte
				_, err := part.Read(data)
				if err != nil {
					log.Println(name, "part read err: ", err)
					panic(err)
				}
				img, format, _ := image.Decode(bytes.NewReader(data))
				log.Println("this is ", name, "format: ", format)
				resizeImg := resize.Resize(uint(img.Bounds().Max.X/2), 0, img, resize.NearestNeighbor)
				err = png.Encode(fileWriter, resizeImg)
				if err != nil {
					log.Println("write ", name, "to body err: ", err)
					panic(err)
				}
				// io.Copy(fileWriter, part)
			}
		}
		contentType := bodyWriter.FormDataContentType()
		bodyWriter.Close()

		// if client does not have enough gpu resource; wait to choose
		clientNO := -1
		for ; clientNO == -1; time.Sleep(time.Second) {
			clientNO = chooseClient("weighted")
		}

		// send file to computing node
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

		// add real time scene to real time processing list
		TimeOutMapLock.Lock()
		TimeOutMap[sceneName] = time.Now().Add(time.Duration(timeout) * time.Second)
		TimeOutMapLock.Unlock()

		sceneLengthLock.Lock()
		sceneLength[sceneName] = picsLength
		sceneLengthLock.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("save file success!"))
		log.Println("[MakeUserFileReceiveHandler] finish")
	}
}

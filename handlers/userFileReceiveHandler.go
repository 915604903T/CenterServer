package handlers

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func MakeUserFileReceiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sceneName := vars["name"]
		log.Print("[MakeUserFileReceiveHandler] receive user file request: ", sceneName)
		defer r.Body.Close()

		/*
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
						name := sceneName + "/" + part.FileName()
						fileWriter, _ := bodyWriter.CreateFormFile("files", name)
						// only resize rgb pic
						if strings.Contains(name, "color") {
							picsLength++
							img, _, _ := image.Decode(part)
							// log.Println("this is ", name, "format: ", format, "size: ", img.Bounds().Max.X, img.Bounds().Max.Y)
							resizeImg := resize.Resize(uint(img.Bounds().Max.X/2), 0, img, resize.NearestNeighbor)
							err = png.Encode(fileWriter, resizeImg)
							if err != nil {
								log.Println("write ", name, "to body err: ", err)
								panic(err)
							}
						} //else {
						io.Copy(fileWriter, part)
						//}
						// io.Copy(fileWriter, part)
					}
				}
			contentType := bodyWriter.FormDataContentType()
			bodyWriter.Close()
		*/

		// if client does not have enough gpu resource; wait to choose
		clientNO := -1
		for ; clientNO == -1; time.Sleep(time.Second) {
			clientNO = chooseClient("weighted")
		}
		sendAddr := ClientAddrs[clientNO]
		url := sendAddr + "/render/scene/" + sceneName
		contentType := r.Header["Content-Type"][0]
		log.Print("[MakeUserFileReceiveHandler] forward the request to client server: ", url, "clientNO:", clientNO)
		log.Println("content type: ", contentType)
		// send to client
		resp, err := http.Post(url, contentType, r.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		respBody, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			log.Fatal("[MakeUserFileReceiveHandler] receive error from model controller: ", string(respBody))
		}
		picsLength, _ := strconv.Atoi(string(respBody))
		log.Println("[MakeUserFileReceiveHandler] ", sceneName, "length: ", picsLength)

		// add video length
		defaultUser := Users[DefaultUserName]
		defaultUser.SceneLengthLock.Lock()
		defaultUser.SceneLength[sceneName] = picsLength
		defaultUser.SceneLengthLock.Unlock()
		// add scene to default user
		SceneUserMapLock.Lock()
		SceneUserMap[sceneName] = DefaultUserName
		SceneUserMapLock.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("save file success!"))
		log.Println("[MakeUserFileReceiveHandler] finish")
	}
}

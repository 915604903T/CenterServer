package handlers

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func MakeRTUserFileReceiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sceneName := vars["name"]
		userName := vars["username"]
		log.Print("[MakeUserFileReceiveHandler] receive user file request: ", sceneName)
		defer r.Body.Close()

		// get the timeout time
		timeout := 0
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
				log.Printf("[MakeUserFileReceiveHandler] FormName=[%s] FormData=[%d s]\n", part.FormName(), timeout)
			} else {
				name := sceneName + "/" + part.FileName()
				fileWriter, _ := bodyWriter.CreateFormFile("files", name)
				io.Copy(fileWriter, part)
			}
		}
		contentType := bodyWriter.FormDataContentType()
		bodyWriter.Close()
		// if client does not have enough gpu resource; wait to choose; init some send para
		clientNO := -1
		for ; clientNO == -1; time.Sleep(time.Second) {
			clientNO = chooseClient("weighted")
		}
		sendAddr := ClientAddrs[clientNO]
		url := sendAddr + "/render/scene/" + sceneName
		log.Print("[MakeUserFileReceiveHandler] forward the request to client server: ", url, "clientNO:", clientNO)
		log.Println("[MakeUserFileReceiveHandler] content type: ", contentType)
		// send to the client
		resp, err := http.Post(url, contentType, bodyBuffer)
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

		// if it is a new user, we need to lock to add it to the map since maybe the same new user can have conflict
		UsersLock.Lock()
		user, ok := Users[userName]
		if !ok {
			newUser := NewUser(userName)
			newUser.SceneLength[sceneName] = picsLength
			newUser.ExpireTime = time.Now().Add(time.Duration(timeout) * time.Second)
			Users[userName] = user
		}
		UsersLock.Unlock()
		if ok {
			// update sceneLength
			user.SceneLengthLock.Lock()
			user.SceneLength[sceneName] = picsLength
			user.SceneLengthLock.Unlock()
			// update expiretime
			user.ExpireTimeLock.Lock()
			user.ExpireTime = time.Now().Add(time.Duration(timeout) * time.Second)
			user.ExpireTimeLock.Unlock()
		}
		// add scene to default user
		SceneUserMapLock.Lock()
		SceneUserMap[sceneName] = DefaultUserName
		SceneUserMapLock.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("save file success!"))
		log.Println("[MakeUserFileReceiveHandler] finish")
	}
}

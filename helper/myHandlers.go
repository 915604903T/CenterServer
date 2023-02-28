package helper

import (
	"net/http"
)

type MyHandlers struct {
	// Deal with user-send video
	UserFileReceiveHandler http.HandlerFunc
	// Deal with user-send video which have realtime request
	RTUserFileReceiveHandler http.HandlerFunc
	// Deal with finish information of model controller
	ModelControllerHandler http.HandlerFunc
	// Deal with global pose from relocalise controller
	RelocaliseControllerHandler http.HandlerFunc
	// Deal with client node resource info
	ClientResourceInfoHandler http.HandlerFunc
	// Deal with mesh return info
	MeshInfoHandler http.HandlerFunc
}

package helper

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var router *mux.Router

// Mark this as a Golang "package"
func init() {
	router = mux.NewRouter()
}

// Router gives access to the underlying router for when new routes need to be added.
func Router() *mux.Router {
	return router
}

func Serve(handlers *MyHandlers) {
	router.HandleFunc("/user/scene/{name}", handlers.UserFileReceiveHandler)
	router.HandleFunc("/sys/model/{name}", handlers.ModelControllerHandler)
	router.HandleFunc("/sys/relocalise/{name}", handlers.RelocaliseControllerHandler)

	tcpPort := 23333
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", tcpPort),
		Handler: router,
	}
	log.Fatal(s.ListenAndServe())
}

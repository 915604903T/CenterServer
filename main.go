package main

import (
	"log"

	"github.com/915604903T/CenterServer/handlers"
	"github.com/915604903T/CenterServer/helper"
)

func main() {
	centerHandlers := helper.MyHandlers{
		UserFileReceiveHandler:      handlers.MakeUserFileReceiveHandler(),
		ModelControllerHandler:      handlers.MakeModelControllerHandler(),
		RelocaliseControllerHandler: handlers.MakeRelocaliseControllerHandler(),
	}
	log.Print("Center server start at port: 23333")
	helper.Serve(&centerHandlers)
}

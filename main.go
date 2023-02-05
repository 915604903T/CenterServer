package main

import (
	"github.com/915604903T/CenterServer/handlers"
	"github.com/915604903T/CenterServer/helper"
)

func main() {
	centerHandlers := helper.MyHandlers{
		UserFileReceiveHandler:      handlers.MakeUserFileReceiveHandler(),
		ModelControllerHandler:      handlers.MakeModelControllerHandler(),
		RelocaliseControllerHandler: handlers.MakeRelocaliseControllerHandler(),
	}

	helper.Serve(&centerHandlers)
}

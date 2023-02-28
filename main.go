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
		ClientResourceInfoHandler:   handlers.MakeClientResourceInfoHandler(),
		MeshInfoHandler:             handlers.MakeMeshInfoHandler(),
	}
	/*
		file_path := "test.png"
		file, err := os.Open(file_path)
		if err != nil {
			log.Fatal(err)
		}
		// decode jpeg into image.Image
		img, err := png.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
		log.Println("img size: ", img.Bounds().Max.X, img.Bounds().Max.Y)
		// resize to width 1000 using Lanczos resampling
		// and preserve aspect ratio
		m := resize.Resize(240, 0, img, resize.NearestNeighbor)
		out, err := os.Create("react.NearestNeighbor.png")
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()
		// write new image to file
		png.Encode(out, m)
		m = resize.Resize(480, 0, img, resize.NearestNeighbor)
		out, err = os.Create("react.png")
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()
		// write new image to file
		png.Encode(out, m)
		log.Println("finish write png")
	*/
	// run a go routine for periodly do relocalisation if candidates exist
	go handlers.RunReloclise()

	// start the server
	log.Print("Center server start at port: 23333")
	helper.Serve(&centerHandlers)
}

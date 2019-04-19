package main

import (
	"enricod/golibraw"
	"fmt"
	"os"
)

func main() {

	fileInfo, err := os.Stat("./DSC05022.ARW")
	if os.IsNotExist(err) {
		fmt.Printf("errore caricamento %s %s", fileInfo.Name(), err.Error())
	} else {
		fmt.Printf("caricamento %s", fileInfo.Name())
		golibraw.Raw2Image(".", fileInfo)

	}

}

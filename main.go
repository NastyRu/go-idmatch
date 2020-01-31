package main

import (
	"github.com/maddevsio/go-idmatch/ocr"
	"fmt"
	"C"
)

//export RecognizeFrame
func RecognizeFrame(name, folder string) {
	p := ocr.Recognize(name, "" ,"" , folder)
	fmt.Printf("Path: %s\n", p)
}

func main() {
	RecognizeFrame("/Users/anastasia/Desktop/working/id-card-detector/k.jpg", "/Users/anastasia/folder")
}

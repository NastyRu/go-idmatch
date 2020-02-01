package main

import (
	"github.com/maddevsio/go-idmatch/ocr"
	"fmt"
	"C"
	"os"
)

//export RecognizeFrame
func RecognizeFrame(name, folder string) {
	fmt.Printf("Name: %s\n", name)
	p := ocr.Recognize(name, "" ,"" , folder)
	fmt.Printf("Path: %s\n", p)
}

func main() {
	RecognizeFrame(os.Args[1], os.Args[2])
}

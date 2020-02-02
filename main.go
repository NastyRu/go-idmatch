package main

import (
	"github.com/NastyRu/go-idmatch/ocr"
	"fmt"
	"C"
	"os"
)

func RecognizeFrame(name, folder string) {
	p := ocr.Recognize(name, "" ,"" , folder)
	fmt.Printf("%s\n", p)
}

func main() {
	RecognizeFrame(os.Args[1], os.Args[2])
}

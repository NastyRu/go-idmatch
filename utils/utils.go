package utils

import (
	"github.com/NastyRu/go-idmatch/log"
	"gocv.io/x/gocv"
)

func showImageInternal(image gocv.Mat, winName string) {
	if !log.IsDebug() {
		return
	}

	window := gocv.NewWindow(winName)
	defer window.Close()
	for {
		window.ResizeWindow(800, 600)
		window.IMShow(image)
		if window.WaitKey(0) >= 0 {
			break
		}
	}
}

func ShowImage(image gocv.Mat) {
	showImageInternal(image, "idmatch")
}

func ShowImageInNamedWindow(image gocv.Mat, winName string) {
	showImageInternal(image, winName)
}

func ShowImageInNamedWindowWithTimeout(image gocv.Mat, winName string, us uint32) {}

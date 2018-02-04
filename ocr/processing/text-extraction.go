package processing

import (
	"crypto/md5"
	"encoding/hex"
	"image"
	"image/color"
	"os"
	"strconv"

	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
)

type block struct {
	x, y, h, w float64
	text       string
}

func TextRegions(img gocv.Mat) [][]image.Point {
	binarized := gocv.NewMat()
	gocv.CvtColor(img, binarized, gocv.ColorBGRToGray)
	kernel := gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{5, 5})
	gocv.MorphologyEx(binarized, binarized, gocv.MorphGradient, kernel)
	gocv.Threshold(binarized, binarized, 0, 255, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	connected := gocv.NewMat()
	kernel = gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{9, 1})
	gocv.MorphologyEx(binarized, connected, gocv.MorphClose, kernel)

	return gocv.FindContours(connected, gocv.RetrievalCComp, gocv.ChainApproxSimple)
}

func extractText(file string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()
	client.SetLanguage("kir", "eng")
	client.SetImage(file)
	return client.Text()
}

func RecognizeRegions(img gocv.Mat, regions [][]image.Point, preview string) (result []block, path string) {
	for k, v := range regions {
		region := gocv.BoundingRect(v)
		// Replace absolute size with relative values
		if region.Dx() < 16 || region.Dy() < 16 || region.Dy() > 64 {
			continue
		}
		roi := img.Region(region)
		file := strconv.Itoa(k) + ".jpeg"
		gocv.IMWrite(file, roi)
		text, err := extractText(file)
		if err != nil {
			continue
		}
		result = append(result, block{
			x:    float64(region.Min.X) / float64(img.Cols()),
			y:    float64(region.Min.Y) / float64(img.Rows()),
			w:    float64(region.Dx()) / float64(img.Cols()),
			h:    float64(region.Dy()) / float64(img.Rows()),
			text: text,
		})
		os.Remove(file)
		gocv.Rectangle(img, gocv.BoundingRect(v), color.RGBA{255, 0, 0, 255}, 2)
	}
	if len(preview) != 0 {
		hash := md5.New()
		hash.Write(img.ToBytes())
		path = preview + "/" + hex.EncodeToString(hash.Sum(nil)) + ".jpeg"
		gocv.IMWrite(path, img)
		// utils.ShowImage(img)
	}
	return result, path
}

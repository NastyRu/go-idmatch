package processing

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"strings"
	"sync"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/templates"

	"github.com/maddevsio/go-idmatch/utils"
	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
)

type block struct {
	x, y, h, w float64
	text       string
}

type extractTextRegionIntCoeff struct {
	w1, h1, w2, h2 int
}

type extractTextRegionFloatCoeff struct {
	w1, h1, w2, h2 float64
}

//
func textRegionsInternal(img gocv.Mat, maxQualitySizes templates.MaxQualitySizesT, fc extractTextRegionIntCoeff) ([][]image.Point, error) {

	symbolHeightCoeff := maxQualitySizes.MaxQualitySymHeight / maxQualitySizes.MaxQualityHeight
	symbolWidthCoeff := maxQualitySizes.MaxQualitySymWidth / maxQualitySizes.MaxQualityWidth

	if fc.w1 == 0 || fc.w2 == 0 || fc.h1 == 0 || fc.h2 == 0 {
		return nil, errors.New("Couldn't find coefficients")
	}

	symbolWidth := int(float64(img.Cols()) * symbolWidthCoeff)
	symbolHeight := int(float64(img.Rows()) * symbolHeightCoeff)

	if symbolWidth < 2 || symbolHeight < 2 {
		return nil, errors.New("Symbol's size too small. Something wrong with region at all")
	}

	original := img.Clone()
	defer original.Close()
	gray := gocv.NewMat()
	defer gray.Close()
	grad := gocv.NewMat()
	defer grad.Close()
	binarization := gocv.NewMat()
	defer binarization.Close()
	opening := gocv.NewMat()
	defer opening.Close()
	connected := gocv.NewMat()
	defer connected.Close()
	kernel := gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{fc.w1, fc.h1})
	defer kernel.Close()

	gocv.CvtColor(original, &gray, gocv.ColorBGRToGray)
	// utils.ShowImageInNamedWindow(gray, "gray")
	gocv.MorphologyEx(gray, &grad, gocv.MorphGradient, kernel)
	// utils.ShowImageInNamedWindow(grad, "gradient")

	gocv.Threshold(grad, &binarization, 0.0, 255.0, gocv.ThresholdBinary|gocv.ThresholdOtsu)
	// utils.ShowImageInNamedWindow(binarization, "binarized")

	kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{fc.w2, fc.h2})
	gocv.MorphologyEx(binarization, &opening, gocv.MorphOpen, kernel)
	// utils.ShowImageInNamedWindow(opening, "opening")

	kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{symbolWidth, 1})
	gocv.MorphologyEx(opening, &connected, gocv.MorphClose, kernel)
	// utils.ShowImageInNamedWindow(connected, "connected")

	regions := gocv.FindContours(connected, gocv.RetrievalCComp, gocv.ChainApproxSimple)
	return regions, nil
}

//TextRegions returns text regions on image
func TextRegions(img gocv.Mat, card templates.Card) ([][]image.Point, error) {
	// showExampleRectangles(img)
	// tryToFindCoeffForNewID(img)
	// buildFloatCoeffs(img)
	// testCoefficientsForID(img)
	w1c := card.TextRegionFilterCoefficients.W1 * float64(img.Cols())
	h1c := card.TextRegionFilterCoefficients.H1 * float64(img.Rows())
	w2c := card.TextRegionFilterCoefficients.W2 * float64(img.Cols())
	h2c := card.TextRegionFilterCoefficients.H2 * float64(img.Rows())
	return textRegionsInternal(img, card.MaxQualitySizes, extractTextRegionIntCoeff{
		int(w1c), int(h1c), int(w2c), int(h2c)})
}

//RecognizeRegions sends found regions to tesseract ocr
func RecognizeRegions(img gocv.Mat, card templates.Card, regions [][]image.Point, preview string) (result []block, path string) {
	//We have to get these values from JSON or somehow from document

	symbolHeightCoeff := card.MaxQualitySizes.MaxQualitySymHeight / card.MaxQualitySizes.MaxQualityHeight
	symbolWidthCoeff := card.MaxQualitySizes.MaxQualitySymWidth / card.MaxQualitySizes.MaxQualityWidth

	symbolWidth := int(float64(img.Cols()) * symbolWidthCoeff)
	symbolHeight := int(float64(img.Rows()) * symbolHeightCoeff)

	gray := gocv.NewMat()
	defer gray.Close()

	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)

	blocks := make(chan block, 15)
	var wg sync.WaitGroup

	for _, v := range regions {
		rect := gocv.BoundingRect(v)
		// roi := img.Region(rect)
		roi := gray.Region(rect)
		if rect.Dx() < symbolWidth || rect.Dy() < symbolHeight/2 || rect.Dy() > symbolHeight*3 {
			continue
		}

		roix4 := gocv.NewMat()
		defer roix4.Close()
		gocv.Resize(roi, &roix4, image.Point{0, 0}, 4, 4, gocv.InterpolationCubic)
		buf, err := gocv.IMEncode(gocv.JPEGFileExt, roix4)
		if err != nil {
			continue
		}

		wg.Add(1)
		go func(buf []byte, v []image.Point) {
			defer wg.Done()
			client := gosseract.NewClient()
			defer client.Close()
			client.SetLanguage("rus")
			client.SetImageFromBytes(buf)
			text, err := client.Text()
			if err != nil {
				return
			}
			// Handle only upper case text. Remove this block if lower case needed
			if text != strings.ToUpper(text) {
				return
			}

			b := block{
				x:    float64(rect.Min.X) / float64(img.Cols()),
				y:    float64(rect.Min.Y) / float64(img.Rows()),
				w:    float64(rect.Dx()) / float64(img.Cols()),
				h:    float64(rect.Dy()) / float64(img.Rows()),
				text: text}

			gocv.Rectangle(&img, gocv.BoundingRect(v), color.RGBA{255, 0, 0, 255}, 2)
			log.Print(log.DebugLevel, fmt.Sprint(b))

			blocks <- b
		}(buf, v)
		// utils.ShowImageInNamedWindow(roix4, fmt.Sprintf("RecognizeRegions: %d %d", rect.Dx(), rect.Dy()))
	}

	wg.Wait()
	close(blocks)

	for b := range blocks {
		result = append(result, b)
	}

	if len(preview) != 0 {
		hash := md5.New()
		hash.Write(img.ToBytes())
		path = preview + "/" + hex.EncodeToString(hash.Sum(nil)) + ".jpeg"
		gocv.IMWrite(path, img)
	}

	utils.ShowImage(img)

	return result, path
}

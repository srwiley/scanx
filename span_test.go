package scanx_test

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"testing"

	"github.com/srwiley/scanx"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func SaveToPngFile(filePath string, m image.Image) error {
	// Create the file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	// Create Writer from file
	b := bufio.NewWriter(f)
	// Write the image into the buffer
	err = png.Encode(b, m)
	if err != nil {
		return err
	}
	err = b.Flush()
	if err != nil {
		return err
	}
	return nil
}

func CompareSpanners(t *testing.T, img1, img2 image.Image, width, height int, op draw.Op) {

	icon, errSvg := oksvg.ReadIcon("testdata/landscapeIcons/beach.svg", oksvg.WarnErrorMode)
	if errSvg != nil {
		fmt.Println("cannot read icon")
		log.Fatal("cannot read icon", errSvg)
		t.FailNow()
	}
	icon.SetTarget(float64(0), float64(0), float64(width), float64(height))

	spanner := scanx.NewImgSpanner(img1)
	spanner.Op = op
	scannerX := scanx.NewScanner(spanner, width, height)
	rasterScanX := rasterx.NewDasher(width, height, scannerX)
	icon.Draw(rasterScanX, 1.0)

	spannerC := &scanx.CompressSpanner{}
	spannerC.Op = op
	spannerC.SetBounds(image.Rect(0, 0, width, height))
	scannerC := scanx.NewScanner(spannerC, width, height)
	rasterScanC := rasterx.NewDasher(width, height, scannerC)
	icon.Draw(rasterScanC, 1.0)
	spannerC.DrawToImage(img2)

	//SaveToPngFile("testdata/imgc.png ", ximgc)
	//SaveToPngFile("testdata/imgx.png ", ximgx)
	var pix1 []uint8
	switch img1 := img1.(type) {
	case *xgraphics.Image:
		pix1 = img1.Pix
	case *image.RGBA:
		pix1 = img1.Pix
	}
	var pix2 []uint8
	switch img2 := img1.(type) {
	case *xgraphics.Image:
		pix2 = img2.Pix
	case *image.RGBA:
		pix2 = img2.Pix
	}

	for i := 0; i < len(pix1); i++ {
		if pix1[i] != pix2[i] {
			t.Error("images do not match at index ", i, pix1[i], pix2[i])
			t.FailNow()
		}
	}
}

func TestSpannersImg(t *testing.T) {
	width := 400
	height := 350

	ximgx := image.NewRGBA(image.Rect(0, 0, width, height))
	ximgc := image.NewRGBA(image.Rect(0, 0, width, height))
	CompareSpanners(t, ximgx, ximgc, width, height, draw.Over)
	CompareSpanners(t, ximgx, ximgc, width, height, draw.Src)

}

func TestSpannersX(t *testing.T) {
	width := 400
	height := 350

	ximgx := xgraphics.New(nil, image.Rect(0, 0, width, height))
	ximgc := xgraphics.New(nil, image.Rect(0, 0, width, height))

	CompareSpanners(t, ximgx, ximgc, width, height, draw.Over)
	CompareSpanners(t, ximgx, ximgc, width, height, draw.Src)

}

// func TestCompose(t *testing.T) {
// 	sp := &scanx.CompressSpanner{}
// 	fmt.Println("cells, index", len(sp.spans))
// 	sp.SetBounds(image.Rect(0, 0, 100, 2))
// 	fmt.Println("cells, index", len(sp.spans))
// 	sp.SetBounds(image.Rect(0, 0, 100, 1))
// 	fmt.Println("cells, index", len(sp.spans))

// 	drawList := func(y int) {
// 		fmt.Print("list at ", y, ":")
// 		cntr := 0
// 		p := sp.spans[y].next
// 		for p != 0 {
// 			spCell := sp.spans[p]
// 			fmt.Print(" sp", spCell)
// 			p = spCell.next
// 			cntr++
// 		}
// 		fmt.Println(" length", cntr)
// 	}

// 	drawList(0)

// 	sp.SetBgColor(color.Black)
// 	sp.SpanOver(0, 20, 40, m)

// 	drawList(0)

// 	sp.SpanOver(0, 5, 10, m)
// 	drawList(0)

// 	sp.SpanOver(0, 80, 90, m)
// 	drawList(0)

// 	sp.SpanOver(0, 60, 70, m)
// 	drawList(0)
// 	sp.SpanOver(0, 70, 75, m)
// 	drawList(0)

// 	sp.Clear()
// 	drawList(0)

// 	sp.SpanOver(0, 20, 40, m)
// 	drawList(0)

// 	sp.SpanOver(0, 30, 50, m)
// 	drawList(0)

// 	sp.SpanOver(0, 10, 25, m)
// 	drawList(0)
// 	sp.SpanOver(0, 33, 37, m)
// 	drawList(0)

// 	sp.Clear()
// 	drawList(0)

// 	sp.SpanOver(0, 20, 40, m)
// 	drawList(0)

// 	sp.SpanOver(0, 20, 40, m)
// 	drawList(0)

// 	sp.SpanOver(0, 10, 40, m)
// 	drawList(0)

// 	sp.SpanOver(0, 20, 50, m)
// 	drawList(0)

// 	sp.Clear()
// 	drawList(0)

// 	sp.SpanOver(0, 20, 30, m)
// 	drawList(0)

// 	sp.SpanOver(0, 40, 50, m)
// 	drawList(0)

// 	sp.SpanOver(0, 10, 60, m)
// 	drawList(0)
// }

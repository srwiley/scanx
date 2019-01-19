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

	icon, errSvg := oksvg.ReadIcon("testdata/landscapeIcons/sea.svg", oksvg.WarnErrorMode)
	//icon, errSvg := oksvg.ReadIcon("testdata/TestShapes.svg", oksvg.WarnErrorMode)
	if errSvg != nil {
		fmt.Println("cannot read icon")
		log.Fatal("cannot read icon", errSvg)
		t.FailNow()
	}
	icon.SetTarget(float64(0), float64(0), float64(width), float64(height))

	spannerC := &scanx.CompressSpanner{}
	spannerC.Op = op
	spannerC.SetBounds(image.Rect(0, 0, width, height))
	scannerC := scanx.NewScanner(spannerC, width, height)
	rasterScanC := rasterx.NewDasher(width, height, scannerC)
	icon.Draw(rasterScanC, 1.0)
	spannerC.DrawToImage(img2)

	spanner := scanx.NewImgSpanner(img1)
	spanner.Op = op
	scannerX := scanx.NewScanner(spanner, width, height)
	rasterScanX := rasterx.NewDasher(width, height, scannerX)
	icon.Draw(rasterScanX, 1.0)

	SaveToPngFile("testdata/imgi.png ", img1)
	SaveToPngFile("testdata/imgc.png ", img2)
	var pix1 []uint8

	switch img1 := img1.(type) {
	case *xgraphics.Image:
		pix1 = img1.Pix
	case *image.RGBA:
		pix1 = img1.Pix
	}
	var pix2 []uint8
	var stride int
	switch img2 := img2.(type) {
	case *xgraphics.Image:
		pix2 = img2.Pix
		stride = img2.Stride
	case *image.RGBA:
		pix2 = img2.Pix
		stride = img2.Stride
	}

	if len(pix1) == 0 {
		t.Error("images are zero sized ")
		t.FailNow()
	}
	i0 := 0
	for y := 0; y < img2.Bounds().Max.Y; y++ {
		i0 = y * stride
		for x := 0; x < img2.Bounds().Max.X; x += 4 {
			if pix1[i0+x] != pix2[i0+x] || pix1[i0+x+1] != pix2[i0+x+1] || pix1[i0+x+2] != pix2[i0+x+2] || pix1[i0+x+3] != pix2[i0+x+3] {
				t.Error("images do not match at index ", y, x/4, "c1", pix1[i0+x], pix1[i0+x+1], pix1[i0+x+2], pix1[i0+x+3],
					"c2", pix2[i0+x], pix2[i0+x+1], pix2[i0+x+2], pix2[i0+x+3])
				t.FailNow()
			}
		}
	}
	// for i := 0; i < len(pix1); i++ {
	// 	if pix1[i] != pix2[i] {
	// 		t.Error("images do not match at index ", i, pix1[i], pix2[i])
	// 		t.FailNow()
	// 	}
	// }
}

func TestSpannersImg(t *testing.T) {
	width := 400
	height := 350
	ximgx := image.NewRGBA(image.Rect(0, 0, width, height))
	ximgc := image.NewRGBA(image.Rect(0, 0, width, height))
	CompareSpanners(t, ximgx, ximgc, width, height, draw.Src)
	ximgx = image.NewRGBA(image.Rect(0, 0, width, height))
	ximgc = image.NewRGBA(image.Rect(0, 0, width, height))
	CompareSpanners(t, ximgx, ximgc, width, height, draw.Over)

}

func TestSpannersX(t *testing.T) {
	width := 400
	height := 350

	ximgx := xgraphics.New(nil, image.Rect(0, 0, width, height))
	ximgc := xgraphics.New(nil, image.Rect(0, 0, width, height))
	CompareSpanners(t, ximgx, ximgc, width, height, draw.Src)

	ximgx = xgraphics.New(nil, image.Rect(0, 0, width, height))
	ximgc = xgraphics.New(nil, image.Rect(0, 0, width, height))
	CompareSpanners(t, ximgx, ximgc, width, height, draw.Over)

}

// func TestCompose(t *testing.T) {
// 	spannerC := &scanx.CompressSpanner{}
// 	spannerC.SetBounds(image.Rect(0, 0, 10, 10))
// 	spannerC.TestSpanAdd()

// }

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

// func (x *CompressSpanner) CheckList(y int, m string) {
// 	cntr := 0
// 	p := x.spans[y].next
// 	for p != 0 {
// 		spCell := x.spans[p]
// 		//fmt.Print(" sp", spCell)
// 		p = spCell.next
// 		if p != 0 {
// 			snCell := x.spans[p]
// 			if spCell.x1 > snCell.x0 {
// 				fmt.Println("bad list at ", y, ":", m)
// 				x.DrawList(y)
// 				os.Exit(1)
// 			}
// 		}
// 		cntr++
// 	}
// }

// //DrawList draws the linked list y
// func (x *CompressSpanner) DrawList(y int) {
// 	fmt.Print("list at ", y, ":")
// 	cntr := 0
// 	p := x.spans[y].next
// 	for p != 0 {
// 		spCell := x.spans[p]
// 		fmt.Print(" ", p, ":sp", spCell)
// 		p = spCell.next
// 		cntr++
// 	}
// 	fmt.Println(" length", cntr)
// }

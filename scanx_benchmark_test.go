// Copyright 2018 The oksvg Authors. All rights reserved.
// created: 2018 by S.R.Wiley
package scanx_test

import (
	"image"

	"testing"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"github.com/srwiley/scanx"
)

func ReadIconSet(folder string, paths []string) (icons []*oksvg.SvgIcon) {
	for _, p := range paths {
		icon, errSvg := oksvg.ReadIcon(folder+p+".svg", oksvg.IgnoreErrorMode)
		if errSvg == nil {
			icons = append(icons, icon)
		}
	}
	return
}

func BenchmarkCompressSpanner5(b *testing.B) {
	RunCompressSpanner(b, 5)
}

func BenchmarkImgSpanner5(b *testing.B) {
	RunImgSpanner(b, 5)
}

func BenchmarkCompressSpanner10(b *testing.B) {
	RunCompressSpanner(b, 10)
}

func BenchmarkImgSpanner10(b *testing.B) {
	RunImgSpanner(b, 10)
}

func BenchmarkCompressSpanner50(b *testing.B) {
	RunCompressSpanner(b, 50)
}

func BenchmarkImgSpanner50(b *testing.B) {
	RunImgSpanner(b, 50)
}

func BenchmarkCompressSpanner250(b *testing.B) {
	RunCompressSpanner(b, 250)
}

func BenchmarkImgSpanner250(b *testing.B) {
	RunImgSpanner(b, 250)
}

func RunCompressSpanner(b *testing.B, mult int) {
	var (
		beachIconNames = []string{
			"beach", "cape", "iceberg", "island",
			"mountains", "sea", "trees", "village"}
		beachIcons = ReadIconSet("testdata/landscapeIcons/", beachIconNames)
		wi, hi     = int(beachIcons[0].ViewBox.W), int(beachIcons[0].ViewBox.H)
		w, h       = wi * mult / 10, hi * mult / 10
		bounds     = image.Rect(0, 0, w, h)
		img        = image.NewRGBA(bounds)

		spannerC    = &scanx.CompressSpanner{}
		scannerC    = scanx.NewScanner(spannerC, w, h)
		rasterScanC = rasterx.NewDasher(w, h, scannerC)
	)
	spannerC.SetBounds(bounds)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ic := range beachIcons {
			ic.SetTarget(0.0, 0.0, float64(bounds.Max.X), float64(bounds.Max.Y))
			ic.Draw(rasterScanC, 1.0)
			spannerC.DrawToImage(img)
			rasterScanC.Clear()
			spannerC.Clear()
		}
	}
}

func RunImgSpanner(b *testing.B, mult int) {
	var (
		beachIconNames = []string{
			"beach", "cape", "iceberg", "island",
			"mountains", "sea", "trees", "village"}
		beachIcons = ReadIconSet("testdata/landscapeIcons/", beachIconNames)
		wi, hi     = int(beachIcons[0].ViewBox.W), int(beachIcons[0].ViewBox.H)
		w, h       = wi * mult / 10, hi * mult / 10
		bounds     = image.Rect(0, 0, w, h)
		img        = image.NewRGBA(bounds)
		//source     = image.NewUniform(color.NRGBA{0, 0, 0, 255})
		//scannerGV = NewScannerGV(w, h, img, img.Bounds())
		//raster    = NewDasher(w, h, scannerGV)
		spanner     = scanx.NewImgSpanner(img)
		scannerX    = scanx.NewScanner(spanner, w, h)
		rasterScanX = rasterx.NewDasher(w, h, scannerX)
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ic := range beachIcons {
			ic.SetTarget(0.0, 0.0, float64(bounds.Max.X), float64(bounds.Max.Y))
			ic.Draw(rasterScanX, 1.0)
			rasterScanX.Clear()
		}
	}
}

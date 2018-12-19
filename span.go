package scanx

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/srwiley/rasterx"
)

type (
	spanCell struct {
		x0, x1, next int
		clr          color.RGBA
	}

	baseSpanner struct {
		// drawing is done with bounds.Min as the origin
		bounds image.Rectangle
		// Op is how pixels are overlayed
		Op      draw.Op
		fgColor color.RGBA
	}

	// CompressSpanner is a Spanner that draws Spans onto a draw.Image
	// interface satisfying struct but it is optimized for *xgraphics.Image
	// and *image.RGBA image types
	// It uses a solid Color only for fg and bg and does not support a color function
	// used by gradients
	CompressSpanner struct {
		baseSpanner
		spans        []spanCell
		bgColor      color.RGBA
		lastY, lastP int
	}

	// ImgSpanner is a Spanner that draws Spans onto *xgraphics.Image
	// or *image.RGBA image types
	// It uses either a color function as a the color source, or a fgColor
	// if colFunc is nil.
	ImgSpanner struct {
		baseSpanner
		pix    []uint8
		stride int

		// xgraphics.Images swap r and b pixel values
		// compared to saved rgb value.
		xpixel    bool
		colorFunc rasterx.ColorFunc
	}
)

// SetFgColor sets the foreground color for blending
func (x *CompressSpanner) SetFgColor(c interface{}) {
	x.fgColor = getColorRGBA(c)
}

// SetBgColor sets the background color for blending
func (x *CompressSpanner) SetBgColor(c interface{}) {
	x.bgColor = getColorRGBA(c)
}

func getColorRGBA(c interface{}) (rgba color.RGBA) {
	switch c := c.(type) {
	case *color.RGBA:
		rgba = *c // direct method why not
	case color.Color:
		r, g, b, a := c.RGBA()
		rgba = color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8)}
	}
	return
}

//Clear clears the current spans
func (x *CompressSpanner) Clear() {
	x.lastY, x.lastP = 0, 0
	x.spans = x.spans[0:0]
	width := x.bounds.Dy()
	for i := 0; i < width; i++ {
		// The first cells are indexed according to the y values
		// to create y separate linked lists corresponding to the
		// image y length. Since index 0 is used by the first of these sentinel cells
		// 0 can and is used for the end of list value by the spanner linked list.
		x.spans = append(x.spans, spanCell{})
	}
}

func (x *CompressSpanner) spansToImage(img draw.Image) {
	for y := 0; y < x.bounds.Dy(); y++ {
		p := x.spans[y].next
		for p != 0 {
			spCell := x.spans[p]
			clr := spCell.clr
			x0, x1 := spCell.x0, spCell.x1
			for x := x0; x < x1; x++ {
				img.Set(y, x, clr)
			}
			p = spCell.next
		}
	}
}

func (x *CompressSpanner) spansToPix(pix []uint8, stride int, xpixel bool) {
	for y := 0; y < x.bounds.Dy(); y++ {
		yo := y * stride
		p := x.spans[y].next
		for p != 0 {
			spCell := x.spans[p]
			i0 := yo + spCell.x0*4
			i1 := i0 + (spCell.x1-spCell.x0)*4
			r, g, b, a := spCell.clr.R, spCell.clr.G, spCell.clr.B, spCell.clr.A
			if xpixel { // R and B are reversed in xgraphics.Image vs image.RGBA
				r, b = b, r
			}
			for i := i0; i < i1; i += 4 {
				pix[i+0] = r
				pix[i+1] = g
				pix[i+2] = b
				pix[i+3] = a
			}
			p = spCell.next
		}
	}
}

//DrawToImage draws the accumulated y spans onto the img
func (x *CompressSpanner) DrawToImage(img image.Image) {
	switch img := img.(type) {
	case *xgraphics.Image:
		x.spansToPix(img.Pix, img.Stride, true)
	case *image.RGBA:
		x.spansToPix(img.Pix, img.Stride, false)
	case draw.Image:
		x.spansToImage(img)
	}
}

// SetBounds sets the spanner boundaries
func (x *CompressSpanner) SetBounds(bounds image.Rectangle) {
	x.bounds = bounds
	x.Clear()
}

func (x *CompressSpanner) blendColor(under color.RGBA, ma uint32) color.RGBA {
	rma := uint32(x.fgColor.R) * ma
	gma := uint32(x.fgColor.G) * ma
	bma := uint32(x.fgColor.B) * ma
	ama := uint32(x.fgColor.A) * ma

	if x.Op != draw.Over || under.A == 0 || ama == 0xFFFF*0xFFFF {
		return color.RGBA{
			uint8(rma / m >> 4),
			uint8(gma / m >> 4),
			uint8(bma / m >> 4),
			uint8(ama / m >> 4)}
	}
	a := (m - (ama / m)) // * 0x101
	return color.RGBA{
		uint8((uint32(under.R)*a + rma) / m >> 4),
		uint8((uint32(under.G)*a + gma) / m >> 4),
		uint8((uint32(under.B)*a + bma) / m >> 4),
		uint8((uint32(under.A)*a + ama) / m >> 4)}
}

func (x *CompressSpanner) addLink(x0, x1, next, pp int, underColor color.RGBA, alpha uint32) int {
	clr := x.blendColor(underColor, alpha)
	if x.spans[pp].x1 >= x0 && ((clr.A == 0 && x.spans[pp].clr.A == 0) || clr == x.spans[pp].clr) { // Just extend the prev span; a new one is not required
		x.spans[pp].x1 = x1
		return pp
	}
	x.spans = append(x.spans,
		spanCell{x0: x0, x1: x1, next: next, clr: clr})
	x.spans[pp].next = len(x.spans) - 1
	return len(x.spans) - 1
}

// GetSpanFunc returns the function that consumes a span described by the parameters.
func (x *CompressSpanner) GetSpanFunc() SpanFunc {
	return x.SpanOver
}

//SpanOver compresses using a solid color and Porter-Duff composition
func (x *CompressSpanner) SpanOver(yi, xi0, xi1 int, ma uint32) {
	if yi != x.lastY {
		x.lastP = yi
		x.lastY = yi
	}
	pp := x.lastP
	p := x.spans[pp].next // first 	qindex  according to local y value
	for p != 0 && xi0 < xi1 {
		sp := x.spans[p]
		if sp.x1 <= xi0 { //sp is before new span
			pp = p
			p = sp.next
			continue
		}
		if sp.x0 >= xi1 { //new span is before sp
			x.addLink(xi0, xi1, p, pp, x.bgColor, ma)
			x.lastP = x.spans[pp].next
			return
		}
		// left span
		if xi0 < sp.x0 {
			pp = x.addLink(xi0, sp.x0, p, pp, x.bgColor, ma)
			xi0 = sp.x0
		} else if xi0 > sp.x0 {
			pp = x.addLink(sp.x0, xi0, p, pp, sp.clr, 0)
		}

		clr := x.blendColor(sp.clr, ma)
		sameClrs := ((clr.A == 0 && x.spans[pp].clr.A == 0) || clr == x.spans[pp].clr)
		if xi1 < sp.x1 {
			// middle span; replaces sp
			if x.spans[pp].x1 >= xi0 && sameClrs {
				x.spans[pp].x1 = xi1
				x.spans[pp].next = sp.next
			} else {
				x.spans[p] = spanCell{x0: xi0, x1: xi1, next: sp.next, clr: clr}
				//right span extends beyond xi1
			}
			x.addLink(xi1, sp.x1, sp.next, p, sp.clr, 0)
			x.lastP = p //sp.next
			return
		}
		// middle span; replaces sp
		if x.spans[pp].x1 >= xi0 && sameClrs {
			x.spans[pp].x1 = sp.x1
			x.spans[pp].next = sp.next
			p = sp.next
			continue
		}
		x.spans[p] = spanCell{x0: xi0, x1: sp.x1, next: sp.next, clr: clr}
		xi0 = sp.x1 // any remaining span starts at sp.x1
		pp = p
		p = sp.next
	}
	x.lastP = pp
	if xi0 < xi1 { // add any remaining span to the end of the chain
		x.addLink(xi0, xi1, 0, pp, x.bgColor, ma)
	}
}

// SetColor sets the color of x to either a color.Color or a rasterx.ColorFunction
func (x *CompressSpanner) SetColor(c interface{}) {
	switch c := c.(type) {
	case color.Color:
		r, g, b, a := c.RGBA()
		x.fgColor = color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
	}
}

// NewImgSpanner returns an ImgSpanner set to draw to the img.
// Img argument must be a *xgraphics.Image or *image.Image type
func NewImgSpanner(img interface{}) (x *ImgSpanner) {
	x = &ImgSpanner{}
	x.SetImage(img)
	return
}

//SetImage set the image that the XSpanner will draw onto
func (x *ImgSpanner) SetImage(img interface{}) {
	switch img := img.(type) {
	case *xgraphics.Image:
		x.pix = img.Pix
		x.stride = img.Stride
		x.xpixel = true
		x.bounds = img.Bounds()
	case *image.RGBA:
		x.pix = img.Pix
		x.stride = img.Stride
		x.xpixel = false
		x.bounds = img.Bounds()
	}
}

// SetColor sets the color of x to either a color.Color or a rasterx.ColorFunction
func (x *ImgSpanner) SetColor(c interface{}) {
	switch c := c.(type) {
	case color.Color:
		x.colorFunc = nil
		r, g, b, a := x.fgColor.RGBA()
		if x.xpixel == true { // apparently r and b values swap in xgraphics.Image
			r, b = b, r
		}
		x.fgColor = color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8)}
	case rasterx.ColorFunc:
		x.colorFunc = c
	}
}

// GetSpanFunc returns the function that consumes a span described by the parameters.
// The next four func declarations are all slightly different
// but in order to reduce code redundancy, this method is used
// to dispatch the function in the draw method.
func (x *ImgSpanner) GetSpanFunc() SpanFunc {
	var (
		useColorFunc = x.colorFunc != nil
		drawOver     = x.Op == draw.Over
	)
	switch {
	case useColorFunc && drawOver:
		return x.SpanColorFunc
	case useColorFunc && !drawOver:
		return x.SpanColorFuncR
	case !useColorFunc && !drawOver:
		return x.SpanFgColorR
	default:
		return x.SpanFgColor
	}
}

const m = 1<<16 - 1

//SpanColorFuncR draw the span using a colorFunc and replaces the previous values.
func (x *ImgSpanner) SpanColorFuncR(yi, xi0, xi1 int, ma uint32) {
	i0 := (yi)*x.stride + (xi0)*4
	i1 := i0 + (xi1-xi0)*4
	cx := xi0
	for i := i0; i < i1; i += 4 {
		rcr, rcg, rcb, rca := x.colorFunc(cx, yi).RGBA()
		if x.xpixel == true {
			rcr, rcb = rcb, rcr
		}
		cx++
		x.pix[i+0] = uint8(rcr * ma / m >> 8)
		x.pix[i+1] = uint8(rcg * ma / m >> 8)
		x.pix[i+2] = uint8(rcb * ma / m >> 8)
		x.pix[i+3] = uint8(rca * ma / m >> 8)
	}
}

//SpanFgColorR draws the span with the fore ground color and replaces the previous values.
func (x *ImgSpanner) SpanFgColorR(yi, xi0, xi1 int, ma uint32) {
	i0 := (yi)*x.stride + (xi0)*4
	i1 := i0 + (xi1-xi0)*4
	cr, cg, cb, ca := x.fgColor.RGBA()
	rma := cr * ma
	gma := cg * ma
	bma := cb * ma
	ama := ca * ma
	for i := i0; i < i1; i += 4 {
		x.pix[i+0] = uint8(rma / m >> 8)
		x.pix[i+1] = uint8(gma / m >> 8)
		x.pix[i+2] = uint8(bma / m >> 8)
		x.pix[i+3] = uint8(ama / m >> 8)

	}
}

//SpanColorFunc draws the span using a colorFunc and the  Porter-Duff composition operator.
func (x *ImgSpanner) SpanColorFunc(yi, xi0, xi1 int, ma uint32) {
	i0 := (yi)*x.stride + (xi0)*4
	i1 := i0 + (xi1-xi0)*4
	cx := xi0

	for i := i0; i < i1; i += 4 {
		// uses the Porter-Duff composition operator.
		rcr, rcg, rcb, rca := x.colorFunc(cx, yi).RGBA()
		if x.xpixel == true {
			rcr, rcb = rcb, rcr
		}
		cx++
		a := (m - (rca * ma / m)) * 0x101
		dr := uint32(x.pix[i+0])
		dg := uint32(x.pix[i+1])
		db := uint32(x.pix[i+2])
		da := uint32(x.pix[i+3])
		x.pix[i+0] = uint8((dr*a + rcr*ma) / m >> 8)
		x.pix[i+1] = uint8((dg*a + rcg*ma) / m >> 8)
		x.pix[i+2] = uint8((db*a + rcb*ma) / m >> 8)
		x.pix[i+3] = uint8((da*a + rca*ma) / m >> 8)
	}
}

//SpanFgColor draw the span using the fore ground color and the Porter-Duff composition operator.
func (x *ImgSpanner) SpanFgColor(yi, xi0, xi1 int, ma uint32) {
	i0 := (yi)*x.stride + (xi0)*4
	i1 := i0 + (xi1-xi0)*4
	// uses the Porter-Duff composition operator.
	cr, cg, cb, ca := x.fgColor.RGBA()
	rma := cr * ma
	gma := cg * ma
	bma := cb * ma
	ama := ca * ma
	if ama == 0xFFFF*0xFFFF { // undercolor is ignored
		for i := i0; i < i1; i += 4 {
			x.pix[i+0] = uint8(rma / m >> 8)
			x.pix[i+1] = uint8(gma / m >> 8)
			x.pix[i+2] = uint8(bma / m >> 8)
			x.pix[i+3] = uint8(ama / m >> 8)
		}
		return
	}
	a := (m - (ama / m)) * 0x101
	for i := i0; i < i1; i += 4 {
		x.pix[i+0] = uint8((uint32(x.pix[i+0])*a + rma) / m >> 8)
		x.pix[i+1] = uint8((uint32(x.pix[i+1])*a + gma) / m >> 8)
		x.pix[i+2] = uint8((uint32(x.pix[i+2])*a + bma) / m >> 8)
		x.pix[i+3] = uint8((uint32(x.pix[i+3])*a + ama) / m >> 8)
	}
}
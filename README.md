# scanx

Warning: Scanx is pre-alpha. There are known bugs being worked on. Please use scanFT or scanGV unless you want to test each svg file your application may use.

Scanx is a fast antialiaser supporting the draw.Image interface and image.RGBA and xgraphics.Image types in particular. It is intended for use with the rasterx package.

Scanx replaces the Painter interface with the Spanner interface that allows for more direct writing to an underlying image type. Scanx has two types that satisfy the Spanner interface; ImgSpanner and CompressSpanner.

ImgSpanner draw into any image that supports the draw.Image interface. It is optimized for image.RGBA and xgraphics.Image types.

CompressSpanner supports the same Image types as ImgSpanner, but stores the spans in y linked lists, where y is the height of the image. It is faster than ImgSpanner for svg icons where the paths overlap significantly, since it only writes to the image after all the spans are collected. The increase in speed is particually significant when drawing to a large image, like a high resolution monitor. However, CompressSpanner does not support gradients, so if you are using them, you should use ImgSpanner instead.

# Example using ImgSpanner:
```golang
bounds     = image.Rect(0, 0, w, h)
img        = image.NewRGBA(bounds)
spanner     = scanx.NewImgSpanner(img)
scanner    = scanx.NewScanner(spanner, w, h)
raster = rasterx.NewDasher(w, h, scanner)
//Use the raster to draw and the results go to the img
``` 
# Example using CompressSpanner:
```golang  
bounds     = image.Rect(0, 0, w, h)
img        = image.NewRGBA(bounds)
spanner    = &scanx.CompressSpanner{}
spanner.SetBounds(bounds)
scanner    = scanx.NewScanner(spanner, w, h)
raster = rasterx.NewDasher(w, h, scanner)
//Use the raster to draw ..
//This draws the accumulated spans onto the image
spanner.DrawToImage(img)
//Get the spanner ready for another image
spanner.Clear()
``` 

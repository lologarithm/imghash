package imghash

import (
	"image"
	"sync"
)

type greyer func(srcImg image.Image, y, width int, output []uint8, wg *sync.WaitGroup)

// GetHash returns the hash for this image.
func GetHash(srcImg image.Image) uint64 {
	greyf := imageToGray

	switch srcImg.(type) {
	case *image.YCbCr:
		// use ycbcr optimized algorithm
		greyf = ycbcrToGray
	}

	// 1. Grayscale
	// greyt := time.Now()
	rect := srcImg.Bounds()
	w, h := rect.Dx(), rect.Dy()
	pix := make([]uint8, 2*w*h)
	grayImage := &image.Gray16{Pix: pix, Stride: 2 * w, Rect: rect}
	var y int
	var st int
	wg := &sync.WaitGroup{}
	for y = rect.Min.Y; y < rect.Max.Y; y++ {
		wg.Add(1)
		st = y * grayImage.Stride
		go greyf(srcImg, y, w, pix[st:st+w*2], wg)
	}
	wg.Wait()
	// log.Printf("  Grayscaling time: %d ms", time.Now().Sub(greyt).Nanoseconds()/int64(time.Millisecond))

	// 2. Break image into 9x8 blocks
	xstride := rect.Max.X / 9
	ystride := rect.Max.Y / 8

	grays := make([]uint32, 8*9)
	// mult := math.MaxUint16 / math.MaxUint8
	xval := rect.Min.X
	yval := rect.Min.Y
	wg.Add(9 * 8)
	for itr := 0; itr < 9*8; itr++ {
		go func(itr, xval, yval int) {
			grays[itr] = getavg(grayImage, xval, yval, xstride, ystride)
			wg.Done()
		}(itr, xval, yval)

		xval += xstride
		if xval >= rect.Max.X {
			xval = rect.Min.X
			yval += ystride
		}
	}
	wg.Wait()

	// resizet := time.Now()
	// thumb := resize.Resize(9, 8, grayImage, resize.Lanczos3)
	// log.Printf("  Sizing time: %d ms", time.Now().Sub(resizet).Nanoseconds()/int64(time.Millisecond))

	// 3. Compare each pixel to the right, if left is brighter, bit = 1
	// compt := time.Now()
	var x int
	var hash uint64
	var bit uint16
	var left, right uint32
	for x = 0; x < 8; x++ {
		for y = 0; y < 8; y++ {
			// r, g, b, _ = thumb.At(x, y).RGBA()
			left = grays[y*8+x]
			// r, g, b, _ = thumb.At(x+1, y).RGBA()/
			right = grays[y*8+x+1] //(r + g + b) / 3
			if left > right {
				hash |= 1 << bit
			}
			bit++
		}
	}
	// log.Printf("  Compare time: %d ms", time.Now().Sub(compt).Nanoseconds()/int64(time.Millisecond))
	return hash
}

func getavg(grayImage *image.Gray16, xval, yval, xstride, ystride int) uint32 {
	divi := uint32(xstride * ystride)
	total := uint32(0)
	for xs := xval; xs < xval+xstride; xs++ {
		for ys := yval; ys < yval+ystride; ys++ {
			total += uint32(grayImage.Pix[grayImage.PixOffset(xs, ys)+1])
		}
	}
	return total / divi
}

func ycbcrToGray(srcImg image.Image, y, width int, output []uint8, wg *sync.WaitGroup) {
	timg := srcImg.(*image.YCbCr)
	var offset int
	for x := 0; x < width; x++ {
		l := timg.Y[timg.YOffset(x, y)]
		output[offset] = 0
		output[offset+1] = uint8(l)
		offset += 2
	}
	wg.Done()
}

func imageToGray(srcImg image.Image, y, width int, output []uint8, wg *sync.WaitGroup) {
	var r, g, b uint32
	var l uint16
	var offset int
	for x := 0; x < width; x++ {
		r, g, b, _ = srcImg.At(x, y).RGBA()
		l = uint16((299*r + 587*g + 114*b + 500) / 1000)
		offset = x * 2
		output[offset] = uint8(l >> 8)
		output[offset+1] = uint8(l)
		// grayImage.SetGray16(x, y, color.Gray16{Y: })
	}
	wg.Done()
}

package imghash

import (
	"image"
	"sync"

	"github.com/nfnt/resize"
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

	// 2. Shrink to 9x8
	// resizet := time.Now()
	thumb := resize.Resize(9, 8, grayImage, resize.Lanczos3)
	// log.Printf("  Sizing time: %d ms", time.Now().Sub(resizet).Nanoseconds()/int64(time.Millisecond))

	// 3. Compare each pixel to the right, if left is brighter, bit = 1
	// compt := time.Now()
	var x int
	var hash uint64
	var bit, left, right, r, g, b uint32
	for x = 0; x < 8; x++ {
		for y = 0; y < 8; y++ {
			r, g, b, _ = thumb.At(x, y).RGBA()
			left = (r + g + b) / 3
			r, g, b, _ = thumb.At(x+1, y).RGBA()
			right = (r + g + b) / 3
			if left > right {
				hash |= 1 << bit
			}
			bit++
		}
	}
	// log.Printf("  Compare time: %d ms", time.Now().Sub(compt).Nanoseconds()/int64(time.Millisecond))
	return hash
}

func ycbcrToGray(srcImg image.Image, y, width int, output []uint8, wg *sync.WaitGroup) {
	timg := srcImg.(*image.YCbCr)
	var offset int
	for x := 0; x < width; x++ {
		output[offset] = timg.Y[timg.YOffset(x, y)]
		output[offset+1] = output[offset]
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

package imghash

import (
	"bytes"
	"image"
	"image/jpeg"
	"log"
	"os"
	"testing"
	"time"
)

func TestHash(t *testing.T) {
	imgFile, err := os.Open("0Uf6biU.jpg")
	if err != nil {
		log.Printf("Failed to read image: %s", err)
		t.FailNow()
	}
	defer imgFile.Close()

	src, err := jpeg.Decode(imgFile)
	if err != nil {
		log.Printf("Failed to decode image: %s", err)
		t.FailNow()
	}
	hash := GetHash(src)
	if hash != 35747317780873727 {
		log.Printf("Incorrect hash created: Expected: %d, Actual: %d", 35747317780873727, hash)
		t.FailNow()
	}
}

func CompareImages(img1 image.Image, img2 image.Image) int {
	h1t1 := time.Now()
	h1 := GetHash(img1)
	h1t2 := time.Now()

	h2t1 := time.Now()
	h2 := GetHash(img2)
	h2t2 := time.Now()

	log.Printf("Hasht1: %d, Hasht2: %d", h1t2.Sub(h1t1).Nanoseconds()/int64(time.Millisecond), h2t2.Sub(h2t1).Nanoseconds()/int64(time.Millisecond))
	// fmt.Printf("\nh1: %064b\nh2: %064b", h1, h2)
	hxor := h1 ^ h2
	// fmt.Printf("\nBA: %064b\n", hxor)
	diff := 0
	for i := uint32(0); i < 64; i++ {
		if hxor&(1<<i) > 0 {
			diff++
		}
	}
	return diff
}

func TestCompareImages(t *testing.T) {
	var imgPairs = []struct {
		file1 string
		file2 string
		diff  int
	}{
		{"0Uf6biU.jpg", "1_Connie.jpg", 34},
		{"1_Connie_scaled.jpg", "1_Connie.jpg", 0},
		{"1_Connie_cropped.jpg", "1_Connie.jpg", 13},
		{"1_Connie_dim.jpg", "1_Connie.jpg", 0},
		{"1_Connie_bright.jpg", "1_Connie.jpg", 0},
	}

	for _, pair := range imgPairs {
		diff := compareImageFiles(pair.file1, pair.file2, t)
		if diff != pair.diff {
			log.Printf("Images are %d bits different but expected: %d", diff, pair.diff)
		}
	}
}

func compareImageFiles(file1, file2 string, t *testing.T) int {
	imgFile1, err := os.Open(file1)
	if err != nil {
		log.Printf("Failed to read image: %s", err)
		t.FailNow()
	}
	defer imgFile1.Close()

	src1, err := jpeg.Decode(imgFile1)
	if err != nil {
		log.Printf("Failed to decode image: %s", err)
		t.FailNow()
	}

	imgFile2, err := os.Open(file2)
	if err != nil {
		log.Printf("Failed to read image: %s", err)
		t.FailNow()
	}
	defer imgFile2.Close()

	src2, err := jpeg.Decode(imgFile2)
	if err != nil {
		log.Printf("Failed to decode image: %s", err)
		t.FailNow()
	}
	diff := CompareImages(src1, src2)
	return diff
}

func TestQualityImages(t *testing.T) {
	imgFile, err := os.Open("1_Connie.jpg")
	if err != nil {
		log.Printf("Failed to read image: %s", err)
		t.FailNow()
	}
	defer imgFile.Close()

	src2, err := jpeg.Decode(imgFile)
	if err != nil {
		log.Printf("Failed to decode image: %s", err)
		t.FailNow()
	}

	src1, _ := shrinkImage(src2, 75, ".jpeg")

	diff := CompareImages(src1, src2)
	if diff != 0 {
		log.Printf("Found diff between same image with diff quality: Expected: %d, Actual: %d", 0, diff)
	}
}

func shrinkImage(oldImg image.Image, quality int, imgType string) (image.Image, error) {
	newImgb := &bytes.Buffer{}
	err := jpeg.Encode(newImgb, oldImg, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, err
	}
	return jpeg.Decode(newImgb)
}

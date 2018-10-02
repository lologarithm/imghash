# ImgHash

ImgHash is an algorithm designed to hash images that are similar closer together than dissimilar images.

This process is known as [Perceptual hashing](https://en.wikipedia.org/wiki/Perceptual_hashing) and uses the following prcedure for comparison.

1. Convert the image to Grayscale.
2. Shrink it to a 9x8 image.
3. For each row of 9, compare each adjacent pair, if left is brighter, set bit = 1.

This results in an 8x8 or uint64 perceptual hash for the given image.

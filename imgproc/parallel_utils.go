package imgproc

import (
	"image"
	"image/color"
	"runtime"
	"sync"
)

//numb of threads
var MaxWorkers = runtime.NumCPU()

//maxworkers
func Parallel(n int, fn func(i int)) {
	if n <= 0 {
		return
	}
	w := MaxWorkers
	if w > n {
		w = n
	}

	sem := make(chan struct{}, w)
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			fn(i)
		}(i)
	}
	wg.Wait()
}

func NextPow2(n int) int {
	k := 1
	for k < n {
		k *= 2
	}
	return k
}

//image to matrix with padding
func ImageToMatrix(img image.Image) ([][]complex128, int, int, int, int) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	wP, hP := NextPow2(width), NextPow2(height)

	matrix := make([][]complex128, hP)
	for i := 0; i < hP; i++ {
		matrix[i] = make([]complex128, wP)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			matrix[y][x] = complex(gray/65535.0, 0)
		}
	}
	return matrix, width, height, wP, hP
}

//grayscale conversion matrix to image
func MatrixToImage(matrix [][]complex128, w, h int) image.Image {
	out := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			val := real(matrix[y][x])
			if val < 0 {
				val = 0
			}
			if val > 1 {
				val = 1
			}
			out.Set(x, y, color.Gray{Y: uint8(val * 255)})
		}
	}
	return out
}

func transpose(matrix [][]complex128) [][]complex128 {
	r, c := len(matrix), len(matrix[0])
	out := make([][]complex128, c)
	for i := range out {
		out[i] = make([]complex128, r)
	}
	for y := 0; y < r; y++ {
		for x := 0; x < c; x++ {
			out[x][y] = matrix[y][x]
		}
	}
	return out
}

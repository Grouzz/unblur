package imgproc

import (
	"math"
	"math/cmplx"
)

func GenerateGaussianKernel(w, h int, sigma float64) [][]complex128 {
	kernel := make([][]complex128, h)
	for i := range kernel {
		kernel[i] = make([]complex128, w)
	}
	cx, cy := float64(w)/2, float64(h)/2
	temp := make([][]float64, h)
	rowSums := make([]float64, h)

	Parallel(h, func(y int) {
		row := make([]float64, w)
		localSum := 0.0
		dy := float64(y) - cy
		for x := 0; x < w; x++ {
			dx := float64(x) - cx
			val := math.Exp(-(dx*dx + dy*dy) / (2 * sigma * sigma))
			row[x] = val
			localSum += val
		}
		temp[y] = row
		rowSums[y] = localSum
	})

	sum := 0.0
	for _, s := range rowSums {
		sum += s
	}
	if sum == 0 {
		sum = 1
	}

	Parallel(h, func(y int) {
		for x := 0; x < w; x++ {
			val := temp[y][x] / sum
			shiftY := (y + h/2) % h
			shiftX := (x + w/2) % w
			kernel[shiftY][shiftX] = complex(val, 0)
		}
	})
	return kernel
}

func ApplyConvolution(image, kernel [][]complex128) [][]complex128 {
	h, w := len(image), len(image[0])
	output := make([][]complex128, h)
	for y := 0; y < h; y++ {
		output[y] = make([]complex128, w)
	}
	Parallel(h, func(y int) {
		for x := 0; x < w; x++ {
			output[y][x] = image[y][x] * kernel[y][x]
		}
	})
	return output
}

func ApplyWiener(image, kernel [][]complex128, K float64) [][]complex128 {
	h, w := len(image), len(image[0])
	output := make([][]complex128, h)
	for y := 0; y < h; y++ {
		output[y] = make([]complex128, w)
	}
	Parallel(h, func(y int) {
		for x := 0; x < w; x++ {
			g := image[y][x]
			hVal := kernel[y][x]
			num := g * cmplx.Conj(hVal)
			mag := cmplx.Abs(hVal)
			den := complex((mag*mag)+K, 0)
			output[y][x] = num / den
		}
	})
	return output
}

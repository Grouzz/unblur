package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"math/cmplx"
	"os"
	"runtime"
	"sync"
	"time"
)


var workers = runtime.NumCPU() //max possible concurrent workers

func main() {
	input_path := flag.String("in", "input.png", "input path")
	output_path := flag.String("out", "output.png", "output path")
	sigma := flag.Float64("sigma", 5.0, "blur radius")
	k_fact := flag.Float64("k", 0.001, "k factor in wiener calculus")
	wFlag := flag.Int("workers", runtime.NumCPU(), "max concurrent workers")
	mode := flag.String("action", "deblur", "choosing between blurring and deblurring")
	flag.Parse()

	workers = *wFlag
	if workers < 1 {
		workers = 1
	}
	//making sure that it uses max OS threads for parallel work
	runtime.GOMAXPROCS(workers)

	fmt.Printf("actually starting the process\naction: %s\nimage: %s -> %s\nsigma: %.2f\nk: %.6f\nworkers: %d\n",
		*mode, *input_path, *output_path, *sigma, *k_fact, workers)

	start := time.Now()

	pixels, w, h, wP, hP, err := load(*input_path)
	if err != nil {
		log.Fatal("loading error: ", err)
	}

	fmt.Printf("height: %d, width: %d\n", h, w)
	fmt.Printf("padded height: %d, padded width: %d\n", hP, wP)

	kernel := generateGaussianKernel(wP, hP, *sigma)

	fmt.Println("fft calculus:")
	fftImage := FFT2D(pixels)
	fftKernel := FFT2D(kernel)
	fmt.Println("done")

	var resultFreq [][]complex128
	if *mode == "blur" {
		fmt.Println("applying blur to the img")
		resultFreq = applyConvolution(fftImage, fftKernel)
	} else {
		fmt.Println("applying filter to deconv")
		resultFreq = applyWiener(fftImage, fftKernel, *k_fact)
	}

	fmt.Println("ifft calculus:")
	finalSpatial := IFFT2D(resultFreq)
	fmt.Println("done")

	err = saveImage(finalSpatial, w, h, *output_path)
	if err != nil {
		log.Fatal("save error: ", err)
	}

	fmt.Printf("ended in %v. result in %s\n", time.Since(start), *output_path)
}

func parallel(n int, fn func(i int)) {
	if n <= 0 {
		return
	}

	w := workers
	if w <= 0 {
		w = 1
	}
	if w > n {
		w = n
	}

	sem := make(chan struct{}, w) //semaphor creation
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		sem <- struct{}{} //taking a  slot
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }() //releasing the taken slot
			fn(i)
		}(i)
	}

	wg.Wait()
}

func load(file_name string) ([][]complex128, int, int, int, int, error) {
	f, err := os.Open(file_name)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	width_padd, height_padd := nextPow2(width), nextPow2(height)

	matrix := make([][]complex128, height_padd)
	for i := 0; i < height_padd; i++ {
		matrix[i] = make([]complex128, width_padd)
	}

	//convert to grey
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			matrix[y][x] = complex(gray/65535.0, 0)
		}
	}

	return matrix, width, height, width_padd, height_padd, nil
}

func saveImage(matrix [][]complex128, w, h int, filename string) error {
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

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, out)
}

func nextPow2(n int) int {
	k := 1
	for k < n {
		k *= 2
	}
	return k
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

//fft: Cooley-Tukey

func FFT1D(data []complex128) []complex128 {
	n := len(data)
	if n <= 1 {
		return data
	}

	even := make([]complex128, n/2)
	odd := make([]complex128, n/2)
	for i := 0; i < n/2; i++ {
		even[i] = data[2*i]
		odd[i] = data[2*i+1]
	}

	even = FFT1D(even)
	odd = FFT1D(odd)

	res := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		t := cmplx.Rect(1, -2*math.Pi*float64(k)/float64(n)) * odd[k]
		res[k] = even[k] + t
		res[k+n/2] = even[k] - t
	}
	return res
}

func IFFT1D(data []complex128) []complex128 {
	n := len(data)
	in := make([]complex128, n)
	for i, v := range data {
		in[i] = cmplx.Conj(v)
	}
	res := FFT1D(in)
	for i := range res {
		res[i] = cmplx.Conj(res[i]) / complex(float64(n), 0)
	}
	return res
}

func FFT2D(matrix [][]complex128) [][]complex128 {
	rows := len(matrix)

	//fft on rows
	pass1 := make([][]complex128, rows)
	parallel(rows, func(i int) {
		pass1[i] = FFT1D(matrix[i])
	})

	//transposing + fft on columns + transposing back
	trans := transpose(pass1)
	ff := func(i int) {
		trans[i] = FFT1D(trans[i])
	}
	parallel(len(trans), ff)

	return transpose(trans)
}

func IFFT2D(matrix [][]complex128) [][]complex128 {
	rows := len(matrix)

	//ifft on rows (parallel)
	pass1 := make([][]complex128, rows)
	ff1 := func(i int) {
		pass1[i] = IFFT1D(matrix[i])
	}
	parallel(rows, ff1)

	trans := transpose(pass1)
	ff2 := func(i int) {
		trans[i] = IFFT1D(trans[i])
	}
	parallel(len(trans), ff2)

	return transpose(trans)
}

// gaussian kernel + circular shift
func generateGaussianKernel(w, h int, sigma float64) [][]complex128 {
	kernel := make([][]complex128, h)
	for i := range kernel {
		kernel[i] = make([]complex128, w)
	}

	cx, cy := float64(w)/2, float64(h)/2

	//computing gaussian values
	temp := make([][]float64, h)
	rowSums := make([]float64, h)
	ff := func(y int) {
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
	}
	parallel(h, ff)

	//reducing the sums
	sum := 0.0
	for _, s := range rowSums {
		sum += s
	}
	if sum == 0 {
		//avoiding division by 0
		sum = 1
	}

	//normalization + circular shift
	parallel(h, func(y int) {
		for x := 0; x < w; x++ {
			val := temp[y][x] / sum
			shiftY := (y + h/2) % h
			shiftX := (x + w/2) % w
			kernel[shiftY][shiftX] = complex(val, 0)
		}
	})

	return kernel
}

func applyConvolution(image, kernel [][]complex128) [][]complex128 {
	h, w := len(image), len(image[0])
	output := make([][]complex128, h)
	for y := 0; y < h; y++ {
		output[y] = make([]complex128, w)
	}

	parallel(h, func(y int) {
		for x := 0; x < w; x++ {
			output[y][x] = image[y][x] * kernel[y][x]
		}
	})

	return output
}

func applyWiener(image, kernel [][]complex128, K float64) [][]complex128 {
	h, w := len(image), len(image[0])
	output := make([][]complex128, h)
	for y := 0; y < h; y++ {
		output[y] = make([]complex128, w)
	}

	roww := func(y int) {
		for x := 0; x < w; x++ {
			g := image[y][x]
			hVal := kernel[y][x]

			num := g * cmplx.Conj(hVal)
			mag := cmplx.Abs(hVal)
			den := complex((mag*mag)+K, 0)
			output[y][x] = num / den
		}
	}

	parallel(h, roww)
	return output
}

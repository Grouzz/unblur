package imgproc

import (
	"math"
	"math/cmplx"
)

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
	pass1 := make([][]complex128, rows)
	Parallel(rows, func(i int) { pass1[i] = FFT1D(matrix[i]) })
	trans := transpose(pass1)
	Parallel(len(trans), func(i int) { trans[i] = FFT1D(trans[i]) })
	return transpose(trans)
}

func IFFT2D(matrix [][]complex128) [][]complex128 {
	rows := len(matrix)
	pass1 := make([][]complex128, rows)
	Parallel(rows, func(i int) { pass1[i] = IFFT1D(matrix[i]) })
	trans := transpose(pass1)
	Parallel(len(trans), func(i int) { trans[i] = IFFT1D(trans[i]) })
	return transpose(trans)
}

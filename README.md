# image blur/deblur

## What it does
Go program that loads an image, converts it to grayscale, pads it to the next power of 2, runs a 2D FFT, then:

- **blur**: multiplies the image spectrum by a Gaussian kernel spectrum (convolution via the frequency domain)
- **deblur**: applies a **Wiener filter** in the frequency domain

## Requirements
- Go
- Input image can be PNG/JPEG

## Run server
From the directory containing your `go.mod` files, run:

```bash
go run cmd/server.go
```

## Run client
From the directory containing your `.go` files, run:

### Blur (example)
You can change file names and the `sigma`/`k` values:

```bash
go run cmd/client.go -in test.jpg -out blurred.png -action blur -sigma 10.0
```

### Deblur
```bash
go run cmd/client.go -in blurred.png -out sharp.png -action deblur -sigma 10.0 -k 0.005
```

## Parameters
This is to help you choose good values during your algorithm tests.

### `-action`
- `blur`: applies a convolution (gaussian blur).
- `deblur`: applies the Wiener filter (attempts to invert the blur).

### `-sigma` (radius)
- The larger it is, the wider/stronger the blur.
- For deblurring to work well, you ideally want to use the same `sigma` that caused the blur.

### `-k` (noise factor for Wiener)
A stabilizing constant that prevents division issues when some frequencies are very small

- Too small (e.g. `0.00001`): sharper result, but lots of noise/artifacts/grain.
- Too large (e.g. `0.1`): less noise, but the image stays blurry.

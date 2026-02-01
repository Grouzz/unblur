# image blur/deblur

## What it does
Go program that loads an image, converts it to grayscale, pads it to the next power-of-2, runs a 2D FFT, then:
- **blur**: multiplies the image spectrum by a Gaussian kernel spectrum (convolution in frequency domain)
- **deblur**: applies a **Wiener filter** in the frequency domain

## Requirements
- Go
- Input image can be PNG/JPEG

## Run
From the directory containing your `.go` file, type:

```bash
go run main.go -in blurred.jpeg -out unblurred.png -action deblur -sigma 5 -k 0.0001
```

### Common examples
**Blur:**
```bash
go run main.go -in input.png -out blurred.png -action blur -sigma 3
```

**Deblur:**
```bash
go run main.go -in blurred.png -out restored.png -action deblur -sigma 5 -k 0.0001
```

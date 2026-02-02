# image blur/deblur

## What it does
Go program that loads an image, converts it to grayscale, pads it to the next power-of-2, runs a 2D FFT, then:
- **blur**: multiplies the image spectrum by a Gaussian kernel spectrum (convolution in frequency domain)
- **deblur**: applies a **Wiener filter** in the frequency domain

## Requirements
- Go
- Input image can be PNG/JPEG

## Run server
From the directory containing your `.go` file, type:

```bash
go run main.go -in blurred.jpeg -out unblurred.png -action deblur -sigma 5 -k 0.0001
go run cmd/server.go
```

## Run client
From the directory containing your `.go` file, type:

**Blur: (example, you can change the names of the png files and the values of k and sigma)**
```bash
go run cmd/client.go -in "test.jpg" -out "flou.png" -action blur -sigma 10.0
```

**Deblur:**
```bash
go run cmd/client.go -in "flou.png" -out "net.png" -action deblur -sigma 10.0 -k 0.005
```

## Résumé technique des paramètres (en français)
Pour t'aider à choisir les bonnes valeurs lors de tes tests d'algorithmes (ce qui fait partie de l'évaluation de performance) :

# -action :

blur : Applique une Convolution (flou gaussien).

deblur : Applique le filtre de Wiener (tentative d'inversion du flou).

# -sigma (Rayon) :

Plus il est grand, plus le flou est large.

Pour que le deblur fonctionne, tu dois idéalement utiliser le même sigma que celui qui a causé le flou.

# -k (Facteur de bruit pour Wiener) :

C'est une valeur pour éviter la division par zéro dans les fréquences faibles.

Trop petit (ex: 0.00001) : L'image sera très nette mais avec beaucoup de "bruit" (artefacts, grains).

Trop grand (ex: 0.1) : L'image restera floue, mais le bruit sera lissé.

Valeur typique : 0.001 à 0.01.
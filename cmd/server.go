package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"net"
	"time"

	"Go_v2/imgproc"
	"Go_v2/protocol"
)

const (
	Port          = ":8080"
	NumberWorkers = 4
)

type Job struct {
	Img        image.Image
	Conf       protocol.Config
	ResultChan chan image.Image
}

func main() {
	//use of max cores
	fmt.Println("Serveur démarré sur", Port)

	jobsChannel := make(chan Job)

	//workers launching
	for w := 1; w <= NumberWorkers; w++ {
		go worker(w, jobsChannel)
	}

	listener, err := net.Listen("tcp", Port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Erreur:", err)
			continue
		}
		go handleClient(conn, jobsChannel)
	}
}

func worker(id int, jobs <-chan Job) {
	for job := range jobs {
		start := time.Now()
		fmt.Printf("Worker %d : Traitement commencé...\n", id)

		//imgproc for image conversion
		matrix, w, h, wP, hP := imgproc.ImageToMatrix(job.Img)

		//kernel generation
		kernel := imgproc.GenerateGaussianKernel(wP, hP, job.Conf.Sigma)

		//fft
		fftImage := imgproc.FFT2D(matrix)
		fftKernel := imgproc.FFT2D(kernel)

		//filtering
		var resultFreq [][]complex128
		if job.Conf.Action == "blur" {
			resultFreq = imgproc.ApplyConvolution(fftImage, fftKernel)
		} else {
			resultFreq = imgproc.ApplyWiener(fftImage, fftKernel, job.Conf.K)
		}

		//ifft and inverse conversion
		finalSpatial := imgproc.IFFT2D(resultFreq)
		outputImg := imgproc.MatrixToImage(finalSpatial, w, h)

		fmt.Printf("Worker %d : Fini en %v\n", id, time.Since(start))
		job.ResultChan <- outputImg
	}
}

func handleClient(conn net.Conn, jobsChannel chan<- Job) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	//config reading
	configData, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}

	var conf protocol.Config
	if err := json.Unmarshal(configData, &conf); err != nil {
		return
	}

	//image reading
	img, _, err := image.Decode(reader)
	if err != nil {
		return
	}

	resultChan := make(chan image.Image)
	jobsChannel <- Job{Img: img, Conf: conf, ResultChan: resultChan}

	//answer
	finalImg := <-resultChan
	png.Encode(conn, finalImg)
}

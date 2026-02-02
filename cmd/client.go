package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net"
	"os"

	"Go_v2/protocol"
)

func main() {
	inputPath := flag.String("in", "input.png", "Input")
	outputPath := flag.String("out", "output.png", "Output")
	sigma := flag.Float64("sigma", 5.0, "Blur radius")
	kFact := flag.Float64("k", 0.001, "K factor")
	mode := flag.String("action", "deblur", "Action")
	serverAddr := flag.String("addr", "127.0.0.1:8080", "Server address")
	flag.Parse()

	conn, err := net.Dial("tcp", *serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	//using shared structure
	conf := protocol.Config{
		Action: *mode,
		Sigma:  *sigma,
		K:      *kFact,
	}

	json.NewEncoder(conn).Encode(conf)

	file, err := os.Open(*inputPath)
	if err != nil {
		log.Fatal("Error opening input file:", err)
	}
	defer file.Close()

	io.Copy(conn, file)
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.CloseWrite()
	}

	//receiving
	imgResult, _, err := image.Decode(conn)
	if err != nil {
		log.Fatal("Error decoding result:", err)
	}

	outFile, _ := os.Create(*outputPath)
	defer outFile.Close()
	png.Encode(outFile, imgResult)

	fmt.Println("Image sauvegardée :", *outputPath)
}

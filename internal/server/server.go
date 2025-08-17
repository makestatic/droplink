// Package server is the server entry point
package server

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"

	"github.com/makestatic/droplink/internal/qr"
)

// Server init & starts the server
func Server(path string) {
	log.Println("Initializing server...")
	port, err := getRandomPort()
	if err != nil {
		log.Fatal(err)
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	log.Println("Server started at ", baseURL)
	{
		qr, err := qr.NewQRCode(baseURL)
		if err != nil {
			log.Fatal("Failed to create QR code:", err)
		}
		if err := qr.Generate(); err != nil {
			log.Fatal("Failed to generate QR code:", err)
		}
		qr.PrintToTerminal()
	}
	http.Handle("/", http.FileServer(http.Dir(path)))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}

// getRandomPort generates a random port
func getRandomPort() (int, error) {
	maxPort := 6969
	minPort := 1024
	port, err := rand.Int(rand.Reader, big.NewInt(int64(maxPort-minPort+1)))
	if err != nil {
		return 0, err
	}
	return minPort + int(port.Int64()), nil
}

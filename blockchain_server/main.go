package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	port := flag.Uint("port", 10111, "TCP Port Number for Blockchain Server")
	flag.Parse()
	app := NewBlockchainServer(uint16(*port))
	fmt.Printf("Starting Blockchain Server on Port %d\n", app.Port())

	go func() { app.Run() }()
	time.Sleep(2 * time.Second)

	initMining()

	wg.Wait()
}

func initMining() {
	url := "http://localhost:10111/mine/start"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to start mining: %v", err)
	}
	defer resp.Body.Close()
}

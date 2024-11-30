package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ahmadexe/prism_chain/blockchain"
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

	dbPath := "volume/blockchain"
    blockchain.InitializeBlockchainDatabase(dbPath)
    defer blockchain.GetDatabaseInstance().Close()

	initMining()

	wg.Wait()
}

func initMining() {
	url := "http://:10111/mine/start"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to start mining: %v", err)
	}
	defer resp.Body.Close()
}

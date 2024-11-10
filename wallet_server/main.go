package main

import "flag"

func main() {
	port := flag.Uint("port", 11101, "TCP Port Number for Wallet Server")
	gateway := flag.String("gateway", ":10111", "Gateway URL for Blockchain Server")
	flag.Parse()
	app := NewWalletServer(uint16(*port), *gateway)
	app.Start()
}

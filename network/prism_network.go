package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ahmadexe/prism_chain/blockchain"
)

var peers []string

func SyncNetwork() *blockchain.Blockchain {
	connectToPeers()

	// TODO: Sync with peers
	
	// TODO: Once the longest chain is found, verify the chain
	// TODO: If verified return the blockchain
	for _, p := range peers {
		fmt.Println(p)
	}

	// Return an empty blockchain for now
	return nil
}

func connectToPeers() {
	for len(peers) < 5 {

		url := "http://localhost:10111/node"

		if len(peers) > 0 {
			url = fmt.Sprintf("http://%s:10111/node", peers[len(peers)-1])
		}

		res, err := http.Get(url)

		if err != nil {
			fmt.Println("Failed to get peers list")
		}

		defer res.Body.Close()

		raw, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Failed to read peers list")
		}

		var data map[string][]string
		err = json.Unmarshal(raw, &data)
		if err != nil {
			fmt.Println("Failed to unmarshal peers list")
		}

		peers = append(peers, data["node"]...)
	}
}

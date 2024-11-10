package network

import (
	"fmt"

	"github.com/ahmadexe/prism_chain/blockchain"
)

var peers []string

func SyncNetwork() *blockchain.Blockchain {
	// TODO: Send request to relay network
	// TODO: Receive list of peers
	// TODO: Update peers list
	// TODO: Sync with peers
	// TODO: Once the longest chain is found, verify the chain
	// TODO: If verified return the blockchain
	for _, p := range peers {
		fmt.Println(p)
	}	
	
	// Return an empty blockchain for now
	return nil
}
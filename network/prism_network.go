package network

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"

	"github.com/ahmadexe/prism_chain/blockchain"
	"github.com/ahmadexe/prism_chain/utils"
)

var peers []string
var IP string

func SyncNetwork() *blockchain.Blockchain {
	// TODO: Send request to relay network
	connectTopeers()
	// TODO: Sync with peers
	// TODO: Once the longest chain is found, verify the chain
	// TODO: If verified return the blockchain
	for _, p := range peers {
		fmt.Println(p)
	}

	// Return an empty blockchain for now
	return nil
}

func connectTopeers() {
	for len(peers) < 5 {
		if len(peers) == 0 {
			res, err := http.Get("http://3.111.196.231:10011/api/v1/rand/node")
			if err != nil {
				log.Print(err)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Print(err)
			}

			peers = append(peers, string(body))
		} else {
			res, err := http.Get("http://" + peers[len(peers)-1] + ":10111" + "/peer")
			if err != nil {
				log.Print(err)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Print(err)
			}
			
			ip := string(body)

			if ip == IP {
				continue
			}

			if utils.Contains(peers, ip) {
				continue
			}

			peers = append(peers, ip)
		}
	}
}

func getIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

func ConnectToRelayNetwork() {
	ip, err := getIP()
	if err != nil {
		log.Print(err)
	}

	IP = ip

	res ,err := http.Post("http://3.111.196.231:10011/api/v1/add/" + ip, "application/json", nil)
	if err != nil {
		log.Print(err)
	}

	res.Body.Close()
}

func GetRandomPeer() string {
	index := rand.Intn(len(peers))
	return peers[index]
}

func GetAllPeers() []string {
	peersCopy := make([]string, len(peers))
	copy(peersCopy, peers)

	return peersCopy
}

package network

import (
	"io"
	"log"
	"math/rand"
	"net/http"

	"github.com/ahmadexe/prism_chain/blockchain"
	"github.com/ahmadexe/prism_chain/utils"
)

var peers []string
var IP string

func SyncNetwork() *blockchain.PeerBlockchain {
	connectTopeers()
	chain := findTheLongestChain()
	return chain
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
				peers = peers[:len(peers)-1]
				continue
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

	res, err := http.Post("http://3.111.196.231:10011/api/v1/add/"+ip, "application/json", nil)
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

func findTheLongestChain() *blockchain.PeerBlockchain {
	var longestChain *blockchain.PeerBlockchain

	for _, p := range peers {
		res, err := http.Get("http://" + p + ":10111")
		if err != nil {
			log.Print(err)
			continue
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Print(err)
			continue
		}

		chain := &blockchain.PeerBlockchain{}
		chain.UnmarshalJSON(body)

		if longestChain == nil {
			longestChain = chain
			continue
		}

		if len(chain.Chain) > len(longestChain.Chain) {
			if verifyChain(chain) {
				longestChain = chain
			}
		}
	}

	return longestChain
}

func verifyChain(chain *blockchain.PeerBlockchain) bool {
	for i := 1; i < len(chain.Chain); i++ {
		if chain.Chain[i].PreviousHash != chain.Chain[i-1].Hash() {
			return false
		}
	}

	return true
}

package network

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/ahmadexe/prism_chain/blockchain"
	"github.com/ahmadexe/prism_chain/utils"
)

var peers []string
var IP string

func SyncNetwork() *blockchain.PeerBlockchain {
	connectToRelayNetwork()
	connectTopeers()
	chain := findTheLongestChain()
	return chain
}

func connectTopeers() {
	var attempts int = 1
	for len(peers) < 2 {
		attempts++

		if attempts == 10 {
			break
		}

		if len(peers) == 0 {
			res, err := http.Get("http://3.111.196.231:10011/api/v1/rand/node")
			if err != nil {
				log.Print(err)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Print(err)
			}

			ip := strings.Trim(string(body), "\"")

			if ip != IP {
				peers = append(peers, ip)
			} 

		} else {
			res, err := http.Get("http://" + peers[len(peers)-1] + ":10111" + "/peer")
			if err != nil {
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

	return strings.Trim(string(ip), "\""), nil
}

func connectToRelayNetwork() {
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

package blockchain

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/ahmadexe/prism_chain/block"
	"github.com/ahmadexe/prism_chain/data"
	"github.com/ahmadexe/prism_chain/utils"
)

var peers []string
var IP string

func SyncNetwork() *BlockchainMeta {
	connectToRelayNetwork()
	connectTopeers()
	chain := findTheLongestChain()
	fmt.Println("Synced with the network")
	fmt.Println("Peers: ", peers)
	fmt.Println("Blockchain: ", chain)
	return chain
}

func connectTopeers() {
	var attempts int = 1
	for len(peers) < 2 {
		attempts++

		if attempts == 100 {
			break
		}

		res, err := http.Get("http://0.0.0.0:10011/api/v1/rand/node")
		if err != nil {
			log.Print(err)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Print(err)
		}

		ip := strings.Trim(string(body), "\"")

		if ip != IP && !utils.Contains(peers, ip) {
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

	res, err := http.Post("http://0.0.0.0:10011/api/v1/add/"+ip, "application/json", nil)
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

func findTheLongestChain() *BlockchainMeta {
	var longestChain *BlockchainMeta

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

		chain := &BlockchainMeta{}
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

func verifyChain(chain *BlockchainMeta) bool {
	for i := 1; i < len(chain.Chain); i++ {
		if chain.Chain[i].PreviousHash != chain.Chain[i-1].Hash() {
			return false
		}
	}

	return true
}

func UpdatePeer(chain *Blockchain) {
	chainRaw, err := chain.MarshalJSON()
	if err != nil {
		log.Print(err)
		return
	}

	for _, p := range peers {
		res, err := http.Post("http://"+p+":10111/sync", "application/json", strings.NewReader(string(chainRaw)))
		if err != nil {
			log.Print(err)
			continue
		}

		res.Body.Close()
	}
}

func UpdatePeersMempool(transaction *block.TransactionRequest) {
	transactionRaw, err := transaction.MarshalJSON()
	if err != nil {
		log.Print(err)
		return
	}

	for _, p := range peers {
		res, err := http.Post("http://"+p+":10111/update/mempool", "application/json", strings.NewReader(string(transactionRaw)))
		if err != nil {
			log.Print(err)
			continue
		}

		res.Body.Close()
	}
}

func UpdatePeersDatapool(data *data.UserData) {
	dataRaw, err := data.MarshalJSON()
	if err != nil {
		log.Print(err)
		return
	}

	for _, p := range peers {
		res, err := http.Post("http://"+p+":10111/update/datapool", "application/json", strings.NewReader(string(dataRaw)))
		if err != nil {
			log.Print(err)
			continue
		}

		res.Body.Close()
	}
}

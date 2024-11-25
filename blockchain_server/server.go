package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/ahmadexe/prism_chain/block"
	"github.com/ahmadexe/prism_chain/blockchain"
	"github.com/ahmadexe/prism_chain/network"
	"github.com/ahmadexe/prism_chain/transaction"
	"github.com/ahmadexe/prism_chain/utils"
	"github.com/ahmadexe/prism_chain/wallet"
)

type BlockchainServer struct {
	port uint16
}

var cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

func (bcs *BlockchainServer) GetBlockchain() *blockchain.Blockchain {
	var minersWallet *wallet.Wallet
	var publicKey, privateKey string
	var option int

	log.Println("Do you want to create a new wallet? (1/0)")
	_, err := fmt.Scanf("%d", &option)
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}

	if option == 1 {
		minersWallet = wallet.NewWallet()
	} else {
		log.Println("Enter the private key")
		_, err := fmt.Scanf("%s", &privateKey)
		if err != nil {
			log.Fatalf("Failed to read input: %v", err)
		}

		log.Println("Enter the public key")
		_, err = fmt.Scanf("%s", &publicKey)
		if err != nil {
			log.Fatalf("Failed to read input: %v", err)
		}

		minersWallet = wallet.GenerateWallet(publicKey, privateKey)
	}

	bc, ok := cache["blockchain"]

	if !ok {
		peerChain := network.SyncNetwork()
		if peerChain != nil {
			chain := blockchain.BuildBlockchain(peerChain.TransactionPool, peerChain.Chain, minersWallet.BlockchainAddress, bcs.Port())

			cache["blockchain"] = chain
			log.Println("Synced with the network")

			log.Printf("Private key: %v\n", minersWallet.PrivateKeyStr())
			log.Printf("Public key: %v\n", minersWallet.PublicKeyStr())
			log.Printf("Blockchain Address key: %v\n", minersWallet.BlockchainAddress)

			return chain
		}

		// Create a new blockchain, this is the first node, a genesis block is created
		bc = blockchain.NewBlockchain(minersWallet.BlockchainAddress, bcs.Port())
		cache["blockchain"] = bc

		log.Println("Created a new blockchain")
		log.Printf("Private key: %v\n", minersWallet.PrivateKeyStr())
		log.Printf("Public key: %v\n", minersWallet.PublicKeyStr())
		log.Printf("Blockchain Address key: %v\n", minersWallet.BlockchainAddress)
	}
	return bc
}

func (bcs *BlockchainServer) GetChain(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		m, _ := json.Marshal(bc)
		w.Write(m)
	default:
		log.Println("Method not allowed")
	}
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{port}
}

func (bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World")
}

func (bcs *BlockchainServer) Transactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockchain()

		m, _ := json.Marshal(struct {
			Transactions []*transaction.Transaction `json:"transactions"`
			Length       int                        `json:"length"`
		}{
			Transactions: bc.TransactionPool,
			Length:       len(bc.TransactionPool),
		})

		w.Write(m)

	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var t block.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("Bad Request")
			return
		}
		if !t.Validate() {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("Bad Request")
			return
		}

		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		signature := utils.SignatureFromString(*t.Signature)

		bc := bcs.GetBlockchain()

		isCreated := bc.CreateTransaction(*t.SenderChainAddress, *t.RecepientChainhainAddress, *t.Value, publicKey, signature)

		w.Header().Add("Content-Type", "application/json")

		if isCreated {
			w.WriteHeader(http.StatusCreated)
			log.Println("Transaction created")
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Bad Request")
	}
}

func (bcs *BlockchainServer) Mine(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		isMined := bc.Mining()
		if isMined {
			log.Println("Mined")
			w.WriteHeader(http.StatusCreated)
			return
		} else {
			log.Println("Not Mined")
			w.WriteHeader(http.StatusConflict)
			return
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("Method not allowed")
	}
}

func (bcs *BlockchainServer) StartMine(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		bc.StartMining()
		m := "Mining started"
		io.WriteString(w, string(m[:]))
		log.Println("Mining started")

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("Method not allowed")
	}
}

func (bcs *BlockchainServer) Amount(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		blockChainAddress := r.URL.Query().Get("blockchain_address")
		chain := bcs.GetBlockchain()
		amount := chain.CalculateBalance(blockChainAddress)
		ar := &blockchain.AmountResponse{Amount: amount}
		m, _ := json.Marshal(ar)

		w.Header().Add("Content-Type", "application/json")
		w.Write(m)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("Method not allowed")
	}
}

func (bcs *BlockchainServer) GetRandomPeer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		peer := network.GetRandomPeer()
		io.WriteString(w, peer)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("Method not allowed")
	}
}

func (bcs *BlockchainServer) IsAlive(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "I'm alive")
}

func (bcs *BlockchainServer) Run() {
	http.HandleFunc("/", bcs.GetChain)
	http.HandleFunc("/transactions", bcs.Transactions)
	http.HandleFunc("/mine", bcs.Mine)
	http.HandleFunc("/mine/start", bcs.StartMine)
	http.HandleFunc("/amount", bcs.Amount)
	http.HandleFunc("/peer", bcs.GetRandomPeer)
	http.HandleFunc("/is_alive", bcs.IsAlive)

	err := http.ListenAndServe(":"+strconv.Itoa(int(bcs.Port())), nil)
	if err != nil {
		log.Fatalf("Failed to start blockchain server: %v", err)
	}
}

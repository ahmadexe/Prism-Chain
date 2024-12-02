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
	"github.com/ahmadexe/prism_chain/data"
	"github.com/ahmadexe/prism_chain/transaction"
	"github.com/ahmadexe/prism_chain/utils"
	"github.com/ahmadexe/prism_chain/wallet"
)

type BlockchainServer struct {
	port uint16
}

var cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

func UpdateCache(meta *blockchain.BlockchainMeta) {
	bc, ok := cache["blockchain"]
	if !ok {
		log.Println("No blockchain found in cache")
		return
	}

	if len(meta.Chain) > len(bc.Chain) {
		cache["blockchain"] = blockchain.BuildBlockchain(meta.TransactionPool, meta.Chain, bc.DataPool, bc.BlockchainAddress, bc.Port)
	}
}

func (bcs *BlockchainServer) GetBlockchain() *blockchain.Blockchain {
	repo := blockchain.GetDatabaseInstance()
	bc, ok := repo.GetBlockchain()
	if !ok {
		return bcs.InitBlockchain()
	}

	return bc
}

func (bcs *BlockchainServer) InitBlockchain() *blockchain.Blockchain {
	var minersWallet *wallet.Wallet
	var publicKey, privateKey string
	var option int

	log.Println("Press 1 to create a new wallet and 0 to fetch a previous one. (1/0)")
	_, err := fmt.Scanf("%d", &option)
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}

	if option == 1 {
		minersWallet = wallet.NewWallet()
		fmt.Println("This is a one time process, you won't see your keys again, copy and save them somewhere safe")
		log.Printf("Private key: %v\n", minersWallet.PrivateKeyStr())
		log.Printf("Public key: %v\n", minersWallet.PublicKeyStr())
		log.Printf("Blockchain Address key: %v\n", minersWallet.BlockchainAddress)
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

	repo := blockchain.GetDatabaseInstance()

	peerChain := blockchain.SyncNetwork()
	if peerChain != nil {
		chain := blockchain.BuildBlockchain(peerChain.TransactionPool, peerChain.Chain, peerChain.Data, minersWallet.BlockchainAddress, bcs.Port())

		repo.SaveBlockchain(chain)
		log.Println("Synced with the network")

		return chain
	}

	// Create a new blockchain, this is the first node, a genesis block is created
	bc := blockchain.NewBlockchain(minersWallet.BlockchainAddress, bcs.Port())
	repo.SaveBlockchain(bc)

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
		bc := bcs.InitBlockchain()
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
		peer := blockchain.GetRandomPeer()
		io.WriteString(w, peer)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("Method not allowed")
	}
}

func (bcs *BlockchainServer) IsAlive(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "I'm alive")
}

func (bcs *BlockchainServer) SyncChain(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPut:
		bc := bcs.GetBlockchain()

		icomingChain := &blockchain.BlockchainMeta{}

		decoder := json.NewDecoder(r.Body)

		err := decoder.Decode(icomingChain)
		if err != nil {
			log.Println("Failed to decode incoming chain")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(icomingChain.Chain) > len(bc.Chain) {
			bc.Chain = icomingChain.Chain
			bc.TransactionPool = icomingChain.TransactionPool

			repo := blockchain.GetDatabaseInstance()
			repo.SaveBlockchain(bc)

			blockchain.UpdatePeer(bc)
		} else {
			log.Println("Incoming chain is not longer than the current chain")
		}

		w.WriteHeader(http.StatusAccepted)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (bcs *BlockchainServer) UpdateMempool(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		bc := bcs.GetBlockchain()

		incomingTransaction := &block.TransactionRequest{}

		decoder := json.NewDecoder(r.Body)

		err := decoder.Decode(incomingTransaction)
		if err != nil {
			log.Println("Failed to decode incoming transaction")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		bc.AddTransaction(*incomingTransaction.SenderChainAddress, *incomingTransaction.RecepientChainhainAddress, *incomingTransaction.Value, utils.PublicKeyFromString(*incomingTransaction.SenderPublicKey), utils.SignatureFromString(*incomingTransaction.Signature))

		w.WriteHeader(http.StatusAccepted)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (bcs *BlockchainServer) UpdateDataPool(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		bc := bcs.GetBlockchain()

		incomingData := &data.UserData{}

		decoder := json.NewDecoder(r.Body)

		err := decoder.Decode(incomingData)

		if err != nil {
			log.Println("Failed to decode incoming data")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		bc.AddData(incomingData)

		w.WriteHeader(http.StatusAccepted)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (bcs *BlockchainServer) AddData(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var d data.UserData
		err := decoder.Decode(&d)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("Bad Request")
			return
		}

		bc := bcs.GetBlockchain()
		bc.AddData(&d)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		log.Println("Data added")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("Method not allowed")
	}
}

func (bcs *BlockchainServer) Run() {
	http.HandleFunc("/", bcs.GetChain)
	http.HandleFunc("/transactions", bcs.Transactions)
	http.HandleFunc("/data", bcs.AddData)
	http.HandleFunc("/mine", bcs.Mine)
	http.HandleFunc("/mine/start", bcs.StartMine)
	http.HandleFunc("/amount", bcs.Amount)
	http.HandleFunc("/peer", bcs.GetRandomPeer)
	http.HandleFunc("/is_alive", bcs.IsAlive)
	http.HandleFunc("/sync", bcs.SyncChain)
	http.HandleFunc("/update/mempool", bcs.UpdateMempool)
	http.HandleFunc("/update/datapool", bcs.UpdateDataPool)

	err := http.ListenAndServe(":"+strconv.Itoa(int(bcs.Port())), nil)
	if err != nil {
		log.Fatalf("Failed to start blockchain server: %v", err)
	}
}

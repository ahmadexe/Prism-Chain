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

const REWARD_PERCENTAGE = 5

func (bcs *BlockchainServer) GetBlockchain() *blockchain.Blockchain {
	repo := blockchain.GetDatabaseInstance()
	bc, ok := repo.GetBlockchain()
	if !ok {
		bcs.InitBlockchain()
		bc, _ = repo.GetBlockchain()
		return bc
	}

	return bc
}

func (bcs *BlockchainServer) InitBlockchain() {
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

		return
	}

	// Create a new blockchain, this is the first node, a genesis block is created
	bc := blockchain.NewBlockchain(minersWallet.BlockchainAddress, bcs.Port())
	repo.SaveBlockchain(bc)
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
			fmt.Println(err)
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

		var isCreated bool

		if t.Share {
			percentage := (*t.Value * REWARD_PERCENTAGE) / 100
			actualValue := *t.Value - percentage

			isCreated = bc.CreateTransaction(*t.SenderChainAddress, *t.RecepientChainAddress, actualValue, publicKey, signature)

			totalDataProviders := 0
			for _, b := range bc.Chain {
				totalDataProviders += len(b.Data)
			}

			reward := percentage / float32(totalDataProviders)
			for _, b := range bc.Chain {
				for _, d := range b.Data {
					if d.BlockchainAddress != *t.SenderChainAddress {
						bc.CreateTransaction(*t.RecepientChainAddress, d.BlockchainAddress, reward, publicKey, signature)
					}
				}
			}
		} else {
			isCreated = bc.CreateTransaction(*t.SenderChainAddress, *t.RecepientChainAddress, *t.Value, publicKey, signature)
		}

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
		bcs.InitBlockchain()
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
	case http.MethodPost:
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
			bc.DataPool = icomingChain.Data

			repo := blockchain.GetDatabaseInstance()
			repo.SaveBlockchain(bc)
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

		bc.AddTransaction(*incomingTransaction.SenderChainAddress, *incomingTransaction.RecepientChainAddress, *incomingTransaction.Value, utils.PublicKeyFromString(*incomingTransaction.SenderPublicKey), utils.SignatureFromString(*incomingTransaction.Signature))

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
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		var d []*data.UserData
		for _, b := range bc.Chain {
			d = append(d, b.Data...)
		}

		m, _ := json.Marshal(struct {
			Data []*data.UserData `json:"data"`
		}{
			Data: d,
		})

		w.Header().Add("Content-Type", "application/json")
		w.Write(m)

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

func (bcs *BlockchainServer) Join(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		address := r.URL.Query().Get("address")

		if address == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		bc := bcs.GetBlockchain()
		bc.DepositJoiningFee(address)

		w.WriteHeader(http.StatusCreated)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (bcs *BlockchainServer) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		blocks := bc.Chain

		var transactions []transaction.Transaction

		for _, b := range blocks {
			for _, t := range b.Transactions {
				transactions = append(transactions, *t)
			}
		}

		if transactions == nil {
			transactions = make([]transaction.Transaction, 0)
		}

		message, _ := json.Marshal(struct {
			Transactions []transaction.Transaction `json:"transactions"`
			Length       int                       `json:"length"`
		}{
			Transactions: transactions,
			Length:       len(transactions),
		})

		w.Header().Add("Content-Type", "application/json")
		w.Write(message)

	default:
		log.Println("Method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (bcs *BlockchainServer) Buy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var br blockchain.BuyRequest

		err := decoder.Decode(&br)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(err)
			return
		}

		if !br.Validate() {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("Bad Request")
			return
		}

		bc := bcs.GetBlockchain()
		bc.BuyCoins(br.RequestAddress, br.Amount)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	http.HandleFunc("/join/", bcs.Join)
	http.HandleFunc("/all_transactions", bcs.GetAllTransactions)
	http.HandleFunc("/buy", bcs.Buy)

	err := http.ListenAndServe(":"+strconv.Itoa(int(bcs.Port())), nil)
	if err != nil {
		log.Fatalf("Failed to start blockchain server: %v", err)
	}
}

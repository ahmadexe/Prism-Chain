package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sync"

	"github.com/ahmadexe/prism_chain/block"
	"github.com/ahmadexe/prism_chain/data"
	"github.com/ahmadexe/prism_chain/transaction"
	"github.com/ahmadexe/prism_chain/utils"
)

type Blockchain struct {
	TransactionPool   []*transaction.Transaction
	Chain             []*block.Block
	DataPool          []*data.UserData
	BlockchainAddress string
	Port              uint16
	mutex             sync.Mutex
}

const (
	MINING_DIFFICULTY = 3
	MINING_REWARD     = 50
	MINING_SENDER     = "PRISM CHAIN"
	MINING_TIMER_SEC  = 20
)

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TransactionPool   []*transaction.Transaction `json:"transactionPool"`
		Chain             []*block.Block             `json:"chain"`
		DataPool          []*data.UserData           `json:"dataPool"`
		BlockchainAddress string                     `json:"blockchainAddress"`
		Port              uint16                     `json:"port"`
	}{
		TransactionPool:   bc.TransactionPool,
		Chain:             bc.Chain,
		DataPool:          bc.DataPool,
		BlockchainAddress: bc.BlockchainAddress,
		Port:              bc.Port,
	})
}

func (bc *Blockchain) UnmarshalJSON(data []byte) error {
	type Alias Blockchain
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(bc),
	}
	return json.Unmarshal(data, &aux)
}

func BuildBlockchain(transactions []*transaction.Transaction, chain []*block.Block, data []*data.UserData, blockchainAddress string, port uint16) *Blockchain {
	return &Blockchain{
		transactions,
		chain,
		data,
		blockchainAddress,
		port,
		sync.Mutex{},
	}
}

func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {
	b := &block.Block{}
	bc := &Blockchain{
		[]*transaction.Transaction{},
		[]*block.Block{},
		[]*data.UserData{},
		blockchainAddress,
		port,
		sync.Mutex{},
	}
	bc.createBlock(0, b.Hash())
	return bc
}

func (bc *Blockchain) createBlock(nonce int, previousHash [32]byte) *block.Block {
	block := block.NewBlock(nonce, previousHash, bc.TransactionPool, bc.DataPool)
	bc.Chain = append(bc.Chain, block)
	bc.DataPool = []*data.UserData{}
	bc.TransactionPool = []*transaction.Transaction{}
	return block
}

func (bc *Blockchain) AddTransaction(senderChainAddress string, recipientChainAddress string, value float32, senderPublicKey *ecdsa.PublicKey, signature *utils.Signature) bool {
	transaction := transaction.NewTransaction(senderChainAddress, recipientChainAddress, value)

	usersBalance := bc.CalculateUserBalance(senderChainAddress)

	repo := GetDatabaseInstance()

	if senderChainAddress == MINING_SENDER {
		bc.TransactionPool = append(bc.TransactionPool, transaction)
		repo.SaveBlockchain(bc)
		return true
	} else if usersBalance >= value {
		bc.TransactionPool = append(bc.TransactionPool, transaction)
		repo.SaveBlockchain(bc)
		return true
	}

	return false
}

func (bc *Blockchain) AddData(userData *data.UserData) {
	bc.DataPool = append(bc.DataPool, userData)
	repo := GetDatabaseInstance()
	repo.SaveBlockchain(bc)
}

func (bc *Blockchain) CreateTransaction(senderChainAddress string, recipientChainAddress string, value float32, senderPublicKey *ecdsa.PublicKey, signature *utils.Signature) bool {
	isTransacted := bc.AddTransaction(senderChainAddress, recipientChainAddress, value, senderPublicKey, signature)


	return isTransacted
}

func (bc *Blockchain) LastBlock() *block.Block {
	return bc.Chain[len(bc.Chain)-1]
}

func (bc *Blockchain) CopyTransactions() []*transaction.Transaction {
	var transactions []*transaction.Transaction
	for _, t := range bc.TransactionPool {
		transaction := &transaction.Transaction{
			SenderChainAddress:    t.SenderChainAddress,
			RecipientChainAddress: t.RecipientChainAddress,
			Value:                 t.Value,
		}

		transactions = append(transactions, transaction)
	}
	return transactions
}

func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transactions []*transaction.Transaction, difficulty int) bool {
	zeroes := strings.Repeat("0", difficulty)
	guessBlock := block.Block{TimeStamp: 0, Nonce: nonce, PreviousHash: previousHash, Transactions: transactions}
	guessHashStr := fmt.Sprintf("%x", guessBlock.Hash())
	return guessHashStr[:difficulty] == zeroes
}

func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactions()
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULTY) {
		nonce++
	}
	return nonce
}

func (bc *Blockchain) Mining() bool {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	if len(bc.TransactionPool) == 0 {
		return false
	}

	bc.AddTransaction(MINING_SENDER, bc.BlockchainAddress, MINING_REWARD, nil, nil)

	nonce := bc.ProofOfWork()

	previousHash := bc.LastBlock().Hash()

	bc.createBlock(nonce, previousHash)

	repo := GetDatabaseInstance()
	repo.SaveBlockchain(bc)

	fmt.Println("Mining is successful!")
	return true
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
}

func (bc *Blockchain) CalculateBalance(address string) float32 {
	var total float32 = 0.0
	for _, b := range bc.Chain {
		for _, t := range b.Transactions {
			value := t.Value
			if address == t.RecipientChainAddress {
				total += value
			} else if address == t.SenderChainAddress {
				total -= value
			}
		}
	}
	return total
}

func (bc *Blockchain) VerifyTransaction(senderPublicKey *ecdsa.PublicKey, sig *utils.Signature, t *transaction.Transaction) bool {
	m, _ := json.Marshal(t)
	hash := sha256.Sum256(m)
	return ecdsa.Verify(senderPublicKey, hash[:], sig.R, sig.S)
}

func (bc *Blockchain) Print() {
	for i, block := range bc.Chain {
		fmt.Println(strings.Repeat("-", 25), i, strings.Repeat("-", 25))
		block.Print()
		fmt.Println(strings.Repeat("-", 53))
	}
}

func (bc *Blockchain) CalculateUserBalance(senderChainAddress string) float32 {
	var total float32 = 0.0
	for _, b := range bc.Chain {
		for _, t := range b.Transactions {
			if senderChainAddress == t.RecipientChainAddress {
				total += t.Value
			}

			if senderChainAddress == t.SenderChainAddress {
				total -= t.Value
			}
		}
	}

	return total
}

func (bc *Blockchain) DepositJoiningFee(userAddress string) {
	for _, b := range bc.Chain {
		for _, t := range b.Transactions {
			if t.SenderChainAddress == MINING_SENDER && t.Value == 10 && t.RecipientChainAddress == userAddress {
				return
			}
		}
	}

	bc.AddTransaction(MINING_SENDER, userAddress, 10, nil, nil)
}

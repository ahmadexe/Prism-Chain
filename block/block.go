package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ahmadexe/prism_chain/data"
	"github.com/ahmadexe/prism_chain/transaction"
)

type Block struct {
	TimeStamp    int64
	Nonce        int
	PreviousHash [32]byte
	Transactions []*transaction.Transaction
	Data         []*data.UserData
}

func NewBlock(nonce int, previousHash [32]byte, transactions []*transaction.Transaction, data []*data.UserData) *Block {
	return &Block{Nonce: nonce, PreviousHash: previousHash, Transactions: transactions, TimeStamp: time.Now().UnixNano(), Data: data}
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TimeStamp    int64                      `json:"timeStamp"`
		Nonce        int                        `json:"nonce"`
		PreviousHash string                     `json:"previousHash"`
		Transactions []*transaction.Transaction `json:"transactions"`
		Data         []*data.UserData           `json:"data"`
	}{
		TimeStamp:    b.TimeStamp,
		Nonce:        b.Nonce,
		PreviousHash: fmt.Sprintf("%x", b.PreviousHash),
		Transactions: b.Transactions,
		Data:         b.Data,
	})
}

func (b *Block) UnmarshalJSON(data []byte) error {
	// Define a temporary type with PreviousHash as a string
	type Alias Block
	aux := &struct {
		PreviousHash string `json:"previousHash"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}

	// Unmarshal into the temporary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert PreviousHash from string to [32]uint8
	bytes, err := hex.DecodeString(aux.PreviousHash)
	if err != nil {
		return fmt.Errorf("invalid previousHash: %w", err)
	}
	if len(bytes) != 32 {
		return fmt.Errorf("invalid previousHash length: got %d, expected 32", len(bytes))
	}
	copy(b.PreviousHash[:], bytes)

	return nil
}

func (b *Block) Print() {
	fmt.Printf("Nonce: %d\n", b.Nonce)
	fmt.Printf("PreviousHash: %x\n", b.PreviousHash)
	fmt.Printf("TimeStamp: %d\n", b.TimeStamp)
	for _, t := range b.Transactions {
		t.Print()
	}
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256(m)
}

type TransactionRequest struct {
	SenderPublicKey       *string  `json:"senderPublicKey"`
	SenderChainAddress    *string  `json:"senderChainAddress"`
	Signature             *string  `json:"signature"`
	RecepientChainAddress *string  `json:"recepientChainAddress"`
	Value                 *float32 `json:"value"`
	Share                 bool     `json:"share"`
}

func (tr *TransactionRequest) Validate() bool {
	return tr.SenderPublicKey != nil && tr.SenderChainAddress != nil && tr.Signature != nil && tr.RecepientChainAddress != nil && tr.Value != nil
}

func (tr *TransactionRequest) MarshalJSON() ([]byte, error) {
	type Alias TransactionRequest
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(tr),
	})
}

func (tr *TransactionRequest) UnmarshalJSON(data []byte) error {
	type Alias TransactionRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(tr),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	return nil
}

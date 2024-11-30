package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Transaction struct {
	SenderChainAddress    string  `json:"senderChainAddress"`
	RecipientChainAddress string  `json:"recipientChainAddress"`
	Value                 float32 `json:"value"`
}

func NewTransaction(senderAddress string, recipientAddress string, val float32) *Transaction {
	return &Transaction{senderAddress, recipientAddress, val}
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	transactionMap := map[string]interface{}{
		"senderChainAddress":    t.SenderChainAddress,
		"recipientChainAddress": t.RecipientChainAddress,
		"value":                 t.Value,
	}

	return json.Marshal(transactionMap)
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	type Alias Transaction

	temp := &struct {
		*Alias
	}{
		Alias: (*Alias)(t), 
	}

	if err := json.Unmarshal(data, temp); err != nil {
		return fmt.Errorf("error unmarshaling transaction: %w", err)
	}

	if t.SenderChainAddress == "" {
		return errors.New("senderChainAddress cannot be empty")
	}
	if t.RecipientChainAddress == "" {
		return errors.New("recipientChainAddress cannot be empty")
	}
	if t.Value <= 0 {
		return errors.New("value must be greater than 0")
	}

	return nil
}

func (t *Transaction) Print() {
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("senderChainAddres: %s\n", t.SenderChainAddress)
	fmt.Printf("recipientChainAddress: %s\n", t.RecipientChainAddress)
	fmt.Printf("value: %f\n", t.Value)
	fmt.Println(strings.Repeat("-", 40))
}

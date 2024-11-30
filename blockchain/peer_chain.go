package blockchain

import (
	"encoding/json"

	"github.com/ahmadexe/prism_chain/block"
	"github.com/ahmadexe/prism_chain/transaction"
)

type BlockchainMeta struct {
	TransactionPool []*transaction.Transaction
	Chain           []*block.Block
}

func (chain *BlockchainMeta) UnmarshalJSON(data []byte) error {
	type Alias BlockchainMeta
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(chain),
	}
	return json.Unmarshal(data, &aux)
}

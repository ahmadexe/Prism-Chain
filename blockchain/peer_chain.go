package blockchain

import (
	"encoding/json"

	"github.com/ahmadexe/prism_chain/block"
	"github.com/ahmadexe/prism_chain/transaction"
)

type PeerBlockchain struct {
	TransactionPool   []*transaction.Transaction
	Chain             []*block.Block
}

func (chain *PeerBlockchain) UnmarshalJSON(data []byte) error {
	type Alias PeerBlockchain
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(chain),
	}
	return json.Unmarshal(data, &aux)
}


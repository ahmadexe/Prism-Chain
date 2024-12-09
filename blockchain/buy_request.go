package blockchain

type BuyRequest struct {
	RequestAddress string  `json:"requestAddress"`
	Amount         float32 `json:"amount"`
}

func (br *BuyRequest) Validate() bool {
	return br.RequestAddress != "" && br.Amount > 0
}

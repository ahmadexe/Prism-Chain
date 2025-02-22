package transaction

type TransactionResponse struct {
	SenderPrivateKey           *string  `json:"senderPrivateKey"`
	SenderPublicKey            *string  `json:"senderPublicKey"`
	SenderBlockchainAddress    *string  `json:"senderBlockchainAddress"`
	RecipientBlockchainAddress *string  `json:"recipientBlockchainAddress"`
	Value                      *float32 `json:"value"`
	Share                      bool     `json:"share"`
}

func (tr *TransactionResponse) Validate() bool {
	return tr.SenderPrivateKey != nil && tr.SenderPublicKey != nil && tr.SenderBlockchainAddress != nil && tr.RecipientBlockchainAddress != nil && tr.Value != nil
}

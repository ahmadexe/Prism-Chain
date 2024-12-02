package data

import "encoding/json"

type UserData struct {
	BlockchainAddress string   `json:"blockchainAddress"`
	Data              []string `json:"data"`
}

func (u *UserData) MarshalJSON() ([]byte, error) {
	type Alias UserData 
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(u),
	})
}

func (u *UserData) UnmarshalJSON(data []byte) error {
	type Alias UserData 
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	return json.Unmarshal(data, aux)
}
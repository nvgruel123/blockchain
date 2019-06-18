package data

import "encoding/json"

type RegisterData struct {
	AssignedId  int32  `json:"assignedId"`
	PeerMapJson string `json:"peerMapJson"`
}

func NewRegisterData(id int32, peerMapJson string) RegisterData {
	data := RegisterData{id, peerMapJson}
	return data
}

func (data *RegisterData) EncodeToJson() (string, error) {
	byteArr, err := json.Marshal(data)
	return string(byteArr), err
}

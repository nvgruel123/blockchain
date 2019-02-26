package p2

import (
	"encoding/json"
)

type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

func (blockchain *BlockChain) Initial() {
	blockchain.Chain = make(map[int32][]Block)
	blockchain.Length = 0
}

func (blockchain *BlockChain) Get(height int) []Block {
	chain := blockchain.Chain
	array := chain[int32(height)]
	return array
}

func (blockchain *BlockChain) Insert(block Block) {

	height := block.Header.Height
	chain := blockchain.Chain
	block_arr := chain[height]
	if len(block_arr) == 0 {
		blockchain.Length += 1
	}
	block_arr = append(block_arr, block)
	chain[height] = block_arr
}

func EncodeToJson(self BlockChain) (string, error) {
	chain := self.Chain
	block_arr := []JsonBlock{}
	for _, v := range chain {
		for _, block := range v {
			json_block := &JsonBlock{}
			json_block.Hash = block.Header.Hash
			json_block.Height = block.Header.Height
			json_block.ParentHash = block.Header.ParentHash
			json_block.Size = block.Header.Size
			json_block.Timestamp = block.Header.Timestamp
			json_block.Mpt = block.Value.Data
			block_arr = append(block_arr, *json_block)
		}
	}
	byte_arr, _ := json.Marshal(block_arr)
	return string(byte_arr), nil
}

func (blockchain *BlockChain) DecodeFromJson(jsonString string) {
	block_arr := &[]JsonBlock{}
	byte_arr := []byte(jsonString)
	json.Unmarshal(byte_arr, block_arr)
	for _, json_block := range *block_arr {
		json_result, _ := json.Marshal(json_block)
		block := DecodeFromJson(string(json_result))
		blockchain.Insert(block)

	}
}

func DecodeJsonToBlockChain(jsonString string) (BlockChain, error) {
	blockchain := BlockChain{}
	blockchain.Initial()
	block_arr := &[]JsonBlock{}
	byte_arr := []byte(jsonString)
	json.Unmarshal(byte_arr, block_arr)
	for _, json_block := range *block_arr {
		json_result, _ := json.Marshal(json_block)
		block := DecodeFromJson(string(json_result))
		blockchain.Insert(block)

	}
	return blockchain, nil
}

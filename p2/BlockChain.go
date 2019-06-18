package p2

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"

	"github.com/nvgruel123/cs686-blockchain-p3-nvgruel123/p1"
	"golang.org/x/crypto/sha3"
)

type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

func (blockchain *BlockChain) Initial() {

	blockchain.Chain = make(map[int32][]Block)
	blockchain.Length = 0

}

func NewBlockChain() BlockChain {
	bc := BlockChain{}
	bc.Initial()
	return bc
}

func (blockchain *BlockChain) Get(height int32) ([]Block, bool) {

	chain := blockchain.Chain
	array := chain[int32(height)]
	if array == nil {
		return []Block{}, false
	}

	return array, true
}

func (blockchain *BlockChain) Insert(block Block) {

	height := block.Header.Height
	chain := blockchain.Chain
	block_arr := chain[height]
	if block.Header.Height > blockchain.Length {
		blockchain.Length = block.Header.Height
	}
	flag := false
	for _, value := range block_arr {
		if value.Header.Hash == block.Header.Hash {
			flag = true
		}
	}
	if flag == false {
		block_arr = append(block_arr, block)
	}
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
			json_block.User = block.User
			block_arr = append(block_arr, *json_block)
		}
	}
	byte_arr, err := json.Marshal(block_arr)

	return string(byte_arr), err
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

func (bc *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bc.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bc.Chain[int32(id)] {
			hashs = append(hashs, block.Header.Hash+"<="+block.Header.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}

func (bc *BlockChain) GenBlock(mpt p1.MerklePatriciaTrie) Block {
	height := bc.Length
	chain := bc.Chain[height]
	var block Block
	length := len(chain)
	pBlock := chain[rand.Intn(length)]
	parentHash := pBlock.Header.Hash
	block.Initial(height+1, parentHash, mpt)
	bc.Insert(block)
	return block
}

func (bc *BlockChain) GetLatestBlocks() []Block {
	return bc.Chain[bc.Length]
}

func (bc *BlockChain) GetParentBlock(block Block) (Block, bool) {

	height := block.Header.Height - 1
	hash := block.Header.ParentHash
	blocks := bc.Chain[height]
	for _, block := range blocks {
		if block.Header.Hash == hash {
			return block, true
		}
	}
	return Block{}, false
}

func (bc *BlockChain) GetFirstBlock() Block {
	return bc.Chain[0][0]
}

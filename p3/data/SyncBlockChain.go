package data

import (
	"sync"

	"github.com/nvgruel123/cs686-blockchain-p3-nvgruel123/p1"
	"github.com/nvgruel123/cs686-blockchain-p3-nvgruel123/p2"
)

type SyncBlockChain struct {
	bc  p2.BlockChain
	mux sync.Mutex
}

func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: p2.NewBlockChain()}
}

func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {

	return sbc.bc.Get(height)
}

func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {

	blockArr, _ := sbc.bc.Get(height)
	for _, block := range blockArr {
		if block.Header.Hash == hash {
			return block, true
		}
	}
	return p2.Block{}, false
}

func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {

	height := insertBlock.Header.Height
	parentHash := insertBlock.Header.ParentHash
	if height > 1 {
		_, flag := sbc.GetBlock(height-1, parentHash)
		return flag
	} else {
		return false
	}
}

func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {

	sbc.mux.Lock()
	bc, _ := p2.DecodeJsonToBlockChain(blockChainJson)
	sbc.bc = bc
	sbc.mux.Unlock()

}

func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {

	result, err := p2.EncodeToJson(sbc.bc)

	return result, err
}

func (sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie) p2.Block {

	sbc.mux.Lock()
	block := sbc.bc.GenBlock(mpt)
	defer sbc.mux.Unlock()

	return block
}

func (sbc *SyncBlockChain) Show() string {
	return sbc.bc.Show()
}

func (sbc *SyncBlockChain) GetLatestBlocks() []p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetLatestBlocks()
}

func (sbc *SyncBlockChain) GetParentBlock(block p2.Block) (p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetParentBlock(block)
}

func (sbc *SyncBlockChain) GetFirstBlock() p2.Block {
	return sbc.bc.GetFirstBlock()
}

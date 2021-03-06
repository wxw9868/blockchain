package blc

import "wxw-blockchain/database"

//区块迭代器
type blockchainIterator struct {
	CurrentBlockHash []byte
	BD               *database.BlockchainDB
}

//获取区块迭代器实例
func NewBlockchainIterator(blc *Blockchain) *blockchainIterator {
	blockchainIterator := &blockchainIterator{blc.BD.View([]byte(LastBlockHashMapping), database.BlockBucket), blc.BD}
	return blockchainIterator
}

//迭代下一个区块信息
func (bi *blockchainIterator) Next() *Block {
	currentByte := bi.BD.View(bi.CurrentBlockHash, database.BlockBucket)
	if len(currentByte) == 0 {
		return nil
	}
	block := Block{}
	block.Deserialize(currentByte)
	bi.CurrentBlockHash = block.PreHash
	return &block
}

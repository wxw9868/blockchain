package main

import "wxw-blockchain/blc"

func main() {

	//blcs := blc.NewBlockchain()
	//genesisBlock := blcs.CreateBlockchainWithGenesisBlock()

	cli := blc.CLI{}
	cli.Run()

	//// 新区块
	//data := "add block 100"
	//genesisBlock.AddBlockToBlockchain(data)
	//
	//data = "add block 200"
	//genesisBlock.AddBlockToBlockchain(data)
	//
	//data = "add block 300"
	//genesisBlock.AddBlockToBlockchain(data)
	//
	//data = "add block 400"
	//genesisBlock.AddBlockToBlockchain(data)
	//
	//data = "add block 500"
	//genesisBlock.AddBlockToBlockchain(data)

	//blcs.PrintAllBlockInfo()
}

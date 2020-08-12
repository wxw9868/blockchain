package cli

import "wxw-blockchain/blc"

func (cli *CLI) printChain() {
	newBlc := blc.NewBlockchain()
	newBlc.PrintAllBlockInfo()
}

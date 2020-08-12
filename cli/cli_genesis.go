package cli

import "wxw-blockchain/blc"

func (cli *CLI) genesis(address string, value int) {
	newBlc := blc.NewBlockchain()
	newBlc.CreataGenesisTransaction(address, value)
}

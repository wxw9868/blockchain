package cli

import "wxw-blockchain/blc"

func (cli *CLI) transfer(from, to string, amount string) {
	newBlc := blc.NewBlockchain()
	newBlc.CreateTransaction(from, to, amount)
}

package cli

import (
	"fmt"
	"wxw-blockchain/blc"
)

func (cli *CLI) transfer(from, to string, amount string) {
	newBlc := blc.NewBlockchain()
	newBlc.CreateTransaction(from, to, amount)
	fmt.Println("已执行转帐命令")
}

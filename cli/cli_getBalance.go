package cli

import (
	"fmt"
	"wxw-blockchain/blc"
)

func (cli *CLI) getBalance(address string) {
	newBlc := blc.NewBlockchain()
	balance := newBlc.GetBalance(address)
	fmt.Printf("地址:%s的余额为：%d\n", address, balance)
}

package cli

import (
	"fmt"
	"wxw-blockchain/blc"
)

func (cli *CLI) printAddressList() {
	fmt.Println("查看所有钱包地址")

	wallet, _ := blc.NewWallet()
	for address, _ := range wallet.Wallets {
		fmt.Println(address)
	}
}

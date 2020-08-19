package cli

import (
	"fmt"
	"wxw-blockchain/blc"
)

func (cli *CLI) createWallet() {
	wallet, _ := blc.NewWallet()
	wallet.CreateWallet()
	fmt.Println(len(wallet.Wallets))
}

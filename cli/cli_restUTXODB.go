package cli

import (
	"fmt"
	"wxw-blockchain/blc"
)

func (cli *CLI) resetUTXODB() {
	bc := blc.NewBlockchain()
	utxos := blc.UTXOHandle{BC: bc}
	utxos.ResetUTXODataBase()
	fmt.Println("已重置UTXO数据库")
}

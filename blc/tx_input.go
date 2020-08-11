package blc

//UTXO输入
type TXInput struct {
	TxHash    []byte
	Index     int
	Signature string
}

package blc

import "bytes"

//UTXO输出
type TXOutput struct {
	Value         int
	PublicKeyHash []byte //公钥hash
}

func (ou *TXOutput) unLockTXOutput(address string) bool {
	publicKeyHash := Ripemd160Hash(address)
	return bytes.Compare(ou.PublicKeyHash, publicKeyHash) == 0
}

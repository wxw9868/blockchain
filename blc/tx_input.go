package blc

import "bytes"

//UTXO输入
type TXInput struct {
	TxHash    []byte
	Index     int
	Signature []byte //数字签名
	PublicKey []byte //公钥
}

func (in *TXInput) unLockTXInput(publicKeyHash []byte) bool {
	ripHash := PublicKeyHash(in.PublicKey)
	return bytes.Compare(ripHash, publicKeyHash) == 0
}

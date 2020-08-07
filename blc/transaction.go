package blc

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	log "github.com/corgi-kx/logcustom"
	"wxw-blockchain/util"
)

//交易列表信息
type Transaction struct {
	TxHash []byte
	//UTXO输入
	Vint []TXInput
	//UTXO输出
	Vout []TXOutput
}

//对此笔交易的输入,输出进行hash运算后存入交易hash(txhash)
func (t *Transaction) hash() {
	tBytes := t.Serialize()
	//加入随机数byte（目的是什么？）
	randomNumber := util.GenerateRealRandom()
	randomByte := util.Int64ToBytes(randomNumber)
	sumByte := bytes.Join([][]byte{tBytes, randomByte}, []byte(""))
	hashByte := sha256.Sum256(sumByte)
	t.TxHash = hashByte[:]
}

//将整笔交易里的成员依次转换成字节数组,拼接成整体后 返回
func (t *Transaction) getTransBytes() []byte {
	if t.TxHash == nil || t.Vout == nil {
		log.Panic("交易信息不完整，无法拼接成字节数组")
		return nil
	}
	transBytes := []byte{}
	transBytes = append(transBytes, t.TxHash...)
	for _, v := range t.Vint {
		transBytes = append(transBytes, v.TxHash...)
		transBytes = append(transBytes, util.Int64ToBytes(int64(v.Index))...)
		transBytes = append(transBytes, v.Signature...)
		transBytes = append(transBytes, v.PublicKey...)
	}
	for _, v := range t.Vout {
		transBytes = append(transBytes, util.Int64ToBytes(int64(v.Value))...)
		transBytes = append(transBytes, v.PublicKeyHash...)
	}
	return transBytes
}

// 将transaction序列化成[]byte
func (t *Transaction) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(t)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

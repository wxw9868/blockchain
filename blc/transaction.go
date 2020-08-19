package blc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	log "github.com/corgi-kx/logcustom"
	"math/big"
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

// IsCoinbase checks whether the transaction is coinbase
func (t *Transaction) IsCoinbase() bool {
	return len(t.Vint) == 1 && len(t.Vint[0].TxHash) == 0 && t.Vint[0].Index == -1
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

func (t *Transaction) Sign(privateKey *ecdsa.PrivateKey, prevTss map[string]Transaction) {
	if t.IsCoinbase() {
		return
	}

	txCopy := t.TrimmedCopy()

	for inID, vint := range txCopy.Vint {
		prevTs := prevTss[hex.EncodeToString(vint.TxHash)]
		txCopy.Vint[inID].Signature = nil
		txCopy.Vint[inID].PublicKey = prevTs.Vout[vint.Index].PublicKeyHash
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vint[inID].PublicKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, privateKey, txCopy.TxHash)
		if err != nil {
			log.Fatal(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		t.Vint[inID].Signature = signature
	}
}

func (tx Transaction) Hash() []byte {
	txCopy := tx
	txCopy.TxHash = []byte{}
	hash := sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

func (t *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vint := range t.Vint {
		inputs = append(inputs, TXInput{vint.TxHash, vint.Index, nil, nil})
	}

	for _, vout := range t.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PublicKeyHash})
	}

	txCopy := Transaction{t.TxHash, inputs, outputs}

	return txCopy
}

func (t *Transaction) Verify(prevTss map[string]Transaction) bool {
	txCopy := t.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vint := range t.Vint {
		prevTs := prevTss[hex.EncodeToString(vint.TxHash)]
		txCopy.Vint[inID].Signature = nil
		txCopy.Vint[inID].PublicKey = prevTs.Vout[vint.Index].PublicKeyHash
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vint[inID].PublicKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vint.Signature)
		r.SetBytes(vint.Signature[:(sigLen / 2)])
		s.SetBytes(vint.Signature[sigLen/2:])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vint.PublicKey)
		x.SetBytes(vint.PublicKey[:keyLen/2])
		y.SetBytes(vint.PublicKey[keyLen/2:])

		rawPublicKey := ecdsa.PublicKey{curve, &x, &y}
		if !ecdsa.Verify(&rawPublicKey, txCopy.TxHash, &r, &s) {
			return false
		}
	}
	return true
}

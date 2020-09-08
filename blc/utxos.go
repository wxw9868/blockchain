package blc

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"wxw-blockchain/database"

	"github.com/boltdb/bolt"
	log "github.com/corgi-kx/logcustom"
)

type UTXOHandle struct {
	BC *Blockchain
}

//重置UTXO数据库
func (u *UTXOHandle) ResetUTXODataBase() {
	//先查找全部未花费UTXO
	utxosMap := u.BC.findAllUTXOs()
	if len(utxosMap) == 0 {
		log.Debug("找不到区块,暂不重置UTXO数据库")
		return
	}
	fmt.Println("utxosMap: ", utxosMap)
	//删除旧的UTXO数据库
	if database.IsBucketExist(u.BC.BD, database.UTXOBucket) {
		u.BC.BD.DeleteBucket(database.UTXOBucket)
	}

	//创建并将未花费UTXO循环添加
	for k, v := range utxosMap {
		u.BC.BD.Put([]byte(k), u.serialize(v), database.UTXOBucket)
	}
}

//根据地址未消费的utxo
func (u *UTXOHandle) findUTXOFromAddress(address string) []*UTXO {
	publicKeyHash := Ripemd160Hash(address)
	utxosSlice := []*UTXO{}
	//获取bolt迭代器，遍历整个UTXO数据库
	//打开数据库
	var DBFileName = "blockchain_" + ListenPort + ".db"
	db, err := bolt.Open(DBFileName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(database.UTXOBucket))
		if b == nil {
			return errors.New("datebase view err: not find bucket")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			txOutputs := u.dserialize(v)
			for _, utxo := range txOutputs.UTXOS {
				if bytes.Equal(utxo.Vout.PublicKeyHash, publicKeyHash) {
					utxosSlice = append(utxosSlice, utxo)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	//关闭数据库
	err = db.Close()
	if err != nil {
		log.Panic("db close err :", err)
	}
	return utxosSlice
}

//传入交易信息,将交易里的输出添加进utxo数据库,并剔除输入信息
func (u *UTXOHandle) Synchrodata() {
	bci := NewBlockchainIterator(u.BC)

	block := bci.Next()

	ins := []*TXInput{}

	outsMap := make(map[string]*TXOuputs)

	// 找到所有我要删除的数据
	for _, ts := range block.Transactions {
		for _, in := range ts.Vint {
			ins = append(ins, &in)
		}
	}

	for _, ts := range block.Transactions {

		utxos := []*UTXO{}

		for index, out := range ts.Vout {
			isSpent := false
			for _, in := range ins {
				publicKeyHash := PublicKeyHash(in.PublicKey)
				if index == in.Index && bytes.Equal(ts.TxHash, in.TxHash) && bytes.Equal(out.PublicKeyHash, publicKeyHash) {
					isSpent = true
					continue
				}

				if !isSpent {
					utxo := &UTXO{Hash: ts.TxHash, Index: index, Vout: out}
					utxos = append(utxos, utxo)
				}
			}
		}

		if len(utxos) > 0 {
			txHash := hex.EncodeToString(ts.TxHash)
			outsMap[txHash] = &TXOuputs{utxos}
		}
	}

	var DBFileName = "blockchain_" + ListenPort + ".db"
	db, err := bolt.Open(DBFileName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(database.UTXOBucket))

		if b != nil {

			// 删除
			for _, in := range ins {
				txOutputsBytes := b.Get(in.TxHash)

				if len(txOutputsBytes) == 0 {
					continue
				}

				txOutputs := u.dserialize(txOutputsBytes)

				utxos := []*UTXO{}

				// 判断是否需要
				isNeedDelete := false

				for _, utxo := range txOutputs.UTXOS {
					if in.Index == utxo.Index && bytes.Equal(utxo.Vout.PublicKeyHash, PublicKeyHash(in.PublicKey)) {
						isNeedDelete = true
					} else {
						utxos = append(utxos, utxo)
					}
				}

				if isNeedDelete {
					b.Delete(in.TxHash)
					if len(utxos) > 0 {
						txHash := hex.EncodeToString(in.TxHash)
						preTXOutputs := outsMap[txHash]
						preTXOutputs.UTXOS = append(preTXOutputs.UTXOS, utxos...)
						outsMap[txHash] = preTXOutputs
					}
				}
			}

			// 新增
			for keyHash, outPuts := range outsMap {
				keyHashBytes, _ := hex.DecodeString(keyHash)
				b.Put(keyHashBytes, u.serialize(outPuts))
			}
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	err = db.Close()
	if err != nil {
		log.Panic("db close err :", err)
	}
}

func (u *UTXOHandle) serialize(tXOuputs *TXOuputs) []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tXOuputs)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func (u *UTXOHandle) dserialize(d []byte) *TXOuputs {
	var model *TXOuputs
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&model)
	if err != nil {
		log.Panic(err)
	}
	return model
}

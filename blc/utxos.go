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

	// 存储未花费的UTXO
	outsMap := make(map[string]*TXOuputs)

	// 找到所有我要删除的数据
	txInputs := []*TXInput{}
	for _, ts := range block.Transactions {
		for _, in := range ts.Vint {
			txInputs = append(txInputs, &in)
		}
	}
	fmt.Println("txInputs1: ", txInputs)

	for _, ts := range block.Transactions {
		utxos := []*UTXO{}
		for index, out := range ts.Vout {
			isSpent := false
			for _, txInput := range txInputs {
				publicKeyHash := PublicKeyHash(txInput.PublicKey)
				if index == txInput.Index && bytes.Equal(ts.TxHash, txInput.TxHash) && bytes.Equal(out.PublicKeyHash, publicKeyHash) {
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
	fmt.Println("outsMap1: ", outsMap)

	var DBFileName = "blockchain_" + ListenPort + ".db"
	db, err := bolt.Open(DBFileName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(database.UTXOBucket))
		fmt.Println("b: ", b)
		if b != nil {
			// 删除
			fmt.Println("txInputs2: ", txInputs)
			for _, txInput := range txInputs {
				txOutputsBytes := b.Get(txInput.TxHash)
				fmt.Println("txOutputsBytes: ", txOutputsBytes)
				if len(txOutputsBytes) == 0 {
					continue
				}
				txOutputs := u.dserialize(txOutputsBytes)
				fmt.Println("txOutputs: ", txOutputs)

				utxos := []*UTXO{}

				// 判断是否需要
				isNeedDelete := false
				fmt.Println("txOutputs.UTXOS: ", txOutputs.UTXOS)
				for _, utxo := range txOutputs.UTXOS {
					if txInput.Index == utxo.Index && bytes.Equal(utxo.Vout.PublicKeyHash, PublicKeyHash(txInput.PublicKey)) {
						isNeedDelete = true
					} else {
						utxos = append(utxos, utxo)
					}
				}
				fmt.Println("isNeedDelete: ", isNeedDelete)
				if isNeedDelete {
					err = b.Delete(txInput.TxHash)
					if err != nil {
						panic(err)
					}
					fmt.Println("utxos:", utxos)
					if len(utxos) > 0 {
						txHash := hex.EncodeToString(txInput.TxHash)
						preTXOutputs := outsMap[txHash]
						preTXOutputs.UTXOS = append(preTXOutputs.UTXOS, utxos...)
						outsMap[txHash] = preTXOutputs
					}
				}
			}

			fmt.Println("outsMap2: ", outsMap)

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

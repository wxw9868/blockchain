package blc

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"wxw-blockchain/database"

	log "github.com/corgi-kx/logcustom"
)

type Blockchain struct {
	BD *database.BlockchainDB //封装的blot结构体
}

//创建区块链实例
func NewBlockchain() *Blockchain {
	blockchain := Blockchain{}
	bd := database.New()
	blockchain.BD = bd
	return &blockchain
}

//创建创世区块交易信息
func (blc *Blockchain) CreataGenesisTransaction(address string, value int) {
	//创世区块数据
	txi := TXInput{[]byte{}, -1, nil, nil}

	txo := TXOutput{value, Ripemd160Hash(address)}

	ts := Transaction{nil, []TXInput{txi}, []TXOutput{txo}}

	ts.hash()

	tss := []Transaction{ts}
	//开始生成区块链的第一个区块
	blc.newGenesisBlockchain(tss)

	//重置utxo数据库，将创世数据存入
	utxos := UTXOHandle{blc}
	utxos.ResetUTXODataBase()
}

func (blc *Blockchain) newGenesisBlockchain(transaction []Transaction) {
	//判断一下是否已生成创世区块
	if len(blc.BD.View([]byte(LastBlockHashMapping), database.BlockBucket)) != 0 {
		log.Fatal("不可重复生成创世区块")
	}
	//生成创世区块
	genesisBlock := newGenesisBlock(transaction)
	//添加到数据库中
	blc.AddBlock(genesisBlock)
}

//添加区块信息到数据库，并更新lastHash
func (blc *Blockchain) AddBlock(block *Block) {
	blc.BD.Put(block.Hash, block.Serialize(), database.BlockBucket)
	bci := NewBlockchainIterator(blc)
	currentBlock := bci.Next()
	if currentBlock == nil || currentBlock.Height < block.Height {
		blc.BD.Put([]byte(LastBlockHashMapping), block.Hash, database.BlockBucket)
	}
}

//创建UTXO交易实例
func (blc *Blockchain) CreateTransaction(from, to string, amount string) {
	//判断一下是否已生成创世区块
	if len(blc.BD.View([]byte(LastBlockHashMapping), database.BlockBucket)) == 0 {
		log.Error("还没有生成创世区块，不可进行转账操作 !")
		return
	}

	var fromSlice []string
	var toSlice []string
	var amountSlice []int

	//对传入的信息进行校验检测
	err := json.Unmarshal([]byte(from), &fromSlice)
	if err != nil {
		log.Error("json err:", err)
		return
	}
	err = json.Unmarshal([]byte(to), &toSlice)
	if err != nil {
		log.Error("json err:", err)
		return
	}
	err = json.Unmarshal([]byte(amount), &amountSlice)
	if err != nil {
		log.Error("json err:", err)
		return
	}
	if len(fromSlice) != len(toSlice) || len(fromSlice) != len(amountSlice) {
		log.Error("转账数组长度不一致")
		return
	}

	var tss []Transaction
	wallet, _ := NewWallet()
	for k, address := range fromSlice {
		publicKey := wallet.Wallets[address].PublicKey

		u := UTXOHandle{blc}
		//获取数据库中的未消费的utxo
		utxos := u.findUTXOFromAddress(address)
		if len(utxos) == 0 {
			log.Errorf("%s 余额为0,不能进行转帐操作", address)
			return
		}
		//将utxos添加上未打包进区块的交易信息
		if tss != nil {
			for _, ts := range tss {
				//先添加未花费utxo 如果有的话就不添加
			tagVout:
				for index, vOut := range ts.Vout {
					if bytes.Compare(vOut.PublicKeyHash, PublicKeyHash(publicKey)) != 0 {
						continue
					}
					for _, utxo := range utxos {
						if bytes.Equal(ts.TxHash, utxo.Hash) && index == utxo.Index {
							continue tagVout
						}
					}
					utxos = append(utxos, &UTXO{ts.TxHash, index, vOut})
				}
				//剔除已花费的utxo
				for _, vInt := range ts.Vint {
					for index, utxo := range utxos {
						if bytes.Equal(vInt.TxHash, utxo.Hash) && vInt.Index == utxo.Index {
							utxos = append(utxos[:index], utxos[index+1:]...)
						}
					}
				}
			}
		}

		//utxos := blc.UTXOs(address, tss)
		balance, indexs := blc.doUTXOs(utxos, amountSlice[k], address)

		//打包交易的核心操作
		var newTXInput []TXInput
		var newTXOutput []TXOutput

		for txHash, indexArray := range indexs {
			txHashBytes, _ := hex.DecodeString(txHash)
			for _, index := range indexArray {
				tXInput := TXInput{
					TxHash:    txHashBytes,
					Index:     index,
					Signature: nil,
					PublicKey: publicKey,
				}
				newTXInput = append(newTXInput, tXInput)
			}
		}
		tXOutput := TXOutput{amountSlice[k], Ripemd160Hash(toSlice[k])}
		newTXOutput = append(newTXOutput, tXOutput)
		//打包交易的核心操作
		tXOutput = TXOutput{balance - amountSlice[k], Ripemd160Hash(address)}
		newTXOutput = append(newTXOutput, tXOutput)

		ts := Transaction{nil, newTXInput, newTXOutput}
		ts.hash()

		//数字签名
		blc.SignTransaction(&ts, wallet.Wallets[address].PrivateKey, tss)

		tss = append(tss, ts)
	}

	//挖矿奖励(奖励给转账用户)
	ts := blc.miningReward(fromSlice[0], 10)
	tss = append(tss, ts)

	blc.addBlockchain(tss)
}

//挖矿奖励
func (blc *Blockchain) miningReward(address string, value int) Transaction {
	//创世区块数据
	txi := TXInput{[]byte{}, -1, nil, []byte{}}

	txo := TXOutput{value, Ripemd160Hash(address)}

	ts := Transaction{nil, []TXInput{txi}, []TXOutput{txo}}

	ts.hash()
	return ts
}

//数字签名
func (blc *Blockchain) SignTransaction(ts *Transaction, privateKey *ecdsa.PrivateKey, tss []Transaction) {
	if ts.IsCoinbase() {
		return
	}

	prevTss := make(map[string]Transaction)

	for _, vint := range ts.Vint {
		prevTs, err := blc.FindTransaction(vint.TxHash, tss)
		if err != nil {
			log.Fatal(err)
		}
		prevTss[hex.EncodeToString(prevTs.TxHash)] = prevTs
	}

	ts.Sign(privateKey, prevTss)
}

func (blc *Blockchain) FindTransaction(ID []byte, tss []Transaction) (Transaction, error) {
	for _, ts := range tss {
		if bytes.Compare(ts.TxHash, ID) == 0 {
			return ts, nil
		}
	}

	bci := NewBlockchainIterator(blc)

	for {
		block := bci.Next()

		for _, ts := range block.Transactions {
			if bytes.Compare(ts.TxHash, ID) == 0 {
				return ts, nil
			}
		}

		if len(block.PreHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("FindTransaction err : Transaction is not found")
}

func (blc *Blockchain) doUTXOs(utxos []*UTXO, amount int, address string) (int, map[string][]int) {
	indexs := make(map[string][]int)
	var balance int
	for _, utxo := range utxos {
		balance = balance + utxo.Vout.Value
		key := hex.EncodeToString(utxo.Hash)
		indexs[key] = append(indexs[key], utxo.Index)
	}
	if amount > balance {
		fmt.Println(address + "余额不足！")
		os.Exit(1)
	}
	return balance, indexs
}

//将交易添加进区块链中(内含挖矿操作)
func (blc *Blockchain) addBlockchain(transaction []Transaction) {
	//查询数据
	preBlockbyte := blc.BD.View(blc.BD.View([]byte(LastBlockHashMapping), database.BlockBucket), database.BlockBucket)
	preBlock := Block{}
	preBlock.Deserialize(preBlockbyte)
	height := preBlock.Height + 1

	_tss := []Transaction{}
	for _, ts := range transaction {
		if !ts.IsCoinbase() {
			if !blc.VerifyTransaction(ts, _tss) {
				log.Panic("ERROR: Invalid transaction")
			}
			_tss = append(_tss, ts)
		}
	}

	//进行挖矿
	newBlock := NewBlock(transaction, blc.BD.View([]byte(LastBlockHashMapping), database.BlockBucket), height)

	//将区块添加到本地库中
	blc.AddBlock(newBlock)

	//将数据同步到UTXO数据库中
	u := UTXOHandle{blc}
	u.ResetUTXODataBase()
	//更新数据
	//u.Synchrodata()
}

func (blc *Blockchain) VerifyTransaction(ts Transaction, tss []Transaction) bool {
	prevTss := make(map[string]Transaction)

	for _, vint := range ts.Vint {
		prevTs, err := blc.FindTransaction(vint.TxHash, tss)
		if err != nil {
			log.Fatal(err)
		}
		prevTss[hex.EncodeToString(prevTs.TxHash)] = prevTs
	}

	return ts.Verify(prevTss)
}

//用户未花费UTXO
func (blc *Blockchain) UTXOs(address string, tss []Transaction) []UTXO {
	var txOutputs []UTXO
	txInputs := make(map[string][]int)

	for _, ts := range tss {
		for _, in := range ts.Vint {
			//用户花费UTXO
			if in.unLockTXInput(Ripemd160Hash(address)) {
				key := hex.EncodeToString(in.TxHash)
				txInputs[key] = append(txInputs[key], in.Index)
			}
		}
	}

	for _, ts := range tss {
	tsVout:
		for index, ou := range ts.Vout {
			if ou.unLockTXOutput(address) {
				if len(txInputs) == 0 {
					txOutputs = append(txOutputs, UTXO{
						Hash:  ts.TxHash,
						Index: index,
						Vout:  ou,
					})
				} else {
					for txHash, indexArray := range txInputs {
						txHashStr := hex.EncodeToString(ts.TxHash)
						if txHash == txHashStr {
							var isUTXO bool
							for _, v := range indexArray {
								if v == index {
									isUTXO = true
									continue tsVout
								}

								if !isUTXO {
									txOutputs = append(txOutputs, UTXO{
										Hash:  ts.TxHash,
										Index: index,
										Vout:  ou,
									})
								}
							}
						} else {
							txOutputs = append(txOutputs, UTXO{
								Hash:  ts.TxHash,
								Index: index,
								Vout:  ou,
							})
						}
					}
				}
			}
		}
	}

	bci := NewBlockchainIterator(blc)
	for {
		block := bci.Next()
		for i := len(block.Transactions) - 1; i >= 0; i-- {
			ts := block.Transactions[i]
			//for _, ts := range block.Transactions {
			if len(block.PreHash) != 0 {
				for _, in := range ts.Vint {
					//用户花费UTXO
					if in.unLockTXInput(Ripemd160Hash(address)) {
						key := hex.EncodeToString(in.TxHash)
						txInputs[key] = append(txInputs[key], in.Index)
					}
				}
			}
		Vout:
			for index, ou := range ts.Vout {
				if ou.unLockTXOutput(address) {
					if txInputs != nil {
						if len(txInputs) != 0 {
							var isUTXO = false
							for txHash, indexArray := range txInputs {
								for _, v := range indexArray {
									if v == index && txHash == hex.EncodeToString(ts.TxHash) {
										isUTXO = true
										continue Vout
									}
								}
							}
							if !isUTXO {
								txOutputs = append(txOutputs, UTXO{
									Hash:  ts.TxHash,
									Index: index,
									Vout:  ou,
								})
							}
						} else {
							txOutputs = append(txOutputs, UTXO{
								Hash:  ts.TxHash,
								Index: index,
								Vout:  ou,
							})
						}
					}
				}
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PreHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return txOutputs
}

//传入地址 返回地址余额信息
func (blc *Blockchain) GetBalance(address string) int {
	var balance int
	uHandle := UTXOHandle{blc}
	utxos := uHandle.findUTXOFromAddress(address)
	for _, v := range utxos {
		balance += v.Vout.Value
	}
	return balance
}

//查找数据库中全部未花费的UTXO
func (blc *Blockchain) findAllUTXOs() map[string]*TXOuputs {
	bci := NewBlockchainIterator(blc)

	// 存储已花费的UTXO的信息
	spentableUTXOsMap := make(map[string][]*TXInput)

	// 存储未花费的UTXO的信息
	utxoMaps := make(map[string]*TXOuputs)

	for {
		block := bci.Next()

		for i := len(block.Transactions) - 1; i >= 0; i-- {
			tXOuputs := &TXOuputs{[]*UTXO{}}

			tx := block.Transactions[i]

			// 判断是否为创世块
			if !tx.IsCoinbase() {
				for _, txInput := range tx.Vint {
					txHash := hex.EncodeToString(txInput.TxHash)
					spentableUTXOsMap[txHash] = append(spentableUTXOsMap[txHash], &txInput)
				}
			}

			txHash := hex.EncodeToString(tx.TxHash)

		WorkOutLoop:
			for index, out := range tx.Vout {

				tXInputs := spentableUTXOsMap[txHash]

				if len(tXInputs) > 0 {

					isSpent := false

					for _, in := range tXInputs {
						outPublicKey := out.PublicKeyHash
						inPublicKey := in.PublicKey

						if bytes.Compare(outPublicKey, PublicKeyHash(inPublicKey)) == 0 {
							if index == in.Index {
								isSpent = true
								continue WorkOutLoop
							}
						}
					}
					if !isSpent {
						utxo := &UTXO{
							Hash:  tx.TxHash,
							Index: index,
							Vout:  out,
						}
						tXOuputs.UTXOS = append(tXOuputs.UTXOS, utxo)
					}
				} else {
					utxo := &UTXO{
						Hash:  tx.TxHash,
						Index: index,
						Vout:  out,
					}
					tXOuputs.UTXOS = append(tXOuputs.UTXOS, utxo)
				}
			}
			// 设置键值对
			utxoMaps[txHash] = tXOuputs
		}

		// 找到创世区块时推出
		var hashInt big.Int
		hashInt.SetBytes(block.PreHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return utxoMaps
}

//打印区块链详细信息
func (blc *Blockchain) PrintAllBlockInfo() {
	blcIterator := NewBlockchainIterator(blc)
	for {
		block := blcIterator.Next()
		if block == nil {
			log.Error("还未生成创世区块!")
			return
		}
		fmt.Println("========================================================================================================")
		fmt.Printf("本块hash         %x\n", block.Hash)
		fmt.Println("  	------------------------------交易数据------------------------------")
		for _, v := range block.Transactions {
			fmt.Printf("   	 本次交易id:  %x\n", v.TxHash)
			fmt.Println("   	  tx_input：")
			for _, vIn := range v.Vint {
				fmt.Printf("			交易id:  %x\n", vIn.TxHash)
				fmt.Printf("			索引:    %d\n", vIn.Index)
				fmt.Printf("			签名信息:    %x\n", vIn.Signature)
			}
			fmt.Println("  	  tx_output：")
			for index, vOut := range v.Vout {
				fmt.Printf("			金额:    %d    \n", vOut.Value)
				fmt.Printf("			地址:    %s\n", vOut.PublicKeyHash)
				if len(v.Vout) != 1 && index != len(v.Vout)-1 {
					fmt.Println("			---------------")
				}
			}
		}
		fmt.Println("  	--------------------------------------------------------------------")
		fmt.Printf("时间戳           %s\n", time.Unix(block.TimeStamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("区块高度         %d\n", block.Height)
		fmt.Printf("随机数           %d\n", block.Nonce)
		fmt.Printf("上一个块hash     %x\n", block.PreHash)
		var hashInt big.Int
		hashInt.SetBytes(block.PreHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	fmt.Println("========================================================================================================")
}

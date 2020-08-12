package blc

import (
	"encoding/hex"
	"encoding/json"
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
	txi := TXInput{[]byte{}, -1, ""}

	txo := TXOutput{value, address}

	ts := Transaction{nil, []TXInput{txi}, []TXOutput{txo}}

	ts.hash()

	tss := []Transaction{ts}
	//开始生成区块链的第一个区块
	blc.newGenesisBlockchain(tss)
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
	newTXInput := []TXInput{}
	newTXOutput := []TXOutput{}

	for k, address := range fromSlice {
		utxos := blc.UTXOs(address)
		balance, indexs := blc.doUTXOs(utxos, amountSlice[k])
		for txHash, indexArray := range indexs {
			txHashBytes, _ := hex.DecodeString(txHash)
			for _, index := range indexArray {
				tXInput := TXInput{
					TxHash:    txHashBytes,
					Index:     index,
					Signature: address,
				}
				newTXInput = append(newTXInput, tXInput)
			}
		}
		tXOutput := TXOutput{amountSlice[k], toSlice[k]}
		newTXOutput = append(newTXOutput, tXOutput)
		//打包交易的核心操作
		tXOutput = TXOutput{balance - amountSlice[k], address}
		newTXOutput = append(newTXOutput, tXOutput)
	}
	ts := Transaction{nil, newTXInput, newTXOutput}
	ts.hash()
	tss = append(tss, ts)
	blc.addBlockchain(tss)
}

func (blc *Blockchain) doUTXOs(utxos []UTXO, amount int) (balance int, indexs map[string][]int) {
	indexs = make(map[string][]int)
	for _, utxo := range utxos {
		balance = balance + utxo.Vout.Value
		key := hex.EncodeToString(utxo.Hash)
		indexs[key] = append(indexs[key], utxo.Index)
	}
	if amount > balance {
		fmt.Println("余额不足！")
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
	//进行挖矿
	newBlock := NewBlock(transaction, blc.BD.View([]byte(LastBlockHashMapping), database.BlockBucket), height)

	//将区块添加到本地库中
	blc.AddBlock(newBlock)
}

//用户未花费
func (blc *Blockchain) UTXOs(address string) []UTXO {
	var txOutputs []UTXO
	txInputs := make(map[string][]int)
	bci := NewBlockchainIterator(blc)
	for {
		block := bci.Next()
		for _, ts := range block.Transactions {
			if len(block.PreHash) != 0 {
				for _, in := range ts.Vint {
					//用户花费UTXO
					if in.Signature == address {
						key := hex.EncodeToString(in.TxHash)
						txInputs[key] = append(txInputs[key], in.Index)
					}
				}
			}
		Vout:
			for index, ou := range ts.Vout {
				if ou.PublicKeyHash == address {
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
	utxos := blc.UTXOs(address)
	for _, utxo := range utxos {
		balance += utxo.Vout.Value
	}
	return balance
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

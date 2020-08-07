package blc

import (
	"encoding/json"
	"fmt"
	"math/big"
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

	txo := TXOutput{value, []byte("genesisBlock")}

	ts := Transaction{nil, []TXInput{txi}, []TXOutput{txo}}

	ts.hash()

	tss := []Transaction{ts}
	//开始生成区块链的第一个区块
	blc.newGenesisBlockchain(tss)

	fmt.Println("已成生成创世区块")
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

//增加区块到区块链里面
func (blc *Blockchain) CreateTransaction(from, to string, amount string) {
	//判断一下是否已生成创世区块
	if len(blc.BD.View([]byte(LastBlockHashMapping), database.BlockBucket)) == 0 {
		log.Error("还没有生成创世区块，不可进行转账操作 !")
		return
	}

	fromSlice := []string{}
	toSlice := []string{}
	amountSlice := []int{}

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

	for i, _ := range fromSlice {
		if i < len(fromSlice)-1 {
			fromSlice = append(fromSlice[:i], fromSlice[i+1:]...)
			toSlice = append(toSlice[:i], toSlice[i+1:]...)
			amountSlice = append(amountSlice[:i], amountSlice[i+1:]...)
		} else {
			fromSlice = append(fromSlice[:i])
			toSlice = append(toSlice[:i])
			amountSlice = append(amountSlice[:i])
		}
	}

	for i, _ := range toSlice {
		if i < len(fromSlice)-1 {
			fromSlice = append(fromSlice[:i], fromSlice[i+1:]...)
			toSlice = append(toSlice[:i], toSlice[i+1:]...)
			amountSlice = append(amountSlice[:i], amountSlice[i+1:]...)
		} else {
			fromSlice = append(fromSlice[:i])
			toSlice = append(toSlice[:i])
			amountSlice = append(amountSlice[:i])
		}
	}

	for i, v := range amountSlice {
		if v < 0 {
			log.Error("转账金额不可小于0，已将此笔交易剔除")
			if i < len(fromSlice)-1 {
				fromSlice = append(fromSlice[:i], fromSlice[i+1:]...)
				toSlice = append(toSlice[:i], toSlice[i+1:]...)
				amountSlice = append(amountSlice[:i], amountSlice[i+1:]...)
			} else {
				fromSlice = append(fromSlice[:i])
				toSlice = append(toSlice[:i])
				amountSlice = append(amountSlice[:i])
			}
		}
	}

	var tss []Transaction

	//打包交易的核心操作
	newTXInput := []TXInput{}
	newTXOutput := []TXOutput{}

	ts := Transaction{nil, newTXInput, newTXOutput[:]}
	ts.hash()
	tss = append(tss, ts)
	//blc.Blocks = append(blc.Blocks, newBlock)
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
				fmt.Printf("			公钥:    %x\n", vIn.PublicKey)
				fmt.Printf("			地址:    %s\n", vIn.PublicKey)
			}
			fmt.Println("  	  tx_output：")
			for index, vOut := range v.Vout {
				fmt.Printf("			金额:    %d    \n", vOut.Value)
				fmt.Printf("			公钥Hash:    %x    \n", vOut.PublicKeyHash)
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

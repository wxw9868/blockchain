package blc

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
)

type Wallet struct {
	Wallets map[string]*bitcoinKey
}

func NewWallet() (*Wallet, error) {
	wallet := new(Wallet)
	wallet.Wallets = make(map[string]*bitcoinKey)
	err := wallet.LoadFromFile()
	return wallet, err
}

func (w *Wallet) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallet Wallet
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallet)
	if err != nil {
		log.Panic(err)
	}
	w.Wallets = wallet.Wallets

	return nil
}

func (w *Wallet) CreateWallet() {
	bitcoinKey := NewBitcoinKey()
	address := bitcoinKey.GetAddressFromPublicKey()
	w.Wallets[address] = bitcoinKey
	//保存所有数据
	w.SaveToFile()
}

//将钱包信息存储到文件
func (w *Wallet) SaveToFile() {
	var content bytes.Buffer

	//注册的目的是为了可以序列化任何类型
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(w)
	if err != nil {
		log.Panic(err)
	}

	//将序列化以后的数据写入到文件，原来文件的数据会被覆盖
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

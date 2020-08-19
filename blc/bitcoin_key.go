package blc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"wxw-blockchain/util"

	log "github.com/corgi-kx/logcustom"
	"golang.org/x/crypto/ripemd160"
)

type bitcoinKey struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
}

//创建公私钥实例
func NewBitcoinKey() *bitcoinKey {
	b := new(bitcoinKey)
	b.newKeyPair()
	return b
}

//生成公私钥对
func (b *bitcoinKey) newKeyPair() {
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	b.PrivateKey = privateKey
	b.PublicKey = append(b.PrivateKey.PublicKey.X.Bytes(), b.PrivateKey.PublicKey.Y.Bytes()...)
}

//根据公钥生成账户地址
func (b *bitcoinKey) getAddress() []byte {
	//1.ripemd160(sha256(publickey)
	publicKeyHash := PublicKeyHash(b.PublicKey)
	//2.最前面添加一个字节的版本信息 versionPublickeyHash
	versionPublicKeyHash := append([]byte{version}, publicKeyHash...)
	//3.sha256(sha256(versionPublickeyHash))  取前四个字节的值
	checkSumHash := checkSumHash(versionPublicKeyHash)
	//4.拼接最终hash versionPublickeyHash + checksumHash
	hash := append(versionPublicKeyHash, checkSumHash...)
	//进行base58加密
	address := util.Base58Encode(hash)
	return address
}

//ripemd160(sha256(publickey)
func PublicKeyHash(publicKey []byte) []byte {
	sha256publicKey := sha256.Sum256(publicKey)
	r := ripemd160.New()
	r.Reset()
	r.Write(sha256publicKey[:])
	ripPublicKey := r.Sum(nil)
	return ripPublicKey
}

func checkSumHash(versionPublicKeyHash []byte) []byte {
	checksumHash := sha256.Sum256(versionPublicKeyHash)
	checksumHash = sha256.Sum256(checksumHash[:])
	return checksumHash[:checksum]
}

//判断是否是有效的比特币地址
func IsVaildBitcoinAddress(address string) bool {
	addressByte := []byte(address)
	hash := util.Base58Decode(addressByte)
	if len(hash) != 25 {
		return false
	}
	versionPublicKeyHash := hash[:len(hash)-checksum]
	checkSumHash1 := hash[len(hash)-checksum:]
	checkSumHash2 := checkSumHash(versionPublicKeyHash)
	if bytes.Compare(checkSumHash1, checkSumHash2[:]) == 0 {
		return true
	} else {
		return false
	}
}

//获取ripemd160
func Ripemd160Hash(address string) []byte {
	addressByte := []byte(address)
	hash := util.Base58Decode(addressByte)
	if len(hash) != 25 {
		return nil
	}
	return hash[1 : len(hash)-checksum]
}

//通过公钥获取地址
func (b *bitcoinKey) GetAddressFromPublicKey() string {
	if b.PublicKey == nil {
		return ""
	}
	return string(b.getAddress())
}

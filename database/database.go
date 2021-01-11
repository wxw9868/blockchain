package database

import (
	"github.com/boltdb/bolt"
	log "github.com/corgi-kx/logcustom"
)

var ListenPort = "9000"

// 仓库类型
type BucketType string

const (
	BlockBucket BucketType = "blocks"
	//AddrBucket  BucketType = "address"
	UTXOBucket BucketType = "utxo"
)

type BlockchainDB struct {
	ListenPort string
}

func New() *BlockchainDB {
	bd := &BlockchainDB{ListenPort}
	return bd
}

//判断仓库是否存在
func IsBucketExist(bt BucketType) bool {
	var isBucketExist bool

	var DBFileName = "blockchain_" + ListenPort + ".db"
	db, err := bolt.Open(DBFileName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bt))
		if bucket == nil {
			isBucketExist = false
		} else {
			isBucketExist = true
		}
		return nil
	})
	if err != nil {
		log.Panic("datebase IsBucketExist err:", err)
	}

	err = db.Close()
	if err != nil {
		log.Panic("db close err :", err)
	}
	return isBucketExist
}

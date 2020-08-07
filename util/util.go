package util

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	log "github.com/corgi-kx/logcustom"
	"math/big"
)

//int64转换成字节数组
func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

////int64转换成字节数组
//func Int64ToBytes(num int64) []byte {
//	buff := new(bytes.Buffer)
//	err := binary.Write(buff,binary.BigEndian,num)
//	if err != nil {
//		log.Panic(err)
//	}
//	return buff.Bytes()
//}

//生成随机数
func GenerateRealRandom() int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000000000000000))
	if err != nil {
		fmt.Println(err)
	}
	return n.Int64()
}

func JsonToArray(data string) (arraySlice []string) {
	//对传入的信息进行校验检测
	err := json.Unmarshal([]byte(data), &arraySlice)
	if err != nil {
		log.Error("json err:", err)
	}
	return
}

package blc

import "math"

//当前本地监听端口
var ListenPort string = "9000"

//挖矿难度值
var TargetBits uint = 16

//随机数不能超过的最大值
const maxInt = math.MaxInt64

//最新区块Hash在数据库中的键
const LastBlockHashMapping = "lastHash"

const version = byte(0x00)

const checksum = 4

const walletFile = "walletFile.bat"

wxw-blockchain createWallet

wxw-blockchain printAddressList

wxw-blockchain genesis -address "13HWMdJDDATndNmKKcmZzimDenXDy2R99v" -value "100"

wxw-blockchain printChain

wxw-blockchain transfer -from [\"13HWMdJDDATndNmKKcmZzimDenXDy2R99v\"] -to [\"1Hw7UhTaFcETESbad3tpKc6XvWHaJDjeNQ\"] -amount [10]

wxw-blockchain transfer -from [\"aaa\"] -to [\"bbb\"] -amount [4]

wxw-blockchain transfer -from [\"wxw\"] -to [\"bbb\"] -amount [10]

wxw-blockchain transfer -from [\"wxw\",\"wxw\"] -to [\"bbb\",\"aaa\"] -amount [10,10]

wxw-blockchain getBalance -address 13HWMdJDDATndNmKKcmZzimDenXDy2R99v

wxw-blockchain getBalance -address 1Hw7UhTaFcETESbad3tpKc6XvWHaJDjeNQ

https://www.upantool.com/jiaocheng/hdd/7252.html


wxw-blockchain transfer -from [\"liyuechun\",\"juncheng\"] -to [\"juncheng\",\"liyuechun\"] -amount [3,4]
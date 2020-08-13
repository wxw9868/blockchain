wxw-blockchain genesis -address "wxw" -value "100"

wxw-blockchain printChain

wxw-blockchain transfer -from [\"wxw\"] -to [\"aaa\"] -amount [10]

wxw-blockchain transfer -from [\"aaa\"] -to [\"bbb\"] -amount [4]

wxw-blockchain transfer -from [\"wxw\"] -to [\"bbb\"] -amount [10]

wxw-blockchain transfer -from [\"wxw\",\"wxw\"] -to [\"bbb\",\"aaa\"] -amount [10,10]

wxw-blockchain getBalance -address bbb

https://www.upantool.com/jiaocheng/hdd/7252.html


wxw-blockchain transfer -from [\"liyuechun\",\"juncheng\"] -to [\"juncheng\",\"liyuechun\"] -amount [3,4]
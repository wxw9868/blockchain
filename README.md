wxw-blockchain createWallet

wxw-blockchain printAddressList

wxw-blockchain genesis -address "1LrA2EwCjm3YDhwKcxvCK1bndTrxrdWMhc" -value "100"

wxw-blockchain printChain

wxw-blockchain transfer -from [\"1LrA2EwCjm3YDhwKcxvCK1bndTrxrdWMhc\"] -to [\"14uK5GbYwo9CRegibsfYqepca3UjoXomB1\"] -amount [10]

wxw-blockchain transfer -from [\"1LrA2EwCjm3YDhwKcxvCK1bndTrxrdWMhc\",\"14uK5GbYwo9CRegibsfYqepca3UjoXomB1\"] -to [\"14uK5GbYwo9CRegibsfYqepca3UjoXomB1\",\"1LrA2EwCjm3YDhwKcxvCK1bndTrxrdWMhc\"] -amount [15,10]

wxw-blockchain resetUTXODB

wxw-blockchain getBalance -address 1LrA2EwCjm3YDhwKcxvCK1bndTrxrdWMhc

wxw-blockchain getBalance -address 14uK5GbYwo9CRegibsfYqepca3UjoXomB1

https://www.upantool.com/jiaocheng/hdd/7252.html


wxw-blockchain transfer -from [\"liyuechun\",\"juncheng\"] -to [\"juncheng\",\"liyuechun\"] -amount [3,4]
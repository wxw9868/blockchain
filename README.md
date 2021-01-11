###操作命令

````
wxw-blockchain createWallet

wxw-blockchain printAddressList

wxw-blockchain genesis -address "17xj49JPUidTy1ZHmsoeq1hMb2ZvPmXyGB" -value "100"

wxw-blockchain printChain

wxw-blockchain getBalance -address 17xj49JPUidTy1ZHmsoeq1hMb2ZvPmXyGB

wxw-blockchain transfer -from [\"17xj49JPUidTy1ZHmsoeq1hMb2ZvPmXyGB\"] -to [\"1Fe9N9tPpKFwrpp6KhP1SR71na2oe96Z65\"] -amount [10]

wxw-blockchain transfer -from [\"1LrA2EwCjm3YDhwKcxvCK1bndTrxrdWMhc\",\"14uK5GbYwo9CRegibsfYqepca3UjoXomB1\"] -to [\"14uK5GbYwo9CRegibsfYqepca3UjoXomB1\",\"1LrA2EwCjm3YDhwKcxvCK1bndTrxrdWMhc\"] -amount [15,10]

wxw-blockchain resetUTXODB
````
wxw-blockchain genesis -address "wxw" -value "100"

wxw-blockchain printChain

wxw-blockchain transfer -from [\"wxw\"] -to [\"aaa\"] -amount [10]

wxw-blockchain transfer -from [\"aaa\"] -to [\"bbb\"] -amount [10]

wxw-blockchain transfer -from [\"wxw001\",\"wxw001\"] -to [\"bbb\",\"ccc\"] -amount [10,10]


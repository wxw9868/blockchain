package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"wxw-blockchain/util"

	"wxw-blockchain/blc"

	log "github.com/corgi-kx/logcustom"
)

type CLI struct {
}

func printUsage() {
	fmt.Println("----------------------------------------------------------------------------- ")
	fmt.Println("Usage:")
	fmt.Println("\tprintAddressList                                   查看所有钱包地址")
	fmt.Println("\tcreateWallet                                      创建钱包")
	fmt.Println("\tgenesis -address DATA -value DATA                 生成创世区块")
	fmt.Println("\tgetBalance -address DATA                          查看用户余额")
	fmt.Println("\ttransfer -from DATA -to DATA -amount DATA         进行转账操作")
	fmt.Println("\tprintChain                                        查看所有区块信息")
	fmt.Println("------------------------------------------------------------------------------")
}

func (cli *CLI) Run() {
	printAddressListCmd := flag.NewFlagSet("printAddressList", flag.ExitOnError)
	genesisCmd := flag.NewFlagSet("genesis", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	transferCmd := flag.NewFlagSet("transfer", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)

	flagGenesisAddress := genesisCmd.String("address", "", "地址")
	flagGenesisValue := genesisCmd.String("value", "", "金额")

	//flagCreateWallet := createWalletCmd.String("address", "", "地址")

	flagBalance := getBalanceCmd.String("address", "", "地址")

	flagFrom := transferCmd.String("from", "", "付款地址")
	flagTo := transferCmd.String("to", "", "收款地址")

	flagAmount := transferCmd.String("amount", "", "交易金额")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "printAddressList":
		err := printAddressListCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createWallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "genesis":
		err := genesisCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getBalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "transfer":
		err := transferCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printChain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		printUsage()
		os.Exit(1)
	}

	if printAddressListCmd.Parsed() {
		//if *flagCreateWallet == "" {
		//	printUsage()
		//	os.Exit(1)
		//}
		cli.printAddressList()
	}

	if createWalletCmd.Parsed() {
		//if *flagCreateWallet == "" {
		//	printUsage()
		//	os.Exit(1)
		//}
		cli.createWallet()
	}

	if genesisCmd.Parsed() {
		if !blc.IsVaildBitcoinAddress(*flagGenesisAddress) {
			log.Error("地址格式错误")
			os.Exit(1)
		}
		if *flagGenesisValue == "" {
			printUsage()
			os.Exit(1)
		}
		v, err := strconv.Atoi(*flagGenesisValue)
		if err != nil {
			log.Fatal(err)
		}
		cli.genesis(*flagGenesisAddress, v)
	}

	if getBalanceCmd.Parsed() {
		if *flagBalance == "" {
			printUsage()
			os.Exit(1)
		}
		cli.getBalance(*flagBalance)
	}

	if transferCmd.Parsed() {
		fromSlice := util.JsonToArray(*flagFrom)
		toSlice := util.JsonToArray(*flagTo)
		for k, v := range fromSlice {
			if !blc.IsVaildBitcoinAddress(v) || !blc.IsVaildBitcoinAddress(toSlice[k]) {
				log.Error("地址格式错误")
				os.Exit(1)
			}
		}
		if *flagAmount == "" {
			printUsage()
			os.Exit(1)
		}
		cli.transfer(*flagFrom, *flagTo, *flagAmount)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

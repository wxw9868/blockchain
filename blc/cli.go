package blc

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct {
}

func printUsage() {
	fmt.Println("----------------------------------------------------------------------------- ")
	fmt.Println("Usage:")
	fmt.Println("\tgenesis -address DATA -value DATA                 生成创世区块")
	fmt.Println("\ttransfer -from DATA -to DATA -amount DATA         进行转账操作")
	fmt.Println("\tprintChain                                        查看所有区块信息")
	fmt.Println("------------------------------------------------------------------------------")
}

func (cli *CLI) genesis(address string, value int) {
	blc := NewBlockchain()
	blc.CreataGenesisTransaction(address, value)
}

func (cli *CLI) transfer(from, to string, amount string) {
	blc := NewBlockchain()
	blc.CreateTransaction(from, to, amount)
}

func (cli *CLI) printChain() {
	blc := NewBlockchain()
	blc.PrintAllBlockInfo()
}

func (cli *CLI) Run() {
	genesisCmd := flag.NewFlagSet("genesis", flag.ExitOnError)
	transferCmd := flag.NewFlagSet("transfer", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)

	flagGenesisAddress := genesisCmd.String("address", "", "地址")
	flagGenesisValue := genesisCmd.String("value", "", "金额")

	flagTransferFrom := transferCmd.String("from", "", "付款地址")
	flagTransferTo := transferCmd.String("to", "", "收款地址")
	flagTransferAmount := transferCmd.String("amount", "", "交易金额")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "genesis":
		err := genesisCmd.Parse(os.Args[2:])
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

	if genesisCmd.Parsed() {
		if *flagGenesisAddress == "" || *flagGenesisValue == "" {
			printUsage()
			os.Exit(1)
		}
		v, err := strconv.Atoi(*flagGenesisValue)
		if err != nil {
			log.Fatal(err)
		}
		cli.genesis(*flagGenesisAddress, v)
	}

	if transferCmd.Parsed() {
		if *flagTransferFrom == "" || *flagTransferTo == "" || *flagTransferAmount == "" {
			printUsage()
			os.Exit(1)
		}
		cli.transfer(*flagTransferFrom, *flagTransferTo, *flagTransferAmount)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

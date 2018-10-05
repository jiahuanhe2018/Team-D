package main

import (
	"Course/console"
	"Course/p2p"
	"flag"
	"fmt"
	"log"

	"Course/blockchain"
	"Course/rpc"

	"Course/wallet"
	golog "github.com/ipfs/go-log"
	gologging "github.com/whyrusleeping/go-logging"
)

func main() {

	// Parse options from the command line
	command  := flag.String("c", "", "mode[ \"chain\" or \"account\"]")
	consensus  := flag.String("consensus", "pow", "mode[ \"pow\" or \"pos\"]")
	listenF := flag.Int("l", 0, "wait for incoming connections[chain mode param]")
	target := flag.String("d", "", "target peer to dial[chain mode param]")
	suffix := flag.String("s", "", "wallet suffix [chain mode param]")
	initAccounts := flag.String("a", "", "init exist accounts whit value 10000")
	secio := flag.Bool("secio", false, "enable secio[chain mode param]")
	seed := flag.Int64("seed", 0, "set random seed for id generation[chain mode param]")
	flag.Parse()


	if *command == "chain" {
		runblockchain(listenF, consensus, target, seed, secio, suffix, initAccounts)
	}else if *command == "account" {
		cli := wallet.WalletCli{}
		cli.Run()
	}else {
		flag.Usage()
	}
}

func runblockchain(listenF *int, consensus *string, target *string, seed *int64, secio *bool, suffix *string, initAccounts *string){

	defaultAccounts := make(map[string]uint64)
	if *initAccounts != ""{
		if wallet.ValidateAddress(*initAccounts) == false {
			fmt.Println("Invalid address")
			return
		}
		defaultAccounts[*initAccounts] = 10000
	}
	genesisBlock := blockchain.GenerateGenesisBlock(defaultAccounts)

	var blocks []blockchain.Block
	blocks = append(blocks, genesisBlock)
	blockchain.BlockchainInstance.Blocks =  blocks

	blockchain.Consensus = *consensus
	blockchain.IsMaster = *target == ""

	// LibP2P code uses golog to log messages. They log with different
	// string IDs (i.e. "swarm"). We can control the verbosity level for
	// all loggers with:
	golog.SetAllLoggers(gologging.INFO) // Change to DEBUG for extra info

	if *listenF == 0 {
		log.Fatal("Please provide a peer port to bind on with -l")
	}

	if *suffix == "" {
		log.Println("option param -s miss [you can't send transacion with this node]")
	}else {
		blockchain.WalletSuffix = *suffix
	}

	go rpc.RunHttpServer(*listenF+1)
	go console.RunConsole()
	if blockchain.IsMaster {
		go blockchain.PickWinner()
	}
	p2p.RunNode(consensus, target, listenF, secio, seed)
}
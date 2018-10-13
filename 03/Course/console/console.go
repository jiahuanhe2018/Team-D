package console

import (
	"Course/blockchain"
	"Course/p2p"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"log"
	"os"
	"strconv"
	"strings"
)

func RunConsole(){
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		commandData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		commandData = strings.Replace(commandData, "\n", "", -1)
		if err != nil {
			log.Fatal(err)
		}
		commands := strings.Split(commandData, " ")
		command := commands[0]

		switch command {
		case "listblocks":
			CommandListBlocks()
		case "listaccounts":
			CommandListAccounts()
		case "send":
			CommandSend(commands)
		case "startmining":
			CommandStartMining()
		case "stopmining":
			CommandStopMining()
		default:
			fmt.Printf("no such command: %s\n", command)
		}

	}
}

func CommandListBlocks()  {

	bytes, err := json.MarshalIndent(blockchain.BlockchainInstance.GetAllBlocks(), "", "  ")
	if err != nil {

		log.Fatal(err)
	}
	fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
}

func CommandListAccounts()  {
	spew.Dump(blockchain.BlockchainInstance.Accounts)
}


func CommandSend(commands []string){
	if len(commands) < 4 {
		fmt.Println("Usage: send from to value")
		return
	}
	sender := commands[1]
	recipient := commands[2]
	amount, err := strconv.ParseUint(commands[3], 10, 64)
	if err != nil {
		fmt.Println("amount is invalid")
		return
	}
	t := blockchain.NewTransaction(sender, recipient, amount, nil)
	blockchain.BlockchainInstance.AddTxPool(t)
	p2p.NodeManagerInstance.BroadcastTransaction(t)
	fmt.Println("new transaction:", sender, recipient, amount)
}

func CommandStartMining()  {
	blockchain.StartMining()
}

func CommandStopMining()  {
	blockchain.StopMining()
}

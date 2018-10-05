package blockchain

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

var WalletSuffix string


// Blockchain is a series of validated Blocks
type Blockchain struct {
	Blocks []Block
	TxPool *TxPool
	mutex *sync.Mutex
}

func (t *Blockchain)AddTxPool(tx *Transaction) int {
	t.TxPool.AllTx = append(t.TxPool.AllTx, *tx)
	return len(t.TxPool.AllTx)
}

func (t *Blockchain) LastBlock() Block {
	return t.Blocks[len(t.Blocks)-1]
}

func (t *Blockchain) GetBalance(address string) uint64 {
	accounts := t.LastBlock().Accounts
	if value, ok := accounts[address]; ok {
		return value
	}
	return 0
}


func (t *Blockchain)PackageTx(newBlock *Block) {
	(*newBlock).Transactions = t.TxPool.AllTx
	AccountsMap := t.LastBlock().Accounts
	for k1, v1 := range AccountsMap {
		fmt.Println(k1, "--", v1)
	}
	unusedTx := make([]Transaction,0)
	for _, v := range t.TxPool.AllTx{
		if value, ok := AccountsMap[v.Sender]; ok {
			if value < v.Amount{
				unusedTx = append(unusedTx, v)
				continue
			}
			AccountsMap[v.Sender] = value-v.Amount
		}
		if value, ok := AccountsMap[v.Recipient]; ok {
			AccountsMap[v.Recipient] = value + v.Amount
		}else {
			AccountsMap[v.Recipient] = v.Amount
		}
	}
    t.TxPool.Clear()
    //余额不够的交易放回交易池
    if len(unusedTx) > 0 {
		for _, v := range unusedTx{
			t.AddTxPool(&v)
		}
	}
	(*newBlock).Accounts = AccountsMap
}



func (t *Blockchain) AppendBlock(newBlock *Block) {
	t.mutex.Lock()
	t.Blocks = append(t.Blocks, *newBlock)
	t.mutex.Unlock()
}

func (t *Blockchain) MarshalBlocks() ([]byte, error) {
	t.mutex.Lock()
	bytes, err := json.Marshal(t.Blocks)
	t.mutex.Unlock()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return bytes, nil
}

func (t *Blockchain) Lock() {
	t.mutex.Lock()
}

func (t *Blockchain) Unlock() {
	t.mutex.Unlock()
}

var BlockchainInstance = Blockchain{
	TxPool : NewTxPool(),
	mutex: &sync.Mutex{},
}


const difficulty = 1
var running = false
var Consensus = "pow"
var IsMaster = true
var ToBroadcastConfirmedtBlocks = make(chan Block)
var ToBroadcastCandidateBlocks = make(chan Block)
var CandidateBlocks = make([]Block, 0)

func StartMining()  {
	running = true

	switch Consensus {
	case "pow":
		go minePOW()
	case "pos":
		go minePOS()
	default:
		log.Println("eroor: consensus not supported!")
	}

}

func StopMining()  {
	running = false
}

func minePOW()  {
	for {
		if !running {
			break
		}
		newBlock := generateBlockPOW(BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1], 1)

		if len(BlockchainInstance.TxPool.AllTx) > 0 {
			BlockchainInstance.PackageTx(&newBlock)
		}else {
			newBlock.Accounts = BlockchainInstance.LastBlock().Accounts
		}

		if isBlockValid(newBlock, BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1]) {
			BlockchainInstance.AppendBlock(&newBlock)
			ToBroadcastConfirmedtBlocks <- newBlock
		}
	}
}

func minePOS()  {
	for {
		if !running {
			break
		}
		newBlock := generateBlockPOS(BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1], 1)

		if len(BlockchainInstance.TxPool.AllTx) > 0 {
			BlockchainInstance.PackageTx(&newBlock)
		}else {
			newBlock.Accounts = BlockchainInstance.LastBlock().Accounts
		}

		if isBlockValid(newBlock, BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1]) {
			if IsMaster {
				CandidateBlocks = append(CandidateBlocks, newBlock)
			}else {
				ToBroadcastCandidateBlocks <- newBlock
			}
		}

		time.Sleep(time.Second*5)
	}
}

func PickWinner() {
	for {
		time.Sleep(time.Second * 15)
		candidates := CandidateBlocks
		if len(candidates) == 0 {
			continue
		}
		r := rand.New(rand.NewSource(time.Now().Unix()))
		block := candidates[r.Intn(len(candidates))]
		BlockchainInstance.AppendBlock(&block)
		CandidateBlocks = []Block{}
	}
}
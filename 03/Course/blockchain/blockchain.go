package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"math/rand"
	"sync"
	"time"
)


const difficulty = 1

var (
	WalletSuffix string
	BlockchainInstance Blockchain
	running = false
	Consensus = "pow"
	IsMaster = true
	ToBroadcastConfirmedtBlocks = make(chan Block)
	ToBroadcastCandidateBlocks = make(chan Block)
	CandidateBlocks = make([]Block, 0)
	//AccountMap = make(map[string] Account)
	dbPath = "./blockchain.db"
)

// Blockchain is a series of validated Blocks
type Blockchain struct {
	DB *leveldb.DB
	Last string
	Accounts map[string] Account
	TxPool *TxPool
	sync.Mutex
}


type Account struct {
	Balance uint64 `json:"balance"`
	State   uint64 `json:"state"`
}


func (t *Blockchain)AddTxPool(tx *Transaction) int {
	t.TxPool.AllTx = append(t.TxPool.AllTx, *tx)
	return len(t.TxPool.AllTx)
}

func (t *Blockchain) LastBlock() *Block {
	data, _ := t.DB.Get([]byte(t.Last), nil)
	return DeserializeBlock(data)
}

func (t *Blockchain) GetBlock(hash string) *Block {
	data, _ := t.DB.Get([]byte(hash), nil)
	return DeserializeBlock(data)
}

func (t *Blockchain) GetAllBlocks() []Block {
	var l = make([]Block, 0)
	var hash = t.Last
	for {
		b := t.GetBlock(hash)
		l = append(l, *b)
		if b.PrevHash == "" {
			break
		}
		hash = b.PrevHash
	}
	return l
}

func (t *Blockchain) GetBalance(address string) uint64 {
	accounts := t.Accounts
	if value, ok := accounts[address]; ok {
		return value.Balance
	}
	return 0
}


func (t *Blockchain)PackageTx(newBlock *Block) {
	newBlock.Transactions = make([]Transaction, 0)
	AccountsMap := t.Accounts
	unusedTx := make([]Transaction,0)
	for _, v := range t.TxPool.AllTx{
		if account, ok := AccountsMap[v.Sender]; ok {
			if account.Balance < v.Amount{
				unusedTx = append(unusedTx, v)
				continue
			}
		}
		newBlock.Transactions = append(newBlock.Transactions, v)
	}
    t.TxPool.Clear()
    //余额不够的交易放回交易池
    if len(unusedTx) > 0 {
		for _, v := range unusedTx{
			t.AddTxPool(&v)
		}
	}
}

func (t *Blockchain)UpdateAccounts(ts []Transaction) {
	AccountsMap := t.Accounts
	for k1, v1 := range AccountsMap {
		fmt.Println(k1, "--", v1)
	}
	for _, v := range ts {
		if account, ok := AccountsMap[v.Sender]; ok {
			if account.Balance < v.Amount{
				log.Println("error no enough balance")
				continue
			}
			account.Balance -= v.Amount
			account.State += 1
			AccountsMap[v.Sender] = account
		}
		if account, ok := AccountsMap[v.Recipient]; ok {
			account.Balance += v.Amount
			account.State += 1
			AccountsMap[v.Recipient] = account
		}else {
			newAccount := new(Account)
			newAccount.Balance = v.Amount
			newAccount.State = 0
			AccountsMap[v.Recipient] = *newAccount
		}
	}
	t.Accounts = AccountsMap
}


func (t *Blockchain) AppendBlock(newBlock *Block) {
	t.Lock()
	println("append block")
	spew.Dump(newBlock)
	t.DB.Put([]byte(newBlock.Hash), newBlock.Serialize(), nil)
	t.DB.Put([]byte("last"), []byte(newBlock.Hash), nil)
	t.Last = newBlock.Hash
	t.UpdateAccounts(newBlock.Transactions)
	data, _ := json.Marshal(t.Accounts)
	t.DB.Put([]byte("accounts"), []byte(data), nil)
	t.Unlock()
}


func (t *Blockchain) MarshalBlocks() ([]byte, error) {
	t.Lock()
	bytes, err := json.Marshal(t.GetAllBlocks())
	t.Unlock()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return bytes, nil
}




func GetDB()(_db *leveldb.DB, err error){
	_db, err = leveldb.OpenFile(dbPath, nil)
	return
}


func InitBlockchain(initAccount string) error{
	var last string
	db, err := GetDB()
	if err != nil {
		return err
	}

	// load accounts
	accounts := make(map[string] Account)
	_accounts, err := db.Get([]byte("accounts"), nil)
	if err != nil {
		newAccount := Account{10000, 0}
		accounts[initAccount] = newAccount
		data, err := json.Marshal(accounts)
		if err != nil {
			return err
		}
		db.Put([]byte("accounts"), []byte(data), nil)
	}else{
		err = json.Unmarshal(_accounts, &accounts)
	}

	// load blocks
	_last, err := db.Get([]byte("last"), nil)
	if err != nil {
		//TODO: add account state
		b :=  GenerateGenesisBlock()
		db.Put([]byte(b.Hash), b.Serialize(), nil)
		db.Put([]byte("last"), []byte(b.Hash), nil)
		last = b.Hash
	}else{
		last = string(_last)
	}

	BlockchainInstance = Blockchain {
		TxPool : NewTxPool(),
		DB: db,
		Last: last,
		Accounts:accounts,
	}
	return nil
}



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
		newBlock := generateBlockPOW(*BlockchainInstance.LastBlock(), 1)

		if len(BlockchainInstance.TxPool.AllTx) > 0 {
			BlockchainInstance.PackageTx(&newBlock)
		}else {
			newBlock.AccountState = ""
		}

		if isBlockValid(newBlock, *BlockchainInstance.LastBlock()) {
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
		newBlock := generateBlockPOS(*BlockchainInstance.LastBlock(), 1)

		if len(BlockchainInstance.TxPool.AllTx) > 0 {
			BlockchainInstance.PackageTx(&newBlock)
		}else {
			newBlock.AccountState = ""
		}

		if isBlockValid(newBlock, *BlockchainInstance.LastBlock()) {
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
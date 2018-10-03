package blockchain

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	//"github.com/davecgh/go-spew/spew"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)
const (
   POW int =1
   POS int =2
)
 
var ConsensusType int
 
var WalletSuffix string

// Block represents each 'item' in the blockchain
type Block struct {
	Index     int `json:"index"`
	Timestamp string `json:"timestamp"`
	Result       int `json:"result"`
	Hash      string `json:"hash"`
	PrevHash  string `json:"prevhash"`
	Proof        uint64           `json:"proof"`
	Transactions []Transaction `json:"transactions"`
	Accounts   map[string]uint64  `json:"accounts"`
	Nonce      string `json:"nonce"`
	Validator  string `json:"Validator"`
}
// type Validator struct {
// 	Address     string `json:"address"`
// 	Balance     int `json:"balance"`
// }
// var Validators []Validator

type Transaction struct {
	Amount    uint64    `json:"amount"`
	Recipient string `json:"recipient"`
	Sender    string `json:"sender"`
	Data      []byte `json:"data"`
}

type TxPool struct {
	AllTx     []Transaction
}

func NewTxPool() *TxPool {
	return &TxPool{
		AllTx:   make([]Transaction, 0),
	}
}


func (p *TxPool)Clear() bool {
	if len(p.AllTx) == 0 {
		return true
	}
	p.AllTx = make([]Transaction, 0)
	return true
}

// Blockchain is a series of validated Blocks
type Blockchain struct {
	Blocks []Block
	TxPool *TxPool
}

func (t *Blockchain) NewTransaction(sender string, recipient string, amount uint64, data []byte) *Transaction {
	transaction := new(Transaction)
	transaction.Sender = sender
	transaction.Recipient = recipient
	transaction.Amount = amount
	transaction.Data = data

	return transaction
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

var BlockchainInstance Blockchain = Blockchain{
	TxPool : NewTxPool(),
}

var mutex = &sync.Mutex{}


func Lock(){
	mutex.Lock()
}

func UnLock(){
	mutex.Unlock()
}

// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It will use secio if secio is true.
func MakeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)

	 
	strflag:=""

	if ConsensusType==POW {
		strflag="pow"
	} else if ConsensusType==POS {
        strflag="pos"
	} else {
		strflag=""
    }

	if secio {
		log.Printf("Now run \"go run main.go -c chain%s -l %d -d %s -secio\" on a different terminal\n", strflag,listenPort+2, fullAddr)
	} else {
		log.Printf("Now run \"go run main.go -c chain%s -l %d -d %s\" on a different terminal\n", strflag,listenPort+2, fullAddr)
	}

	return basicHost, nil
}

func HandleStream(s net.Stream) {

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
    if ConsensusType==POW {
		go ReadData_Pow(rw)
		go WriteData_Pow(rw)
	} else if ConsensusType==POS {
        go ReadData_Pos(rw)
		go WriteData_Pos(rw)
	} else {
		go ReadData(rw)
		go WriteData(rw)
    }
	// stream 's' will stay open until you close it (or the other side closes it).
}
var candidateBlocks = make(chan Block)
var ValidatorsChan = make(chan map[string]int)
var tempBlocks []Block
var Validators = make(map[string]int)

type MiningBlock struct {
	Blocks []Block
}
var MiningBlockHash = make(map[string]Block)
func ReadData_Pow(rw *bufio.ReadWriter) {

	go func() {

		for {
			tempblock := <-candidateBlocks
			 
			mutex.Lock()
			 
			BlocksTemp:=BlockchainInstance
			mutex.Unlock()
			
			newblock:=generateBlock_Pow(BlocksTemp.Blocks[len(BlocksTemp.Blocks)-1],tempblock.Result)
            if len(BlocksTemp.TxPool.AllTx) > 0 {
				BlocksTemp.PackageTx(&newblock)
			}else {
				newblock.Accounts = BlocksTemp.LastBlock().Accounts
			}
			 
			if IsBlockValid(newblock, BlocksTemp.Blocks[len(BlocksTemp.Blocks)-1]) {
				 
				 
				BlocksTemp.Blocks = append(BlocksTemp.Blocks, newblock)
				 
			}
	
			bytes, err := json.Marshal(BlocksTemp.Blocks)
			if err != nil {
				log.Println(err)
			}
	
			//spew.Dump(BlockchainInstance.Blocks)
	
			mutex.Lock()
		 	BlocksTemp.Blocks=[]Block{}
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()
			fmt.Println("============")
			fmt.Print(">")
			 
		}
	 
	}()
		
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			
			mutex.Lock()
			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}
			if len(chain)==1 &&  chain[0].Hash=="" {
				//candidateBlocks <- chain[0]
				//bytes, err := json.MarshalIndent(chain, "", "  ")
				//if err != nil {
				//	log.Fatal(err)
				//}
				//fmt.Printf("print[%s]", string(bytes))

				hashstr:=CalculateHash(chain[0])
				_,ok := MiningBlockHash[hashstr]
				if !ok {
					MiningBlockHash[hashstr]=chain[0]
					candidateBlocks <- chain[0]
					rw.WriteString(fmt.Sprintf("%s\n", str))
					rw.Flush()
				} else {
					
				}
				
			} else if len(chain) > len(BlockchainInstance.Blocks) {
				BlockchainInstance.Blocks = chain
				bytes, err := json.MarshalIndent(BlockchainInstance.Blocks, "", "  ")
				if err != nil {
					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
				fmt.Println("Mining Complete")
				fmt.Print("> ")
			}
			mutex.Unlock()
		}
	}
}
func ReadData_Pos(rw *bufio.ReadWriter) {
	 
	go func() {
			for {
				  
					time.Sleep(15 * time.Second)
					mutex.Lock()
					temp := tempBlocks
					mutex.Unlock()

					lotteryPool := []string{}
					if len(temp) > 0 {

						// slightly modified traditional proof of stake algorithm
						// from all validators who submitted a block, weight them by the number of staked tokens
						// in traditional proof of stake, validators can participate without submitting a block to be forged
					OUTER:
						for _, block := range temp {
							// if already in lottery pool, skip
							for _, node := range lotteryPool {
								if block.Validator == node {
									continue OUTER
								}
							}

							// lock list of validators to prevent data race
							mutex.Lock()
							setValidators :=Validators
							mutex.Unlock()

							k, ok := setValidators[block.Validator]
							if ok {
								for i := 0; i < k; i++ {
									lotteryPool = append(lotteryPool, block.Validator)
								}
							}
						}

						// randomly pick winner from lottery pool
						
						// s := rand.NewSource(time.Now().Unix())
						// r := rand.New(s)
						// lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]
						 //lotteryWinner := lotteryPool[len(lotteryPool)-1]
					
						 //s:=mrand.NewSource(time.Now().Unix())
						 //r:=mrand.New(s)
						 mrand.Seed(time.Now().Unix())
						 lotteryWinner := lotteryPool[mrand.Intn(len(lotteryPool))]

						// add block of winner to blockchain and let all the other nodes know
						for _, block := range temp {
							if block.Validator == lotteryWinner {
								mutex.Lock()
							    BlocksTemp:=BlockchainInstance
			 
								newblock,_:=generateBlock(BlocksTemp.Blocks[len(BlocksTemp.Blocks)-1],block.Result,block.Validator)
								BlocksTemp.Blocks = append(BlocksTemp.Blocks, newblock)
								mutex.Unlock()
								
								bytes, err := json.Marshal(BlocksTemp.Blocks)
								if err != nil {
									log.Println(err)
								}
								rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
								rw.Flush()
								break
							}
						}
					}

					mutex.Lock()
					tempBlocks = []Block{}
					mutex.Unlock()
			}
	}()
	 
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			
			
			chain := make([]Block, 0)
			validator:=make(map[string]int)
			 
			if err := json.Unmarshal([]byte(str), &validator); err != nil {
					
			} else {
				 
				if len(validator) > len(Validators) {
					mutex.Lock() 
					Validators=validator
					//fmt.Printf("%s", str)
					mutex.Unlock()
				}
				continue
			}
			 
			
			mutex.Lock()
            if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}
			if len(chain)==1 &&  chain[0].Hash=="" {
				//print("begin minnering"+string(chain[0].Index))
				//candidateBlocks <- chain[0]

				hashstr:=CalculateHash(chain[0])
				_,ok := MiningBlockHash[hashstr]
				if !ok {
					MiningBlockHash[hashstr]=chain[0]
					 
					tempBlocks = append(tempBlocks, chain[0])
					 
					rw.WriteString(fmt.Sprintf("%s\n", str))
					rw.Flush()
				} else {
					
				}

			} else if len(chain) > len(BlockchainInstance.Blocks) {
				BlockchainInstance.Blocks = chain
				bytes, err := json.MarshalIndent(BlockchainInstance.Blocks, "", "  ")
				if err != nil {
					log.Fatal(err)
				}
				tempBlocks=[]Block{}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
				fmt.Printf("\nwinning complete.validator=%s",BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1].Validator)
				fmt.Print("\nEnter a new Result> ")
			}
			
			mutex.Unlock()
		}
	}
}
func ReadData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {

			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			if len(chain) > len(BlockchainInstance.Blocks) {
				BlockchainInstance.Blocks = chain
				bytes, err := json.MarshalIndent(BlockchainInstance.Blocks, "", "  ")
				if err != nil {
					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}
}
func WriteData_Pow(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(BlockchainInstance.Blocks)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()
	 

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		sendData = strings.Replace(sendData, "\n", "", -1)
		sendData = strings.Replace(sendData, "\r", "", -1) //suncj
		_result, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}
		chain := make([]Block, 0)
		mutex.Lock()
		newBlock := generateNewBlock(BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1], _result)
		chain=append(chain,newBlock)
		bytes, err := json.Marshal(chain)
		if err != nil {
			log.Println(err)
		}
        //print(string(bytes))
		//MiningBlockInstance.Blocks=append(MiningBlockInstance.Blocks,newBlock)
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		
		//candidateBlocks <- newBlock
		mutex.Unlock()

		 
	}

}
func WriteData_Pos(rw *bufio.ReadWriter) {
	
	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(BlockchainInstance.Blocks)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()
    go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(Validators)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()
	stdReader := bufio.NewReader(os.Stdin)
	var address string
	//var validators = make(map[string]int)
	 
		 
			fmt.Print("Enter token balance: ")
			sendData, err := stdReader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}

			sendData = strings.Replace(sendData, "\n", "", -1)
			sendData = strings.Replace(sendData, "\r", "", -1) //suncj
			_result, err := strconv.Atoi(sendData)
			if err != nil {
				log.Fatal(err)
			}
            mutex.Lock()
			t := time.Now()
			address= calculateHash(t.String())
			Validators[address] = _result
			bytes, err := json.Marshal(Validators)
			if err != nil {
				log.Println(err)
			}
			
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()
			
			for {
					fmt.Print("Enter a new Result: ")
					sendData, err := stdReader.ReadString('\n')
					if err != nil {
						log.Fatal(err)
					}

					sendData = strings.Replace(sendData, "\n", "", -1)
					sendData = strings.Replace(sendData, "\r", "", -1) //suncj
					_result, err := strconv.Atoi(sendData)
					if err != nil {
						log.Fatal(err)
					}
					chain := make([]Block, 0)

					mutex.Lock()
					newBlock := generateNewBlock(BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1], _result)
					newBlock.Validator=address
					chain=append(chain,newBlock)
					bytes, err := json.Marshal(chain)
					if err != nil {
						log.Println(err)
					}
					rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
					rw.Flush()
					mutex.Unlock()
			}
			

			 
		 
	 
	

}
func WriteData(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(BlockchainInstance.Blocks)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		sendData = strings.Replace(sendData, "\n", "", -1)
		sendData = strings.Replace(sendData, "\r", "", -1) //suncj
		_result, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}
		newBlock := GenerateBlock(BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1], _result)

		if len(BlockchainInstance.TxPool.AllTx) > 0 {
			BlockchainInstance.PackageTx(&newBlock)
		}else {
			newBlock.Accounts = BlockchainInstance.LastBlock().Accounts
		}

		if IsBlockValid(newBlock, BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1]) {
			mutex.Lock()
			BlockchainInstance.Blocks = append(BlockchainInstance.Blocks, newBlock)
			mutex.Unlock()
		}

		bytes, err := json.Marshal(BlockchainInstance.Blocks)
		if err != nil {
			log.Println(err)
		}

		//spew.Dump(BlockchainInstance.Blocks)

		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}

}
 

// make sure block is valid by checking index, and comparing the hash of the previous block
func IsBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if CalculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// SHA256 hashing
func CalculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.Result) + block.PrevHash+block.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func GenerateBlock(oldBlock Block, Result int) Block {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Result = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = CalculateHash(newBlock)

	return newBlock
}
func generateNewBlock(oldBlock Block, Result int) Block {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Result = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash =""

	return newBlock
}
func generateBlock_Pow(oldBlock Block, Result int) Block {
	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Result = Result
	newBlock.PrevHash = oldBlock.Hash
	Difficulty:= 1

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		if !isHashValid(CalculateHash(newBlock), Difficulty) {
			fmt.Println(CalculateHash(newBlock), " do more work!")
			time.Sleep(time.Second)
			continue
		} else {
			fmt.Println(CalculateHash(newBlock), " work done!")
			newBlock.Hash = CalculateHash(newBlock)
			break
		}

	}
	return newBlock
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}
func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
func generateBlock(oldBlock Block, Result int, address string) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Result = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = CalculateHash(newBlock)
	newBlock.Validator = address

	return newBlock, nil
}
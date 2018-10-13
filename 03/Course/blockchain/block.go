package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)



// Block represents each 'item' in the blockchain
type Block struct {
	Index     int `json:"index"`
	Timestamp string `json:"timestamp"`
	Result       int `json:"result"`
	Hash      string `json:"hash"`
	PrevHash  string `json:"prevhash"`
	Proof        uint64           `json:"proof"`
	Transactions []Transaction `json:"transactions"`
	AccountState  string  `json:"accountstat"`
	Difficulty int `json:"difficulty"`
	Nonce      string `json:"nonce"`
}

func (b *Block) Serialize() ([]byte){
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	encoder.Encode(b)
	return result.Bytes()
}

func DeserializeBlock(data []byte) *Block{
	decoder := gob.NewDecoder(bytes.NewReader(data))
	b := &Block{}
	decoder.Decode(b)
	return b
}

// make sure block is valid by checking index, and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// SHA256 hasing
func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.Result) + block.PrevHash + block.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlockPOW(oldBlock Block, Result int) Block {
	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Result = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficulty = difficulty

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		if !isHashValid(calculateHash(newBlock), newBlock.Difficulty) {
			//fmt.Println(calculateHash(newBlock), " do more work!")
			time.Sleep(time.Second)
			continue
		} else {
			fmt.Println(calculateHash(newBlock), " work done!")
			newBlock.Hash = calculateHash(newBlock)
			break
		}
	}
	return newBlock
}

func generateBlockPOS(oldBlock Block, Result int) Block {
	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Result = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficulty = 0
	newBlock.Hash = calculateHash(newBlock)
	return newBlock
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}


func GenerateGenesisBlock () Block{
	t := time.Now()

	genesisBlock := Block{
		0,
		t.String(),
		0,
		"",
		"",
		100,
		nil,
		"",
		difficulty,
		"",
	}
	genesisBlock.Hash = calculateHash(genesisBlock)
	return genesisBlock
}
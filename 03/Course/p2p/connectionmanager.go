package p2p

import (
	"bytes"
	"crypto/sha256"
	"Course/blockchain"
	"bufio"
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"
)

type Connection struct {
	mutex sync.Mutex
	rw *bufio.ReadWriter
}

func NewConnection(rw *bufio.ReadWriter) *Connection{
	c := & Connection{
		mutex: sync.Mutex{},
		rw: rw,
	}
	return c
}

func (c *Connection) SendRaw(s string) {
	c.mutex.Lock()
	c.rw.WriteString(s)
	c.rw.WriteByte('\n')
	c.rw.Flush()
	c.mutex.Unlock()
}

type Message struct {
	Type string
	Hash string
	Data *[]byte
}

func NewMessage(Type string, object interface{}) (m *Message){
	Data, err := json.Marshal(object)
	if err != nil {
		log.Println("marshal data error:", err)
		return
	}
	m = &Message {
		Type: Type,
		Data: &Data,
	}
	ss := [][]byte {[]byte(Type), Data}
	s :=  bytes.Join(ss, []byte(""))
	h := sha256.New()
	h.Write(s)
	m.Hash = hex.EncodeToString(h.Sum(nil))
	return m
}


func (c *Connection)SendMessage(m *Message) {
	bytes, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return
	}
	messageMap.Store(m.Hash, true)
	c.SendRaw(string(bytes))
}

func (c *Connection)SendCandidateBlock(b *blockchain.Block) {
	log.Println("send block")
	m := NewMessage("candidateblock", b)
	c.SendMessage(m)
}

func (c *Connection)SendConfirmedBlock(b *blockchain.Block) {
	log.Println("send block")
	m := NewMessage("confirmedblock", b)
	c.SendMessage(m)
}

/*
func (c *Connection)SendBlockchain(b *blockchain.Blockchain) {
	//TODO: add mutex
	log.Println("send blockchain")
	m := NewMessage("blockchain", b.Blocks)
	c.SendMessage(m)
}*/

func (c *Connection)SendTransaction(t *blockchain.Transaction) {
	log.Println("send transaction")
	m := NewMessage("transaction", t)
	c.SendMessage(m)
}



type ConnectionManager struct {
	mutex sync.Mutex
	connections []*Connection
}

func NewConnectionManager()  *ConnectionManager {
	m := & ConnectionManager{
		connections: make([]*Connection, 0),
		mutex: sync.Mutex{},
	}
	return m
}

func (self *ConnectionManager) AddConnection(c *Connection)  {
	log.Println("add connection")
	self.mutex.Lock()
	self.connections = append(self.connections, c)
	self.mutex.Unlock()
}

func (self *ConnectionManager) BroadcastRaw(msg string)  {
	for _, connection := range self.connections {
		connection.SendRaw(msg)
	}
}

func (self *ConnectionManager) BroadcastCandidateBlock(b *blockchain.Block)  {
	for _, connection := range self.connections {
		connection.SendCandidateBlock(b)
	}
}
func (self *ConnectionManager) BroadcastConfirmedBlock(b *blockchain.Block)  {
	for _, connection := range self.connections {
		connection.SendConfirmedBlock(b)
	}
}
/*
func (self *ConnectionManager) BroadcastBlockchain(b *blockchain.Blockchain)  {
	for _, connection := range self.connections {
		connection.SendBlockchain(b)
	}
}*/

func (self *ConnectionManager) BroadcastTransaction(t *blockchain.Transaction)  {
	for _, connection := range self.connections {
		connection.SendTransaction(t)
	}
}

var NodeManagerInstance = NewConnectionManager()

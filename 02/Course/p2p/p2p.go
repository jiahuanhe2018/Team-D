package p2p

import (
	"Course/blockchain"
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"io"
	"log"
	mrand "math/rand"
	"sync"
	"time"
)


var messageMap = sync.Map{}

func RunNode(consensus *string, target *string, listenF *int, secio *bool, seed *int64){

	go BroadcastConfirmedBlocks()
	// Make a host that listens on the given multiaddress
	ha, err := MakeBasicHost(*listenF, *secio, *seed)
	if err != nil {
		log.Fatal(err)
	}

	if *target == "" {
		log.Println("listening for connections")
		// Set a stream handler on host A. /p2p/1.0.0 is
		// a user-defined protocol name.
		ha.SetStreamHandler("/p2p/1.0.0", HandleStream)

		select {} // hang forever
		/**** This is where the listener code ends ****/
	} else {
		ha.SetStreamHandler("/p2p/1.0.0", HandleStream)

		// The following code extracts target's peer ID from the
		// given multiaddress
		ipfsaddr, err := ma.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}

		pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}

		peerid, err := peer.IDB58Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}

		// Decapsulate the /ipfs/<peerID> part from the target
		// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		// We have a peer ID and a targetAddr so we add it to the peerstore
		// so LibP2P knows how to contact it
		ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

		log.Println("opening stream")
		// make a new stream from host B to host A
		// it should be handled on host A by the handler we set above because
		// we use the same /p2p/1.0.0 protocol
		s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		// Create a buffered stream so that read and writes are non blocking.
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		connection := NewConnection(rw)
		NodeManagerInstance.AddConnection(connection)
		// Create a thread to read and write data.
		go SyncFullBlockchain(connection)
		go RecieveMessage(connection)
		go BroadcastCandidateBlocks()

		select {} // hang forever

	}
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
	if secio {
		log.Printf("Now run \"go run main.go -c chain -consensus %s -s lzhx -l %d -d %s -secio\" on a different terminal\n", blockchain.Consensus, listenPort+2, fullAddr)
	} else {
		log.Printf("Now run \"go run main.go -c chain -consensus %s -s lzhx -l %d -d %s\" on a different terminal\n", blockchain.Consensus, listenPort+2, fullAddr)
	}

	return basicHost, nil
}

func HandleStream(s net.Stream) {
	log.Println("Got a new stream!")
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	connection := NewConnection(rw)
	NodeManagerInstance.AddConnection(connection)
	go RecieveMessage(connection)
	go SyncFullBlockchain(connection)

	// stream 's' will stay open until you close it (or the other side closes it).
}

func RecieveMessage(c *Connection) {

	for {
		str, err := c.rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			message := Message{}
			json.Unmarshal([]byte(str), &message)

			if _, ok := messageMap.Load(message.Hash);  ok {
				log.Println("Receive Duplicated Message", message.Type)
				continue
			}else {
				log.Println("Receive Message ", message.Type)
			}
			messageMap.Store(message.Hash, true)

			if message.Type == "blockchain" {
				chain := make([]blockchain.Block, 0)
				if err := json.Unmarshal(*message.Data, &chain); err != nil {
					log.Fatal(err)
				}
				blockchain.BlockchainInstance.Lock()
				//TODO: add mutex
				if len(chain) > len(blockchain.BlockchainInstance.Blocks) {
					blockchain.BlockchainInstance.Blocks = chain
					bytes, err := json.MarshalIndent(blockchain.BlockchainInstance.Blocks, "", "  ")
					if err != nil {
						log.Fatal(err)
					}
					// Green console color: 	\x1b[32m
					// Reset console color: 	\x1b[0m
					fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
				}
				blockchain.BlockchainInstance.Unlock()
			} else if message.Type == "transaction" {
				tx := &blockchain.Transaction{}
				err := json.Unmarshal(*message.Data, tx)
				if err != nil {
					log.Println(err)
					continue
				}
				blockchain.BlockchainInstance.AddTxPool(tx)
			} else if message.Type == "candidateblock" {
				b := blockchain.Block{}
				err := json.Unmarshal(*message.Data, &b)
				if err != nil {
					log.Println("unmarshal block error", err)
					continue
				}
				if blockchain.IsMaster {
					blockchain.CandidateBlocks = append(blockchain.CandidateBlocks, b)
					//no forward
					continue
				}
			} else if message.Type == "confirmedblock" {
				b := blockchain.Block{}
				err := json.Unmarshal(*message.Data, &b)
				if err != nil {
					log.Println("unmarshal block error", err)
					continue
				}
				spew.Dump(b)
				blockchain.BlockchainInstance.AppendBlock(&b)
			}

			NodeManagerInstance.BroadcastRaw(str)
		}
	}
}


func SyncFullBlockchain(c *Connection) {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			c.SendBlockchain(&blockchain.BlockchainInstance)
		}
	}()

}

func BroadcastConfirmedBlocks(){
	for b := range blockchain.ToBroadcastConfirmedtBlocks {
		//TODO: Only broadcast new block
		log.Println(b)
		//NodeManagerInstance.BroadcastBlockchain(&blockchain.BlockchainInstance)
		NodeManagerInstance.BroadcastConfirmedBlock(&b)
	}
}

func BroadcastCandidateBlocks(){
	for b := range blockchain.ToBroadcastCandidateBlocks {
		log.Println(b)
		NodeManagerInstance.BroadcastCandidateBlock(&b)
	}
}


package solo

import (
	"firstchain/common"
	"firstchain/consensus"
	"firstchain/consensus/blockpool"
	"math/big"
	"time"

	"github.com/libp2p/go-libp2p-crypto"

	dbtypes "firstchain/basic"
	"firstchain/event"
	"firstchain/p2p"
	"firstchain/state"
	"firstchain/txpool"
)

var (
	log = common.GetLogger("consensus")
)

const (
	BP               = iota  // block producer
	NBP                      // not block producer
	REMOVE_THRESHOLD = 65535 // height threshold of clearing block pool
)

type Blockchain interface {
	LastBlock() *dbtypes.Block
}

type SoloEngine struct {
	config    *Config
	chain     Blockchain
	validator consensus.BlockValidator
	blockPool *blockpool.BlockPool
	txPool    *txpool.TxPool
	state     *state.StateDB
	event     *event.TypeMux

	processLock chan struct{} // channel lock to prevent concurrent block processing
	quitCh      chan struct{}

	address common.Address
	privKey crypto.PrivKey

	newBlockSub  event.Subscription // listen for the new block from block pool
	newTxsSub    event.Subscription // listen for the new pending transactions from txpool
	consensusSub event.Subscription // listen for the new proposed block executed by executor
	receiptsSub  event.Subscription // listen for the receipts after executing
	commitSub    event.Subscription // listen for the commit block completed from executor
	errSub       event.Subscription // listen for the executor's error event
}

func New(config *common.Config, state *state.StateDB, chain Blockchain, blValidator consensus.BlockValidator, txValidator consensus.TxValidator) (*SoloEngine, error) {
	conf := newConfig(config)
	soloEngine := &SoloEngine{
		config:    conf,
		event:     event.GetEventhub(),
		chain:     chain,
		blockPool: blockpool.NewBlockPool(config, blValidator, nil, log, common.ProposeBlockMsg),
	}

	if conf.BP {
		privKey, err := crypto.UnmarshalPrivateKey(conf.PrivKey)
		if err != nil {
			return nil, err
		}
		soloEngine.privKey = privKey
		soloEngine.address, err = common.GenAddrByPrivkey(privKey)
		if err != nil {
			return nil, err
		}
		soloEngine.txPool = txpool.NewTxPool(config, txValidator, state, true, false)
	} else {
		soloEngine.txPool = txpool.NewTxPool(config, txValidator, state, true, true)
	}
	return soloEngine, nil
}

func (solo *SoloEngine) Start() error {
	solo.newBlockSub = solo.event.Subscribe(&dbtypes.BlockReadyEvent{})
	solo.commitSub = solo.event.Subscribe(&dbtypes.CommitCompleteEvent{})
	solo.newTxsSub = solo.event.Subscribe(&dbtypes.ExecPendingTxEvent{})
	solo.consensusSub = solo.event.Subscribe(&dbtypes.ConsensusEvent{})
	solo.receiptsSub = solo.event.Subscribe(&dbtypes.NewReceiptsEvent{})

	solo.processLock <- struct{}{}

	go solo.listen()
	go solo.txPool.Start()
	return nil
}

func (solo *SoloEngine) listen() {
	for {
		select {
		case ev := <-solo.newTxsSub.Chan():
			txs := ev.(*dbtypes.ExecPendingTxEvent).Txs
			go solo.proposeBlock(txs)
		case ev := <-solo.newBlockSub.Chan():
			block := ev.(*dbtypes.BlockReadyEvent).Block
			go solo.broadcast(block)
			go solo.process()
		case ev := <-solo.commitSub.Chan():
			block := ev.(*dbtypes.CommitCompleteEvent).Block
			go solo.commitComplete(block)
		case ev := <-solo.consensusSub.Chan():
			// this channel will be passed when executor completes to propose block,
			// and always done by bp
			event := ev.(*dbtypes.ConsensusEvent)
			go solo.validateAndCommit(event.Block, event.Receipts)
		case ev := <-solo.receiptsSub.Chan():
			// this channel will be passed when executor completes to process block and ge receipts,
			// and always done by not-bp
			rev := ev.(*dbtypes.NewReceiptsEvent)
			go solo.validateAndCommit(rev.Block, rev.Receipts)
		case ev := <-solo.errSub.Chan():
			err := ev.(*dbtypes.ErrOccurEvent).Err
			log.Errorf("error occurs at executor processing: %s", err)
			solo.processLock <- struct{}{}
			if !solo.config.BP {
				go solo.process()
			}
		case <-solo.quitCh:
			solo.newTxsSub.Unsubscribe()
			solo.newBlockSub.Unsubscribe()
			solo.commitSub.Unsubscribe()
			solo.consensusSub.Unsubscribe()
			solo.receiptsSub.Unsubscribe()
			return
		}
	}
}

func (solo *SoloEngine) Stop() error {
	solo.quitCh <- struct{}{}
	solo.txPool.Stop()
	return nil
}

func (solo *SoloEngine) Address() common.Address {
	return solo.address
}

// process processes the block received from the solo BP.
// It is invoked by not-BP peers.
func (solo *SoloEngine) process() {
	<-solo.processLock
	block := solo.blockPool.GetBlock(solo.chain.LastBlock().Height() + 1)
	if block != nil {
		go solo.event.Post(&dbtypes.ExecBlockEvent{block})
	} else {
		solo.processLock <- struct{}{}
	}
}

// proposeBlock proposes a new block with given transactions retrieved from tx_pool.
func (solo *SoloEngine) proposeBlock(txs dbtypes.Transactions) {
	<-solo.processLock
	header := &dbtypes.Header{
		ParentHash: solo.chain.LastBlock().Hash(),
		Height:     solo.chain.LastBlock().Height() + 1,
		Coinbase:   solo.Address(),
		Extra:      solo.config.Extra,
		Time:       new(big.Int).SetInt64(time.Now().Unix()),
		GasLimit:   solo.config.GasLimit,
	}

	block := dbtypes.NewBlock(header, txs)
	go solo.event.Post(&dbtypes.ProposeBlockEvent{block})
	log.Infof("Block producer %s propose a new block, height = #%d", solo.Address(), block.Height())
}

func (solo *SoloEngine) broadcast(block *dbtypes.Block) error {
	data, err := block.Serialize()
	if err != nil {
		return err
	}
	go solo.event.Post(&p2p.BroadcastEvent{
		Typ:  common.ProposeBlockMsg,
		Data: data,
	})
	return nil
}

func (solo *SoloEngine) Finalize(header *dbtypes.Header, state *state.StateDB, txs dbtypes.Transactions, receipts dbtypes.Receipts) (*dbtypes.Block, error) {
	root, err := state.IntermediateRoot()
	if err != nil {
		return nil, err
	}
	header.StateRoot = root
	header.ReceiptsHash = receipts.Hash()

	header.TxRoot = txs.Hash()
	newBlk := dbtypes.NewBlock(header, txs)
	newBlk.PubKey, err = solo.privKey.GetPublic().Bytes()
	if err != nil {
		return nil, err
	}
	newBlk.Sign(solo.privKey)
	return newBlk, nil
}

func (solo *SoloEngine) validateAndCommit(block *dbtypes.Block, receipts dbtypes.Receipts) error {
	if err := solo.validator.ValidateState(block, solo.state, receipts); err != nil {
		log.Errorf("invalid block state, err:%s", err)
		go solo.event.Post(&dbtypes.RollbackEvent{})
		solo.processLock <- struct{}{}
		return err
	}
	solo.commit(block)
	return nil
}

func (solo *SoloEngine) commit(block *dbtypes.Block) {
	go solo.event.Post(&dbtypes.CommitBlockEvent{block})
}

func (solo *SoloEngine) commitComplete(block *dbtypes.Block) {
	solo.blockPool.UpdateChainHeight(block.Height())
	solo.txPool.Drop(block.Transactions)

	solo.processLock <- struct{}{}
	go solo.broadcast(block)
	// if this peer is a NBP
	if !solo.config.BP {
		go solo.process() // call next process
	}
}

func (solo *SoloEngine) Protocols() []p2p.Protocol {
	return nil
}
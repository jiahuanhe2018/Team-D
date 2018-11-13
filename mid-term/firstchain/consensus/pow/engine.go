package pow

import (
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p-crypto"
	"math"
	"math/big"
	"runtime"
	"sync/atomic"
	"time"
	"firstchain/common"
	"firstchain/consensus"
	"firstchain/consensus/blockpool"
	 
	"firstchain/state"
	"firstchain/txpool"
	dbtypes "firstchain/basic"
	"firstchain/event"
	"firstchain/p2p"
)

var (
	log = common.GetLogger("consensus")
)

const (
	//maxNonce                = math.MaxUint64 << 5  
	maxNonce              = math.MaxUint32 //<< 5  suncj
	blockGapForDiffAdjust = 20160
	maxDiffculty   		  = math.MaxUint64 //<< 5  suncj
	minerReward           = 998
)
 
type consensusInfo struct {
	Difficulty uint64 `json:"Difficulty"` // difficulty target bits for mining
	Nonce      uint64 `json:"nonce"`      // computed result
}

func (ci *consensusInfo) Serialize() ([]byte, error) {
	return json.Marshal(ci)
}

// ProofOfWork implements proof-of-work consensus algorithm.
type ProofOfWork struct {
	config           *Config
	event            *event.TypeMux
	chain            consensus.Blockchain
	state            *state.StateDB
	blockPool        *blockpool.BlockPool
	txPool           *txpool.TxPool
	blValidator      consensus.BlockValidator // block validator
	csValidator      *csValidator             // consensus validator
	blockNum         uint64                   // new block num at certain difficulty period
	currMiningHeader *dbtypes.Header            // block header that being mined currently

	address common.Address
	privKey crypto.PrivKey

	quitCh      chan struct{}
	processLock chan struct{}
	mineStopCh  chan struct{} // listen for outer event to make miner stop

	// cache
	currDiff uint64 // current difficulty

	consensusSub event.Subscription // listen for the new proposed block executed by executor
	newTxsSub    event.Subscription // listen for the new pending transactions from txpool
	newBlockSub  event.Subscription // listen for the new block from block pool
	commitSub    event.Subscription // listen for the commit block completed from executor
	receiptsSub  event.Subscription // listen for the receipts executed by executor
	errSub       event.Subscription // listen for the executor's error event
}

func New(config *common.Config, state *state.StateDB, chain consensus.Blockchain, blValidator consensus.BlockValidator, txValidator consensus.TxValidator) (*ProofOfWork, error) {
	conf := newConfig(config)

	csValidator := newCsValidator(chain)

	pow := &ProofOfWork{
		config:      conf,
		chain:       chain,
		state:       state,
		blValidator: blValidator,
		csValidator: csValidator,
		event:       event.GetEventhub(),
		quitCh:      make(chan struct{}),
		blockPool:   blockpool.NewBlockPool(config, blValidator, csValidator, log, common.ProposeBlockMsg),
	}

	// if is miner node
	if conf.Miner {
		privKey, err := crypto.UnmarshalPrivateKey(conf.PrivKey)
		if err != nil {
			return nil, err
		}
		pow.privKey = privKey
		pow.address, err = common.GenAddrByPrivkey(privKey)
		if err != nil {
			return nil, err
		}
		pow.txPool = txpool.NewTxPool(config, txValidator, state, true, false)
	} else {
		pow.txPool = txpool.NewTxPool(config, txValidator, state, true, true)
	}

	return pow, nil
}

func (pow *ProofOfWork) Start() error {
	pow.consensusSub = pow.event.Subscribe(&dbtypes.ConsensusEvent{})
	pow.commitSub = pow.event.Subscribe(&dbtypes.CommitCompleteEvent{})
	pow.receiptsSub = pow.event.Subscribe(&dbtypes.NewReceiptsEvent{})
	pow.newBlockSub = pow.event.Subscribe(&dbtypes.BlockReadyEvent{})
	pow.errSub = pow.event.Subscribe(&dbtypes.ErrOccurEvent{})

	go pow.listen()
	return nil
}

func (pow *ProofOfWork) listen() {
	for {
		select {
		case ev := <-pow.consensusSub.Chan():
			event := ev.(*dbtypes.ConsensusEvent)
			go pow.seal(event.Block, event.Receipts)
		case <-pow.newBlockSub.Chan():
			go pow.process()
		case ev := <-pow.receiptsSub.Chan():
			rev := ev.(*dbtypes.NewReceiptsEvent)
			go pow.validateAndCommit(rev.Block, rev.Receipts)
		case ev := <-pow.commitSub.Chan():
			block := ev.(*dbtypes.CommitCompleteEvent).Block
			go pow.commitComplete(block)
		case ev := <-pow.errSub.Chan():
			err := ev.(*dbtypes.ErrOccurEvent).Err
			log.Errorf("error occurs at executor process: %s", err)
			pow.processLock <- struct{}{}
			go pow.process()
		case <-pow.quitCh:
			pow.consensusSub.Unsubscribe()
			return
		}
	}
}

func (pow *ProofOfWork) Stop() error {
	close(pow.quitCh)
	return nil
}

func (pow *ProofOfWork) Addr() common.Address {
	return pow.address
}

func (pow *ProofOfWork) difficulty() (uint64, error) {
	if diff := atomic.LoadUint64(&pow.currDiff); diff != 0 {
		return diff, nil
	}
	last := pow.chain.LastBlock()
	consensus := consensusInfo{}
	err := json.Unmarshal(last.ConsensusInfo(), &consensus)
	if err != nil {
		return 0, err
	}
	atomic.StoreUint64(&pow.currDiff, consensus.Difficulty)
	return consensus.Difficulty, nil
}

// proposeBlock proposes a new pure block
func (pow *ProofOfWork) proposeBlock() {
	last := pow.chain.LastBlock()
	header := &dbtypes.Header{
		ParentHash: last.ParentHash(),
		Height:     last.Height(),
		Coinbase:   pow.address,
		Time:       new(big.Int).SetInt64(time.Now().Unix()),
		GasLimit:   pow.config.GasLimit,
	}

	block := dbtypes.NewBlock(header, pow.txPool.Pending())
	log.Infof("Block producer %s propose a new block, height = #%d", pow.Addr(), block.Height())
	go pow.event.Post(&dbtypes.ProposeBlockEvent{block})
}

func (pow *ProofOfWork) process() {
	block := pow.blockPool.GetBlock(pow.chain.LastBlock().Height() + 1)
	if block == nil {
		return
	}
	if pow.currMiningHeader != nil && pow.currMiningHeader.Height == block.Height() {
		if err := pow.csValidator.Validate(block); err == nil {
			pow.mineStopCh <- struct{}{}
		} else {
			return
		}
	}
	<-pow.processLock
	if block != nil {
		go pow.event.Post(&dbtypes.ExecBlockEvent{block})
	} else {
		pow.processLock <- struct{}{}
	}
}

// run performs proof-of-work.
// It will return error if other miner found a block with the same height
func (pow *ProofOfWork) run(block *dbtypes.Block) error {
	var (
		abort = make(chan struct{})
		found = make(chan uint64)
		n     = runtime.NumCPU()
		avg   = maxNonce/n
	)
	newDiff, err := pow.adjustDiff()
	if err != nil {
		return fmt.Errorf("cannot adjust difficulty when start to mine block #%d", block.Height())
	}
	pow.currMiningHeader = block.Header
	for i := 0; i < n; i++ {
		go newWorker(newDiff, uint64(avg*i), uint64(avg*(i+1)), block.Header).run(found, abort)
	}
	defer close(found)

	var nonce uint64
	select {
	case nonce = <-found:
		close(abort)
	case <-pow.mineStopCh:
		// close all mining work
		close(abort)
		pow.processLock <- struct{}{}
		return fmt.Errorf("stop mining, a block with the same height #%d found", block.Height())
	}

	csinfo := &consensusInfo{
		Difficulty: newDiff,
		Nonce:      nonce,
	}
	data, err := csinfo.Serialize()
	if err != nil {
		return fmt.Errorf("failed to encode consensus info, err:%s", err)
	}
	block.Header.ConsensusInfo = data
	pow.currMiningHeader = nil
	return nil
}

func (pow *ProofOfWork) Finalize(header *dbtypes.Header, state *state.StateDB, txs dbtypes.Transactions, receipts dbtypes.Receipts) (*dbtypes.Block, error) {
	// reward miner
	pow.state.AddBalance(header.Coinbase, new(big.Int).SetUint64(minerReward))

	// calculate state root
	root, err := state.IntermediateRoot()
	if err != nil {
		return nil, err
	}
	header.StateRoot = root
	header.ReceiptsHash = receipts.Hash()

	header.TxRoot = txs.Hash()
	newBlk := dbtypes.NewBlock(header, txs)
	newBlk.PubKey, err = pow.privKey.GetPublic().Bytes()
	if err != nil {
		return nil, err
	}
	newBlk.Sign(pow.privKey)
	return newBlk, nil
}

func (pow *ProofOfWork) seal(block *dbtypes.Block, receipts dbtypes.Receipts) error {
	if err := pow.run(block); err != nil {
		go pow.event.Post(&dbtypes.RollbackEvent{})
		return err
	}

	return pow.validateAndCommit(block, receipts)
}

// validateAndCommit validate block's state and consensus info, and finally commit it.
func (pow *ProofOfWork) validateAndCommit(block *dbtypes.Block, receipts dbtypes.Receipts) error {
	if err := pow.blValidator.ValidateState(block, pow.state, receipts); err != nil {
		log.Errorf("invalid block state, err:%s", err)
		go pow.event.Post(&dbtypes.RollbackEvent{})
		return err
	}

	if err := pow.csValidator.Validate(block); err != nil {
		log.Errorf("invalid block consensus info, err:%s", err)
		return err
	}
	pow.commit(block)
	return nil
}

func (pow *ProofOfWork) commit(block *dbtypes.Block) {
	go pow.event.Post(&dbtypes.CommitBlockEvent{block})
}

func (pow *ProofOfWork) commitComplete(block *dbtypes.Block) {
	pow.blockPool.UpdateChainHeight(block.Height())
	pow.txPool.Drop(block.Transactions)

	pow.processLock <- struct{}{}
	go pow.broadcast(block)
	if !pow.config.Miner {
		go pow.process()
	} else {
		go pow.proposeBlock()
	}
}

func (pow *ProofOfWork) broadcast(block *dbtypes.Block) error {
	data, err := block.Serialize()
	if err != nil {
		return err
	}

	go pow.event.Post(&p2p.BroadcastEvent{
		Typ:  common.ProposeBlockMsg,
		Data: data,
	})
	return nil
}

func (pow *ProofOfWork) Protocols() []p2p.Protocol {
	return nil
}

// adjustDiff adjust difficulty of mining blocks
func (pow *ProofOfWork) adjustDiff() (uint64, error) {
	curr := pow.chain.LastBlock()
	if atomic.LoadUint64(&pow.blockNum) < blockGapForDiffAdjust {
		return pow.difficulty()
	}
	old := pow.chain.GetBlockByHeight(curr.Height() - blockGapForDiffAdjust)
	currDiff, err := pow.difficulty()
	if err != nil {
		return 0, err
	}
	newDiff := computeNewDiff(currDiff, curr, old)
	if newDiff > maxDiffculty {
		newDiff = maxDiffculty
	}
	atomic.StoreUint64(&pow.currDiff, newDiff)
	atomic.StoreUint64(&pow.blockNum, 0)
	return newDiff, nil
}

func (pow *ProofOfWork) rollback() {
	pow.event.Post(&dbtypes.RollbackEvent{})
}

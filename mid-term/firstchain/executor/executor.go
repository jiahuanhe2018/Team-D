package executor

import (
	"firstchain/chain"
	"firstchain/common"
	"firstchain/consensus"
	"firstchain/state"
	"fmt"
	"sync"
	"sync/atomic"

	batcher "github.com/yyh1102/go-batcher"

	dbtypes "firstchain/basic"
	"firstchain/event"

	"github.com/pkg/errors"
)

var (
	log = common.GetLogger("executor")

	errReceiptNumInvalid = errors.New("the number of transactions not equal to receipt")
)

// Processor represents the interface of block processor
type Processor interface {
	Process(block *dbtypes.Block) (dbtypes.Receipts, error)
}

type Executor struct {
	conf      *common.Config
	db        *dbtypes.SmallDB
	chain     *chain.Blockchain // Blockchain wrapper
	batch     batcher.Batch     // Batch for creating new block
	state     *state.StateDB
	engine    consensus.Engine
	event     *event.TypeMux
	validator *BlockValidator
	quitCh    chan struct{}

	receiptsCache sync.Map     // Receipts cache, map[uint64]types.Receipts
	processing    atomic.Value // Processing state, 1 means processing, 0 means idle

	// execution context
	gasLimit  uint64 // Block gas limit
	versionId int    // Snapshot id

	execBlockSub    event.Subscription // Subscribe new block ready event from block_pool
	proposeBlockSub event.Subscription // Subscribe propose new block event
	commitSub       event.Subscription // Subscribe state commit event
	rollbackSub     event.Subscription // Subscribe rollback event
}

func New(config *common.Config, database dbtypes.Database, chain *chain.Blockchain, engine consensus.Engine) *Executor {
	executor := &Executor{
		conf:      config,
		db:        dbtypes.NewSmallDB(database),
		chain:     chain,
		engine:    engine,
		event:     event.GetEventhub(),
		validator: NewBlockValidator(config, chain),
		quitCh:    make(chan struct{}),
		
	}
	return executor
}

func (ex *Executor) Init(statedb *state.StateDB) error {
	genesis := ex.chain.Genesis()
	if genesis == nil {
		newGenesis, err := ex.createGenesis(statedb)
		if err != nil {
			log.Errorf("failed to create genesis when init executor, %s", err)
			return err
		}
		genesis = newGenesis
	}
	statedbgenesis, err := state.New(ex.db.LDB(), genesis.StateRoot().Bytes())
	if err != nil {
		log.Errorf("failed to init state when init executor, %s", err)
		return err
	}
	ex.state = statedbgenesis
	//ex.commit(newGenesis)
	return nil
}

func (ex *Executor) Start() error {
	ex.execBlockSub = ex.event.Subscribe(&dbtypes.ExecBlockEvent{})
	ex.proposeBlockSub = ex.event.Subscribe(&dbtypes.ProposeBlockEvent{})
	ex.commitSub = ex.event.Subscribe(&dbtypes.CommitBlockEvent{})
	ex.rollbackSub = ex.event.Subscribe(&dbtypes.RollbackEvent{})
	go ex.listen()
	return nil
}

func (ex *Executor) listen() {
	for {
		select {
		case ev := <-ex.proposeBlockSub.Chan():
			block := ev.(*dbtypes.ProposeBlockEvent).Block
			if err := ex.proposeBlock(block); err != nil {
				log.Errorf("failed to propose block #%d, err:%s", block.Height(), err)
				go ex.event.Post(&dbtypes.ErrOccurEvent{err})
				ex.rollback()
			}
		case ev := <-ex.execBlockSub.Chan():
			block := ev.(*dbtypes.ExecBlockEvent).Block
			if err := ex.applyBlock(block); err != nil {
				log.Errorf("failed to apply block %s, err:%s", block.Hash(), err)
				go ex.event.Post(&dbtypes.ErrOccurEvent{err})
				ex.rollback()
			}
		case ev := <-ex.commitSub.Chan():
			block := ev.(*dbtypes.CommitBlockEvent).Block
			if err := ex.commit(block); err != nil {
				log.Errorf("failed to commit block %s, and roll back. err:%s", block.Hash(), err)
				go ex.event.Post(&dbtypes.ErrOccurEvent{err})
				ex.rollback()
			}
		case <-ex.rollbackSub.Chan():
			ex.rollback()
		case <-ex.quitCh:
			ex.proposeBlockSub.Unsubscribe()
			ex.commitSub.Unsubscribe()
			ex.execBlockSub.Unsubscribe()
			return
		}
	}
}

func (ex *Executor) Stop() error {
	close(ex.quitCh)
	return nil
}

func (ex *Executor) lastHeight() uint64 {
	return ex.chain.LastBlock().Height()
}

// processState get the current processing state, and returns 1 processing, or 0 idle
func (ex *Executor) processState() int {
	if p := ex.processing.Load(); p != nil {
		return p.(int)
	}
	return 0
}

// Validate validate block body.
func (ex *Executor) validate(block *dbtypes.Block) error {
	return ex.validator.ValidateBody(block)
}

// applyBlockBlock process the validation and execute the received block.
func (ex *Executor) applyBlock(block *dbtypes.Block) error {
	ex.versionId = ex.state.Snapshot()

	if err := ex.validate(block); err != nil {
		log.Errorf("block is invalid, %s", err)
		return err
	}
	if currHeight := ex.chain.LastBlock().Height(); block.Height() != currHeight+1 {
		return fmt.Errorf("block height is not match, demand #%d, got #%d", currHeight+1, block.Height())
	}
	ex.state.UpdateCurrHeight(block.Height())
	receipts, err := ex.Process(block)
	if err != nil {
		log.Errorf("failed to execute block #%d, err:%s", block.Height(), err)
		ex.rollback()
		return err
	}

	// Save receipts to cache
	ex.receiptsCache.Store(block.Height(), receipts)

	// Add block in memory blockchain
	if err := ex.chain.AddBlock(block); err != nil {
		log.Errorf("failed to add block to blockchain cache, err:%s", err)
		return err
	}

	// Send receipts to engine
	go ex.event.Post(&dbtypes.NewReceiptsEvent{
		Block:    block,
		Receipts: receipts,
	})

	return nil
}

// proposeBlock executes new transactions from tx_pool and pack a new block.
// The new block is created by consensus engine and does not include state_root, tx_root and receipts_root.
func (ex *Executor) proposeBlock(block *dbtypes.Block) error {
	ex.versionId = ex.state.Snapshot()

	if err := ex.validate(block); err != nil {
		log.Errorf("block is invalid, %s", err)
		return err
	}

	ex.state.UpdateCurrHeight(block.Height())
	receipts, err := ex.Process(block)
	if err != nil {
		log.Errorf("failed to exec block #%d, err:%s", block.Height(), err)
		return err
	}

	// Save receipts to cache
	ex.receiptsCache.Store(block.Height(), receipts)

	// Add block in memory blockchain
	if err := ex.chain.AddBlock(block); err != nil {
		log.Errorf("failed to add block to blockchain cache, err:%s", err)
		return err
	}

	newBlk, err := ex.engine.Finalize(block.Header, ex.state, block.Transactions, receipts)
	if err != nil {
		log.Errorf("failed to finalize the block #%d, err:%s", block.Height(), err)
		return err
	}

	go ex.event.Post(&dbtypes.ConsensusEvent{newBlk, receipts})
	return nil
}

func (ex *Executor) resetVersion() {
	ex.versionId = -1
}

func (ex *Executor) rollback() {
	if ex.versionId == -1 {
		return
	}
	ex.state.RevertToSnapshot(ex.versionId)
	ex.versionId = -1
}

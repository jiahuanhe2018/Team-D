package boot

import (
	dbtypes "firstchain/basic"
	"firstchain/chain"
	"firstchain/common"
	"firstchain/consensus"
	"firstchain/consensus/pow"
	"firstchain/consensus/solo"
	"firstchain/consensus/vrf_bft"
	"firstchain/executor"

	"firstchain/state"
	"fmt"
)

var (
	log = common.GetLogger("firstchain")
)

// Small implements the firstchain full node service
type Small struct {
	config *common.Config
	db     dbtypes.Database

	engine   consensus.Engine
	executor *executor.Executor
	state    *state.StateDB
	network  *Network
	chain    *chain.Blockchain
	SmallDB  *dbtypes.SmallDB
}

func New(config *common.Config) (*Small, error) {
	ldb, err := dbtypes.NewLDBDataBase("./database")
	if err != nil {
		log.Errorf("Cannot create db, err:%s", err)
		return nil, err
	}
	// Create state db
	statedb, err := state.New(ldb, nil)
	if err != nil {
		log.Errorf("cannot init state, err:%s", err)
		return nil, err
	}

	network := NewNetwork(config)
	bc, err := chain.NewBlockchain(ldb)
	if err != nil {
		log.Error("Failed to create blockchain")
		return nil, err
	}

	small := &Small{
		config:  config,
		db:      ldb,
		network: network,
		chain:   bc,
		state:   statedb,
		SmallDB: dbtypes.NewSmallDB(ldb),
	}
	engineName := config.GetString(common.EngineName)
	blockValidator := executor.NewBlockValidator(config, bc)
	txValidator := executor.NewTxValidator(config, statedb)
	switch engineName {
	case common.SoloEngine:
		small.engine, err = solo.New(config, statedb, bc, blockValidator, txValidator)
	case common.PowEngine:
		small.engine, err = pow.New(config, statedb, bc, blockValidator, txValidator)
	case common.VrfBftEngine:
		small.engine, err = vrf_bft.New(config)
	default:
		return nil, fmt.Errorf("unknown consensus engine %s", engineName)
	}
	if err != nil {
		return nil, err
	}

	small.executor = executor.New(config, ldb, bc, small.engine)
	small.executor.Init(statedb)
	return small, nil
}

func (small *Small) Start() error {
	// Collect protocols and register in the protocol manager

	// start network
	small.network.Start()
	small.executor.Start()
	small.engine.Start()

	return nil
}

func (small *Small) init() error {
	return nil
}

func (small *Small) Config() *common.Config {
	return small.config
}

func (small *Small) RawDB() dbtypes.Database {
	return small.db
}

func (small *Small) DB() *dbtypes.SmallDB {
	return small.SmallDB
}

func (small *Small) Chain() *chain.Blockchain {
	return small.chain
}

func (small *Small) Network() *Network {
	return small.network
}

func (small *Small) StateDB() *state.StateDB {
	return small.state
}

func (small *Small) Executor() *executor.Executor {
	return small.executor
}

func (small *Small) Close() {
	small.engine.Stop()
	small.executor.Stop()
	small.network.Stop()
	small.db.Close()

	log.Info("exit system")
}

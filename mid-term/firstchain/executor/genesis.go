package executor

import (
	dbtypes "firstchain/basic"
	"firstchain/common"
	"firstchain/state"
	"math/big"
	"time"
)

func (ex *Executor) createGenesis(statedb *state.StateDB) (*dbtypes.Block, error) {
	// statedb, err := state.New(ex.db.LDB(), nil)
	// if err != nil {
	// 	log.Errorf("failed to init state, %s", err)
	// 	return nil, err
	// }
	stateRoot, err := statedb.Commit(dbtypes.GetBatch(ex.db.LDB(), 0))
	if err != nil {
		log.Errorf("compute state root failed, %s", err)
		return nil, err
	}
	genesis := dbtypes.NewBlock(
		&dbtypes.Header{
			ParentHash: common.Sha256(common.Hash{}.Bytes()),
			Height:     0,
			StateRoot:  stateRoot,
			Coinbase:   common.HexToAddress("0x0000"),
			Time:       new(big.Int).SetInt64(int64(time.Now().UnixNano())),
			Extra:      []byte("everyone can hold its asset in the blockchain"),
		},
		nil)

	if err := ex.chain.SetGenesis(genesis); err != nil {
		log.Errorf("failed to set genesis to blockchain, %s", err)
		return nil, err
	}
	return genesis, nil
}

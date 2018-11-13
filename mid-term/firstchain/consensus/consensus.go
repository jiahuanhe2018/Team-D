package consensus

import (
	"firstchain/common"
	"firstchain/state"
	dbtypes "firstchain/basic"
	"firstchain/p2p"
)

type Engine interface {
	Start() error
	Stop() error
	// protocol handled at p2p layer
	Protocols() []p2p.Protocol
	// Finalize a valid block
	Finalize(header *dbtypes.Header, state *state.StateDB, txs dbtypes.Transactions, receipts dbtypes.Receipts) (*dbtypes.Block, error)
}

type BlockValidator interface {
	ValidateHeader(header *dbtypes.Header) error
	ValidateHeaders(headers []*dbtypes.Header) (chan struct{}, chan error)
	ValidateBody(b *dbtypes.Block) error
	ValidateState(b *dbtypes.Block, state *state.StateDB, receipts dbtypes.Receipts) error
}

type TxValidator interface {
	ValidateTxs(dbtypes.Transactions) (dbtypes.Transactions, dbtypes.Transactions)
	ValidateTx(*dbtypes.Transaction) error
}

type Blockchain interface {
	LastBlock() *dbtypes.Block      // Last block in memory
	LastFinalBlock() *dbtypes.Block // Last irreversible block
	GetBlockByHeight(height uint64) *dbtypes.Block
	GetBlockByHash(hash common.Hash) *dbtypes.Block
}

//func New(config *common.Config, engineName string, state *state.StateDB, chain *chain.Blockchain, id peer.ID, blValidator *executor.BlockValidator, txValidator *executor.TxValidator) (Engine, error) {
//	switch engineName {
//	case common.SoloEngine:
//		return solo.New(config, state, chain, blValidator, txValidator)
//	case common.PowEngine:
//		return pow.New(config, state, chain, blValidator, txValidator)
//	case common.VrfBftEngine:
//		return vrf.New(config, state, chain, id, blValidator)
//	case common.DposBftEngine:
//		return dpos_bft.New(config, state, chain, id, blValidator)
//	default:
//		return nil, errors.New("invalid consensus engine name")
//	}
//}

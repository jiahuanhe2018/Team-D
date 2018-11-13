package vrf_bft

import 
(
	"firstchain/common"
	 

	"firstchain/state"
 
	dbtypes "firstchain/basic"
 
	"firstchain/p2p"
)

type VrfBft struct {
}

func New(config *common.Config) (*VrfBft, error) {

	return nil,nil
}

func (vb *VrfBft) Start() error {

	return nil
}
func (vb *VrfBft) Stop() error {

	return nil
}
// protocol handled at p2p layer
func  (vb *VrfBft) Protocols() []p2p.Protocol{

	return nil
}
// Finalize a valid block
func  (vb *VrfBft) Finalize(header *dbtypes.Header, state *state.StateDB, txs dbtypes.Transactions, receipts dbtypes.Receipts) (*dbtypes.Block, error){

	return nil,nil
}


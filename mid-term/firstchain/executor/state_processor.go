package executor

import (
	dbtypes "firstchain/basic"
	"firstchain/chain"
	"firstchain/common"
	"firstchain/state"
	"firstchain/vm"
	"firstchain/vm/evm"
)

type vmType int

const (
	EVM vmType = iota
	EWASM
	JSVM
)

// Process apply transaction in state
func (ex *Executor) Process(block *dbtypes.Block) (dbtypes.Receipts, error) {
	var (
		receipts     dbtypes.Receipts
		totalGasUsed uint64
		header       = block.Header
	)

	for i, tx := range block.Transactions {
		ex.state.Prepare(tx.Hash(), block.Hash(), uint32(i))
		receipt, gasUsed, err := ex.applyTransaction(ex.conf, ex.chain, nil, ex.state, header, tx)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
		totalGasUsed += gasUsed
	}
	block.Header.GasUsed = totalGasUsed

	return receipts, nil
}

func (ex *Executor) applyTransaction(cfg *common.Config, bc *chain.Blockchain, author *common.Address, statedb *state.StateDB, header *dbtypes.Header, tx *dbtypes.Transaction) (*dbtypes.Receipt, uint64, error) {
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms
	vmenv := NewVM(cfg, tx, header, bc, author, statedb)
	// Apply the tx to current state
	_, gasUsed, failed, err := ApplyTx(vmenv, tx)
	if err != nil {
		return nil, 0, err
	}
	// Get intermediate root of current state
	root, err := statedb.IntermediateRoot()
	if err != nil {
		return nil, 0, err
	}
	receipt := &dbtypes.Receipt{
		PostState: root,
		Status:    failed,
		TxHash:    tx.Hash(),
		GasUsed:   gasUsed,
	}
	if tx.To.Nil() {
		// Create contract call
		receipt.SetContractAddress(common.CreateAddress(tx.From, tx.Nonce))
	}

	return receipt, gasUsed, nil
}

func NewVM(config *common.Config, tx *dbtypes.Transaction, header *dbtypes.Header, bc *chain.Blockchain, author *common.Address, statedb *state.StateDB) vm.VM {
	vmType := config.GetString("vm.type")
	switch vmType {
	case vm.EVM:
		// create a vm config
		cfg := evm.Config{}
		// Create a new context to be used in the EVM environment
		context := evm.NewEVMContext(tx, header, bc, author)
		return evm.NewEVM(context, statedb, cfg)
	default:
		return nil
	}
}

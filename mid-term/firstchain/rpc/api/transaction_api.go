package api

import (
	"firstchain/common"
	dbtypes "firstchain/basic"
	 
	"firstchain/event"
	"firstchain/rpc/utils"
	"firstchain/boot"
	"firstchain/executor"
	"firstchain/state"
)

type TransactionAPI struct {
	SC *boot.Small
}

func (api *TransactionAPI) GetTransaction(hash common.Hash) (*utils.Transaction, common.Hash, uint64) {
	txMeta, err := api.SC.DB().GetTxMeta(hash)
	if err != nil {
		return nil, common.Hash{}, 0
	}

	block := api.SC.Chain().GetBlock(txMeta.Hash, txMeta.Height)
	if block != nil {
		return nil, common.Hash{}, 0
	}
	tx := block.Transactions[txMeta.TxIndex]
	return convertTransaction(tx), block.Hash(), block.Height()
}

func (api *TransactionAPI) SendTransaction(tx *dbtypes.Transaction) {
	ev := event.GetEventhub()
	go ev.Post(&dbtypes.NewTxEvent{tx})
}

// Call executes the given transaction on the state for the given block height.
// It doesn't make and changes in the state/blockchain and is useful to execute and retrieve values.
func (api *TransactionAPI) Call(tx *dbtypes.Transaction) ([]byte, error) {
	chain := api.SC.Chain()
	conf := api.SC.Config()
	statedb, err := state.New(api.SC.RawDB(), chain.LastBlock().StateRoot().Bytes())
	if err != nil {
		return nil, err
	}
	vmenv := executor.NewVM(conf, tx, chain.LastBlock().Header, chain, &tx.From, statedb)
	ret, _, _, err := executor.ApplyTx(vmenv, tx)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (api *TransactionAPI) GetReceipt(txHash common.Hash) *utils.Receipt {
	receipt, err := api.SC.DB().GetReceipt(txHash)
	if err != nil {
		return nil
	}
	return convertReceipt(receipt)
}

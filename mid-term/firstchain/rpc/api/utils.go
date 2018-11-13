package api

import (
	"firstchain/common"
	dbtypes "firstchain/basic"
	"firstchain/rpc/utils"
)

func convertBlock(blk *dbtypes.Block) *utils.Block {
	if blk == nil {
		return nil
	}
	var txs []*utils.Transaction
	for _, tx := range blk.Transactions {
		txs = append(txs, &utils.Transaction{
			Nonce:     tx.Nonce,
			GasPrices: tx.GasPrice,
			GasLimit:  tx.GasLimit,
			Value:     tx.Value,
			From:      tx.From.Hex(),
			To:        tx.To.Hex(),
			Payload:   tx.Payload,
		})
	}
	return &utils.Block{
		Header: utils.Header{
			ParentHash:    blk.ParentHash().Hex(),
			Height:        blk.Height(),
			StateRoot:     blk.StateRoot().Hex(),
			TxRoot:        blk.TxRoot().Hex(),
			ReceiptHash:   blk.ReceiptsHash().Hex(),
			Coinbase:      blk.Coinbase().Hex(),
			Time:          blk.Time(),
			GasUsed:       blk.GasUsed(),
			GasLimit:      blk.GasLimit(),
			Extra:         blk.Extra(),
			ConsensusInfo: blk.ConsensusInfo(),
		},
		Transactions: txs,
		Pubkey:       common.Bytes2Hex(blk.PubKey),
		Signature:    common.Bytes2Hex(blk.Signature),
	}
}

func convertHeader(header *dbtypes.Header) *utils.Header {
	if header == nil {
		return nil
	}
	return &utils.Header{
		ParentHash:    header.ParentHash.Hex(),
		Height:        header.Height,
		StateRoot:     header.StateRoot.Hex(),
		TxRoot:        header.TxRoot.Hex(),
		ReceiptHash:   header.ReceiptsHash.Hex(),
		Coinbase:      header.Coinbase.Hex(),
		Time:          header.Time,
		GasUsed:       header.GasUsed,
		GasLimit:      header.GasLimit,
		Extra:         header.Extra,
		ConsensusInfo: header.ConsensusInfo,
	}
}

func convertTransaction(tx *dbtypes.Transaction) *utils.Transaction {
	if tx == nil {
		return nil
	}
	return &utils.Transaction{
		Nonce:     tx.Nonce,
		GasPrices: tx.GasPrice,
		GasLimit:  tx.GasLimit,
		Value:     tx.Value,
		From:      tx.From.Hex(),
		To:        tx.To.Hex(),
		Payload:   tx.Payload,
		PubKey:    common.Bytes2Hex(tx.PubKey),
		Signature: common.Bytes2Hex(tx.Signature),
	}
}

// convertReceipt convert the field of `hash` type in Receipt to human-friendly hex format
func convertReceipt(rp *dbtypes.Receipt) *utils.Receipt {
	if rp == nil {
		return nil
	}
	receipt := &utils.Receipt{
		PostState:       rp.PostState.Hex(),
		Status:          rp.Status,
		TxHash:          rp.TxHash.Hex(),
		ContractAddress: rp.ContractAddress.Hex(),
		GasUsed:         rp.GasUsed,
	}

	for _, log := range rp.Logs {
		l := &utils.Log{
			Address:     log.Address.Hex(),
			Data:        log.Data,
			BlockHeight: log.BlockHeight,
			TxHash:      log.TxHash.Hex(),
			TxIndex:     log.TxIndex,
			BlockHash:   log.BlockHash.Hex(),
			Index:       log.Index,
			Removed:     log.Removed,
		}

		for _, topic := range log.Topics {
			l.Topics = append(l.Topics, topic.Hex())
		}

		receipt.Logs = append(receipt.Logs, l)
	}
	return receipt
}

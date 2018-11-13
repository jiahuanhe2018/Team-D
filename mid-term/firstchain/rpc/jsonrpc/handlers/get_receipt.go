package handlers

import (
	"context"
	"github.com/intel-go/fastjson"
	"github.com/osamingo/jsonrpc"
	"firstchain/common"
	"firstchain/rpc/api"
	"firstchain/rpc/utils"
)

type getReceiptParams struct {
	TxHash string `json:"tx_hash"`
}

type getReceiptResult struct {
	Receipt utils.Receipt
}

type GetReceiptHandler struct {
	Api *api.TransactionAPI
}

func (h GetReceiptHandler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	var p getReceiptParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	txHash := common.HexToHash(p.TxHash)
	receipt := h.Api.GetReceipt(txHash)
	if receipt == nil {
		return nil, utils.ErrNotFound("receipt not found")
	}

	return *receipt, nil
}

func (h GetReceiptHandler) Name() string {
	return "GetReceipt"
}

func (h GetReceiptHandler) Params() interface{} {
	return getReceiptParams{}
}

func (h GetReceiptHandler) Result() interface{} {
	return getReceiptResult{}
}
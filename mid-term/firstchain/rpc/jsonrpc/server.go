package jsonrpc

import (
	"github.com/osamingo/jsonrpc"
	"net/http"
	"firstchain/rpc/api"
	"firstchain/rpc/jsonrpc/handlers"
	"firstchain/boot"
)

type Handler interface {
	jsonrpc.Handler
	Name() string
	Params() interface{}
	Result() interface{}
}

func InitHandler(T *boot.Small) []Handler {
	chainApi := &api.ChainAPI{T}
	txApi := &api.TransactionAPI{T}
	return []Handler{
		handlers.GetBlockHandler{chainApi},
		handlers.GetBlockHashHandler{chainApi},
		handlers.GetHeaderHandler{chainApi},
		handlers.GetTxHandler{txApi},
		handlers.GetReceiptHandler{txApi},
		handlers.SendTxHandler{txApi},
	}
}

func StartJsonRPCServer(T *boot.Small) error {
	mr := jsonrpc.NewMethodRepository()

	for _, s := range InitHandler(T) {
		mr.RegisterMethod(s.Name(), s, s.Params(), s.Result())
	}

	http.Handle("/", mr)
	http.HandleFunc("/debug", mr.ServeDebug)

	if err := http.ListenAndServe(":8081", http.DefaultServeMux); err != nil {
		return err
	}
	return nil
}

package blockchain


type Transaction struct {
	Amount    uint64    `json:"amount"`
	Recipient string `json:"recipient"`
	Sender    string `json:"sender"`
	Data      []byte `json:"data"`
}

func NewTransaction(sender string, recipient string, amount uint64, data []byte) *Transaction {
	transaction := new(Transaction)
	transaction.Sender = sender
	transaction.Recipient = recipient
	transaction.Amount = amount
	transaction.Data = data

	return transaction
}

type TxPool struct {
	AllTx     []Transaction
}

func NewTxPool() *TxPool {
	return &TxPool{
		AllTx:   make([]Transaction, 0),
	}
}

func (p *TxPool)Clear() bool {
	if len(p.AllTx) == 0 {
		return true
	}
	p.AllTx = make([]Transaction, 0)
	return true
}
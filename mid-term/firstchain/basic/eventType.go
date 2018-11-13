package basic

 

/*
	Blockchain events
 */

/*
	Block events
 */
type NewBlockEvent struct {
	Block *Block
}

type BlockBroadcastEvent struct{}

type ExecBlockEvent struct {
	Block *Block
}

type ExecFinishEvent struct {
	Res bool // exec result.If success,set true
}

// BlockReadyEvent will be post after block pool received a block and store into pool.
type BlockReadyEvent struct {
	Block *Block
}

type ProposeBlockEvent struct {
	Block *Block
}

// ConsensusEvent will be posted after a new block proposed by the BP
// completed execution without errors
type ConsensusEvent struct {
	Block    *Block
	Receipts Receipts
}

type CommitBlockEvent struct {
	Block *Block
}

type CommitCompleteEvent struct {
	Block *Block
}

// NewReceiptsEvent will be posted after a block from other nodes come in,
// completed execution without errors and passed verification.
type NewReceiptsEvent struct {
	Block    *Block
	Receipts Receipts
}

// ErrOccurEvent will be posted when some errors occur during executor processing.
type ErrOccurEvent struct {
	Err error
}

/*
	Transaction events
 */
type NewTxEvent struct {
	Tx *Transaction
}

type NewTxsEvent struct {
	Txs Transactions
}

type ExecPendingTxEvent struct {
	Txs Transactions
}

type TxBroadcastEvent struct{}

type RollbackEvent struct{}

package basic

import (
	"strconv"

	"firstchain/common"
	//"firstchain/types"
)

/*
	** Hash is block header hash


	"LastHeader" => the latest block header hash
	"LastBlock" => the latest block hash
	"WorldState" => the latest world state root hash

	"h" + block height + "n" => block hash
	"h" + block height + block hash => header
	"H" + block hash => block height
	"b" + block height + block hash => block
	"r" + txHash => block receipt
	"l" + txHash => transaction meta data {hash,height,txIndex}
*/

const (
	KeyLastHeader = "LastHeader"
	KeyLastBlock  = "LastBlock"
)

// SmallDB stores and manages blockchain data
type SmallDB struct {
	db Database
}

func NewSmallDB(db Database) *SmallDB {
	return &SmallDB{db}
}

func (tdb *SmallDB) LDB() Database {
	return tdb.db
}

func (tdb *SmallDB) GetLastBlock() (common.Hash, error) {
	data, err := tdb.db.Get([]byte(KeyLastBlock))
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}

func (tdb *SmallDB) PutLastBlock(batch Batch, hash common.Hash, sync, flush bool) error {
	if err := batch.Put([]byte(KeyLastBlock), hash.Bytes()); err != nil {
		return err
	}
	if flush {
		if sync {
			return batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

func (tdb *SmallDB) GetLastHeader() (common.Hash, error) {
	data, err := tdb.db.Get([]byte(KeyLastHeader))
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}

func (tdb *SmallDB) PutLastHeader(batch Batch, hash common.Hash, sync, flush bool) error {
	if err := batch.Put([]byte(KeyLastHeader), hash.Bytes()); err != nil {
		return err
	}
	if flush {
		if sync {
			return batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

func (tdb *SmallDB) GetHash(height uint64) (common.Hash, error) {
	var hash common.Hash
	data, err := tdb.db.Get([]byte("h" + strconv.FormatUint(height, 10) + "n"))
	if err != nil {
		return hash, err
	}
	//hash = common.DecodeHash(data)
	return common.BytesToHash(data), nil
}

func (tdb *SmallDB) PutHash(batch Batch, height uint64, hash common.Hash, sync, flush bool) error {
	if err := batch.Put([]byte("h"+strconv.FormatUint(height, 10)+"n"), hash[:]); err != nil {
		return err
	}
	if flush {
		if sync {
			return batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

func (tdb *SmallDB) GetHeader(height uint64, hash common.Hash) (*Header, error) {
	data, err := tdb.db.Get([]byte("h" + strconv.FormatUint(height, 10) + hash.String()))
	if err != nil {
		return nil, err
	}
	header := Header{}
	header.Desrialize(data)
	return &header, nil
}
func (tdb *SmallDB) PutHeader(batch Batch, header *Header, sync, flush bool) error {
	data, _ := header.Serialize()
	if err := batch.Put([]byte("h"+strconv.FormatUint(header.Height, 10)+header.Hash().String()), data); err != nil {
		return err
	}
	if flush {
		if sync {
			return batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

func (tdb *SmallDB) GetHeight(hash common.Hash) (uint64, error) {
	data, err := tdb.db.Get([]byte("H" + hash.String()))
	if err != nil {
		return 0, err
	}
	return common.Bytes2Uint(data), nil
}

func (tdb *SmallDB) PutHeight(batch Batch, hash common.Hash, height uint64, sync, flush bool) error {
	if err := batch.Put([]byte("H"+hash.String()), common.Uint2Bytes(height)); err != nil {
		return err
	}
	if flush {
		if sync {
			return batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

func (tdb *SmallDB) GetBlock(height uint64, hash common.Hash) (*Block, error) {
	data, err := tdb.db.Get([]byte("b" + strconv.FormatUint(height, 10) + hash.String()))
	if err != nil {
		return nil, err
	}
	block := Block{}
	block.Deserialize(data)
	return &block, nil
}

func (tdb *SmallDB) PutBlock(batch Batch, block *Block, sync, flush bool) error {
	height := block.Height()
	hash := block.Hash()
	data, _ := block.Serialize()
	if err := batch.Put([]byte("b"+strconv.FormatUint(height, 10)+hash.String()), data); err != nil {
		return err
	}
	if flush {
		if sync {
			return batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

func (tdb *SmallDB) GetReceipt(txHash common.Hash) (*Receipt, error) {
	data, err := tdb.db.Get([]byte("r" + txHash.String()))
	if err != nil {
		return nil, err
	}
	var receipt Receipt
	err = receipt.Deserialize(data)
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (tdb *SmallDB) PutReceipt(batch Batch, txHash common.Hash, receipt *Receipt, sync, flush bool) error {
	data, err := receipt.Serialize()
	if err != nil {
		return err
	}
	if err := batch.Put([]byte("r"+txHash.String()), data); err != nil {
		return err
	}

	if flush {
		if sync {
			return batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

func (tdb *SmallDB) GetTxMeta(txHash common.Hash) (*TxMeta, error) {
	data, err := tdb.db.Get([]byte("l" + txHash.String()))
	if err != nil {
		log.Errorf("Cannot find txMeta with txHash %s", txHash.Hex())
		return nil, err
	}
	txMeta := &TxMeta{}
	txMeta.Deserialize(data)
	return txMeta, nil
}

// PutTxMetas put transactions' meta to db in batch
func (tdb *SmallDB) PutTxMetas(batch Batch, txs Transactions, hash common.Hash, height uint64, sync, flush bool) error {
	for i, tx := range txs {
		txMeta := &TxMeta{
			Hash:    hash,
			Height:  height,
			TxIndex: uint64(i),
		}
		data, _ := txMeta.Serialize()
		if err := batch.Put([]byte("l"+tx.Hash().String()), data); err != nil {
			return err
		}
	}
	if flush {
		if sync {
			return batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

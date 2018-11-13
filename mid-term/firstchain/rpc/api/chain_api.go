package api

import (
	"firstchain/common"
	"firstchain/rpc/utils"
	"firstchain/boot"
)

type ChainAPI struct {
	SC *boot.Small
}

func (api *ChainAPI) GetBlock(hash common.Hash, height uint64) *utils.Block {
	blk := api.SC.Chain().GetBlock(hash, height)
	if blk == nil {
		return nil
	}
	return convertBlock(blk)
}

func (api *ChainAPI) GetBlockHash(height uint64) common.Hash {
	return api.SC.Chain().GetHash(height)
}

func (api *ChainAPI) GetBlockByHash(hash common.Hash) *utils.Block {
	blk := api.SC.Chain().GetBlockByHash(hash)
	return convertBlock(blk)
}

func (api *ChainAPI) GetBlockByHeight(height uint64) *utils.Block {
	blk := api.SC.Chain().GetBlockByHeight(height)
	return convertBlock(blk)
}

func (api *ChainAPI) GetHeader(hash common.Hash, height uint64) *utils.Header {
	header := api.SC.Chain().GetHeader(hash, height)
	return convertHeader(header)
}

func (api *ChainAPI) GetHeaderByHash(hash common.Hash) *utils.Header {
	height, err := api.SC.DB().GetHeight(hash)
	if err != nil {
		return nil
	}
	return api.GetHeader(hash, height)
}

// Height returns the latest block height
func (api *ChainAPI) Height() uint64 {
	return api.SC.Chain().LastHeight()
}

// Hash returns the latest block hash
func (api *ChainAPI) Hash() common.Hash {
	return api.SC.Chain().LastBlock().Hash()
}


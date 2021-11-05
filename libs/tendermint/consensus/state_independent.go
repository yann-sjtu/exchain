package consensus

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/okex/exchain/libs/tendermint/types"
)

var (
	errorNotExist     = errors.New("Target not exist")
	errorAlreadyExist = errors.New("Target Already exist")
	errorInvalidParam = errors.New("Invalid params")
)

type blockBuffer struct {
	proposal          *types.Proposal
	proposalBlockPart *types.PartSet
	proposalBlock     *types.Block
	complete          bool
}

type BlocksManager struct {
	//height => blockBuffer
	blocksBuffer map[int64]*blockBuffer
	param        *types.BlockParams
	mutex        sync.RWMutex //may not use
}

func NewBlockManager() *BlocksManager {
	return &BlocksManager{blocksBuffer: make(map[int64]*blockBuffer, 0)}
}

func (b *BlocksManager) AddProposal(p *types.Proposal) error {
	if p == nil {
		return errorInvalidParam
	}

	_, ok := b.blocksBuffer[p.Height]
	if ok {
		return errorAlreadyExist
	}
	b.blocksBuffer[p.Height] = &blockBuffer{
		proposal:          p,
		proposalBlock:     nil,
		proposalBlockPart: types.NewPartSetFromHeader(p.BlockID.PartsHeader),
		complete:          false,
	}
	return nil
}

func (b *BlocksManager) AddBlockPart(height int64, part *types.Part) error {
	buf, ok := b.blocksBuffer[height]
	if !ok || buf.proposalBlockPart == nil {
		return errorNotExist
	}
	added, err := buf.proposalBlockPart.AddPart(part)
	if err != nil {
		return errorNotExist
	}
	if added && buf.proposalBlockPart.IsComplete() {
		_, err = cdc.UnmarshalBinaryLengthPrefixedReader(
			buf.proposalBlockPart.GetReader(),
			&buf.proposalBlock,
			b.param.MaxBytes,
		)
		if err != nil {
			return err
		}
		//mark block as completed
		buf.complete = true
	}
	return nil
}

func (b *BlocksManager) GetBlock(height int64) *types.Block {
	buf, ok := b.blocksBuffer[height]
	if !ok {
		return nil
	}
	if buf.complete {
		return buf.proposalBlock
	}
	return nil
}

func (b *BlocksManager) GetProposal(height int64) *types.Proposal {
	buf, ok := b.blocksBuffer[height]
	if !ok {
		return nil
	}
	return buf.proposal //whatever it is nil or not
}

func (b *BlocksManager) EraseBlock(height int64) {
	b.blocksBuffer[height] = nil
}

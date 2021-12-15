package types

import (
	"bytes"

	"github.com/okex/exchain/libs/tendermint/libs/kv"

	"github.com/tendermint/go-amino"
)

func MarshalPubKeyToAmino(pubkey PubKey) ([]byte, error) {
	var buf bytes.Buffer
	fieldKeysType := [2]byte{1<<3 | 2, 2<<3 | 2}
	for pos := 1; pos <= 2; pos++ {
		switch pos {
		case 1:
			if pubkey.Type == "" {
				break
			}
			err := buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return nil, err
			}
			err = amino.EncodeStringToBuffer(&buf, pubkey.Type)
			if err != nil {
				return nil, err
			}
		case 2:
			if len(pubkey.Data) == 0 {
				break
			}
			err := buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return nil, err
			}
			err = amino.EncodeByteSliceToBuffer(&buf, pubkey.Data)
			if err != nil {
				return nil, err
			}
		default:
			panic("unreachable")
		}
	}
	return buf.Bytes(), nil
}

func MarshalValidatorUpdateToAmino(valUpdate ValidatorUpdate) ([]byte, error) {
	var buf bytes.Buffer
	var err error
	fieldKeysType := [2]byte{1<<3 | 2, 2 << 3}
	for pos := 1; pos <= 2; pos++ {
		lBeforeKey := buf.Len()
		var noWrite bool

		switch pos {
		case 1:
			var data []byte
			data, err = MarshalPubKeyToAmino(valUpdate.PubKey)
			if err != nil {
				return nil, err
			}
			if len(data) == 0 {
				noWrite = true
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return nil, err
			}
			err = amino.EncodeByteSliceToBuffer(&buf, data)
			if err != nil {
				return nil, err
			}
		case 2:
			if valUpdate.Power == 0 {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return nil, err
			}
			err = amino.EncodeUvarintToBuffer(&buf, uint64(valUpdate.Power))
			if err != nil {
				return nil, err
			}
		default:
			panic("unreachable")
		}

		if noWrite {
			buf.Truncate(lBeforeKey)
		}
	}
	return buf.Bytes(), nil
}

func MarshalBlockParamsToAmino(params BlockParams) ([]byte, error) {
	var buf bytes.Buffer
	fieldKeysType := [2]byte{1 << 3, 2 << 3}
	for pos := 1; pos <= 2; pos++ {
		lBeforeKey := buf.Len()
		var noWrite bool
		err := buf.WriteByte(fieldKeysType[pos-1])
		if err != nil {
			return nil, err
		}
		switch pos {
		case 1:
			if params.MaxBytes == 0 {
				noWrite = true
				break
			}
			err = amino.EncodeUvarintToBuffer(&buf, uint64(params.MaxBytes))
			if err != nil {
				return nil, err
			}
		case 2:
			if params.MaxGas == 0 {
				noWrite = true
				break
			}
			err = amino.EncodeUvarintToBuffer(&buf, uint64(params.MaxGas))
			if err != nil {
				return nil, err
			}
		default:
			panic("unreachable")
		}

		if noWrite {
			buf.Truncate(lBeforeKey)
		}
	}
	return buf.Bytes(), nil
}

func MarshalEvidenceParamsToAmino(params EvidenceParams) ([]byte, error) {
	var buf bytes.Buffer
	fieldKeysType := [2]byte{1 << 3, 2 << 3}
	for pos := 1; pos <= 2; pos++ {
		lBeforeKey := buf.Len()
		var noWrite bool
		err := buf.WriteByte(fieldKeysType[pos-1])
		if err != nil {
			return nil, err
		}
		switch pos {
		case 1:
			if params.MaxAgeNumBlocks == 0 {
				noWrite = true
				break
			}
			err = amino.EncodeUvarintToBuffer(&buf, uint64(params.MaxAgeNumBlocks))
			if err != nil {
				return nil, err
			}
		case 2:
			if params.MaxAgeDuration == 0 {
				noWrite = true
				break
			}
			err = amino.EncodeUvarintToBuffer(&buf, uint64(params.MaxAgeDuration))
			if err != nil {
				return nil, err
			}
		default:
			panic("unreachable")
		}

		if noWrite {
			buf.Truncate(lBeforeKey)
		}
	}
	return buf.Bytes(), nil
}

func MarshalValidatorParamsToAmino(params ValidatorParams) ([]byte, error) {
	var buf bytes.Buffer
	var err error
	fieldKeysType := [1]byte{1<<3 | 2}
	for pos := 1; pos <= 1; pos++ {
		switch pos {
		case 1:
			if len(params.PubKeyTypes) == 0 {
				break
			}
			for i := 0; i < len(params.PubKeyTypes); i++ {
				err = buf.WriteByte(fieldKeysType[pos-1])
				if err != nil {
					return nil, err
				}
				err = amino.EncodeStringToBuffer(&buf, params.PubKeyTypes[i])
				if err != nil {
					return nil, err
				}
			}
		default:
			panic("unreachable")
		}
	}
	return buf.Bytes(), nil
}

func MarshalEventToAmino(event Event) ([]byte, error) {
	var buf bytes.Buffer
	var err error
	fieldKeysType := [2]byte{1<<3 | 2, 2<<3 | 2}
	for pos := 1; pos <= 2; pos++ {
		switch pos {
		case 1:
			if event.Type == "" {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return nil, err
			}
			err = amino.EncodeUvarintToBuffer(&buf, uint64(len(event.Type)))
			if err != nil {
				return nil, err
			}
			_, err = buf.WriteString(event.Type)
			if err != nil {
				return nil, err
			}
		case 2:
			for i := 0; i < len(event.Attributes); i++ {
				err = buf.WriteByte(fieldKeysType[pos-1])
				if err != nil {
					return nil, err
				}
				data, err := kv.MarshalPairToAmino(event.Attributes[i])
				if err != nil {
					return nil, err
				}
				err = amino.EncodeByteSliceToBuffer(&buf, data)
				if err != nil {
					return nil, err
				}
			}
		default:
			panic("unreachable")
		}
	}
	return buf.Bytes(), nil
}

func (event Event) AminoSize() int {
	var size int
	if event.Type != "" {
		size += 1 + amino.EncodedStringSize(event.Type)
	}
	for _, attr := range event.Attributes {
		attrSize := attr.AminoSize()
		size += 1 + amino.UvarintSize(uint64(attrSize)) + attrSize
	}
	return size
}

func (event Event) MarshalToAmino() ([]byte, error) {
	return event.marshalToAminoWithSizeCompute2()
}

func (event Event) marshalToAmino() ([]byte, error) {
	var buf = &bytes.Buffer{}
	err := event.MarshalToAminoToBuffer(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (event Event) marshalToAminoWithSizeCompute() ([]byte, error) {
	var buf = &bytes.Buffer{}
	buf.Grow(event.AminoSize())
	err := event.marshalToAminoToBuffer(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (event Event) marshalToAminoWithSizeCompute2() ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(event.AminoSize())
	err := event.marshalToAminoToBuffer2(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var eventBufferPool = amino.NewBufferPool()

func (event Event) marshalToAminoWithPool() ([]byte, error) {
	var buf = eventBufferPool.Get()
	defer eventBufferPool.Put(buf)
	buf.Grow(event.AminoSize())
	err := event.MarshalToAminoToBuffer(buf)
	if err != nil {
		return nil, err
	}
	return amino.GetBytesBufferCopy(buf), nil
}

func (event Event) MarshalToAminoToBuffer(buf *bytes.Buffer) error {
	return event.marshalToAminoToBuffer2(buf)
}

func (event Event) marshalToAminoToBuffer(buf *bytes.Buffer) error {
	var err error
	fieldKeysType := [2]byte{1<<3 | 2, 2<<3 | 2}
	for pos := 1; pos <= 2; pos++ {
		switch pos {
		case 1:
			if event.Type == "" {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return err
			}
			err = amino.EncodeUvarintToBuffer(buf, uint64(len(event.Type)))
			if err != nil {
				return err
			}
			_, err = buf.WriteString(event.Type)
			if err != nil {
				return err
			}
		case 2:
			for i := 0; i < len(event.Attributes); i++ {
				err = buf.WriteByte(fieldKeysType[pos-1])
				if err != nil {
					return err
				}
				data, err := kv.MarshalPairToAmino(event.Attributes[i])
				if err != nil {
					return err
				}
				err = amino.EncodeByteSliceToBuffer(buf, data)
				if err != nil {
					return err
				}
			}
		default:
			panic("unreachable")
		}
	}
	return nil
}

func (event Event) marshalToAminoToBuffer2(buf *bytes.Buffer) error {
	var err error
	fieldKeysType := [2]byte{1<<3 | 2, 2<<3 | 2}
	for pos := 1; pos <= 2; pos++ {
		switch pos {
		case 1:
			if event.Type == "" {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return err
			}
			err = amino.EncodeUvarintToBuffer(buf, uint64(len(event.Type)))
			if err != nil {
				return err
			}
			_, err = buf.WriteString(event.Type)
			if err != nil {
				return err
			}
		case 2:
			for i := 0; i < len(event.Attributes); i++ {
				err = buf.WriteByte(fieldKeysType[pos-1])
				if err != nil {
					return err
				}

				_size := event.Attributes[i].AminoSize()
				lenAfterKey := buf.Len()
				err = amino.EncodeUvarintToBuffer(buf, uint64(_size))
				lenBeforeValue := buf.Len()
				err = event.Attributes[i].MarshalToAminoToBuffer(buf)
				if err != nil {
					return err
				}
				lenAfterValue := buf.Len()
				if lenAfterValue-lenBeforeValue != _size {
					buf.Truncate(lenAfterKey)
					data, err := kv.MarshalPairToAmino(event.Attributes[i])
					if err != nil {
						return err
					}
					err = amino.EncodeByteSliceToBuffer(buf, data)
					if err != nil {
						return err
					}
				}
			}
		default:
			panic("unreachable")
		}
	}
	return nil
}

var responseDeliverTxBufferPool = amino.NewBufferPool()

func MarshalResponseDeliverTxToAmino(tx *ResponseDeliverTx) ([]byte, error) {
	if tx == nil {
		return nil, nil
	}
	var buf = responseDeliverTxBufferPool.Get()
	defer responseDeliverTxBufferPool.Put(buf)
	fieldKeysType := [8]byte{1 << 3, 2<<3 | 2, 3<<3 | 2, 4<<3 | 2, 5 << 3, 6 << 3, 7<<3 | 2, 8<<3 | 2}
	for pos := 1; pos <= 8; pos++ {
		lBeforeKey := buf.Len()
		var noWrite bool
		err := buf.WriteByte(fieldKeysType[pos-1])
		if err != nil {
			return nil, err
		}
		switch pos {
		case 1:
			if tx.Code == 0 {
				noWrite = true
				break
			}
			err = amino.EncodeUvarintToBuffer(buf, uint64(tx.Code))
			if err != nil {
				return nil, err
			}
		case 2:
			if len(tx.Data) == 0 {
				noWrite = true
				break
			}
			err = amino.EncodeByteSliceToBuffer(buf, tx.Data)
			if err != nil {
				return nil, err
			}
		case 3:
			if tx.Log == "" {
				noWrite = true
				break
			}
			err = amino.EncodeStringToBuffer(buf, tx.Log)
			if err != nil {
				return nil, err
			}
		case 4:
			if tx.Info == "" {
				noWrite = true
				break
			}
			err = amino.EncodeStringToBuffer(buf, tx.Info)
			if err != nil {
				return nil, err
			}
		case 5:
			if tx.GasWanted == 0 {
				noWrite = true
				break
			}
			err = amino.EncodeUvarintToBuffer(buf, uint64(tx.GasWanted))
			if err != nil {
				return nil, err
			}
		case 6:
			if tx.GasUsed == 0 {
				noWrite = true
				break
			}
			err = amino.EncodeUvarintToBuffer(buf, uint64(tx.GasUsed))
			if err != nil {
				return nil, err
			}
		case 7:
			eventsLen := len(tx.Events)
			if eventsLen == 0 {
				noWrite = true
				break
			}
			data, err := tx.Events[0].MarshalToAmino()
			if err != nil {
				return nil, err
			}
			err = amino.EncodeByteSliceToBuffer(buf, data)
			if err != nil {
				return nil, err
			}
			for i := 1; i < eventsLen; i++ {
				err = buf.WriteByte(fieldKeysType[pos-1])
				if err != nil {
					return nil, err
				}
				data, err = tx.Events[i].MarshalToAmino()
				if err != nil {
					return nil, err
				}
				err = amino.EncodeByteSliceToBuffer(buf, data)
				if err != nil {
					return nil, err
				}
			}
		case 8:
			if tx.Codespace == "" {
				noWrite = true
				break
			}
			err = amino.EncodeStringToBuffer(buf, tx.Codespace)
			if err != nil {
				return nil, err
			}
		default:
			panic("unreachable")
		}

		if noWrite {
			buf.Truncate(lBeforeKey)
		}
	}
	return amino.GetBytesBufferCopy(buf), nil
}

func MarshalResponseBeginBlockToAmino(beginBlock *ResponseBeginBlock) ([]byte, error) {
	if beginBlock == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	fieldKey := byte(1<<3 | 2)
	for i := 0; i < len(beginBlock.Events); i++ {
		err := buf.WriteByte(fieldKey)
		if err != nil {
			return nil, err
		}
		data, err := MarshalEventToAmino(beginBlock.Events[i])
		if err != nil {
			return nil, err
		}
		err = amino.EncodeByteSliceToBuffer(&buf, data)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (beginBlock ResponseBeginBlock) MarshalToAmino() ([]byte, error) {
	return beginBlock.marshalToAminoWithSizeCompute2()
}

func (beginBlock ResponseBeginBlock) AminoSize() int {
	var size int
	for _, event := range beginBlock.Events {
		_size := event.AminoSize()
		size += 1 + amino.UvarintSize(uint64(_size)) + _size
	}
	return size
}

func (beginBlock ResponseBeginBlock) marshalToAmino() ([]byte, error) {
	var buf = &bytes.Buffer{}
	err := beginBlock.MarshalToAminoToBuffer(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (beginBlock ResponseBeginBlock) marshalToAminoWithSizeCompute() ([]byte, error) {
	var buf = &bytes.Buffer{}
	buf.Grow(beginBlock.AminoSize())
	err := beginBlock.MarshalToAminoToBuffer(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (beginBlock ResponseBeginBlock) marshalToAminoWithSizeCompute2() ([]byte, error) {
	var buf = &bytes.Buffer{}
	buf.Grow(beginBlock.AminoSize())
	fieldKey := byte(1<<3 | 2)
	for i := 0; i < len(beginBlock.Events); i++ {
		err := buf.WriteByte(fieldKey)
		if err != nil {
			return nil, err
		}
		lenAfterKey := buf.Len()
		event := beginBlock.Events[i]
		eventSize := event.AminoSize()
		err = amino.EncodeUvarintToBuffer(buf, uint64(eventSize))
		if err != nil {
			return nil, err
		}
		lenBeforeValue := buf.Len()
		err = event.MarshalToAminoToBuffer(buf)
		if err != nil {
			return nil, err
		}
		lenAfterValue := buf.Len()
		if lenAfterValue-lenBeforeValue != eventSize {
			buf.Truncate(lenAfterKey)
			data, err := event.MarshalToAmino()
			if err != nil {
				return nil, err
			}
			err = amino.EncodeByteSliceToBuffer(buf, data)
			if err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}

func (beginBlock ResponseBeginBlock) MarshalToAminoToBuffer(buf *bytes.Buffer) error {
	fieldKey := byte(1<<3 | 2)
	for i := 0; i < len(beginBlock.Events); i++ {
		err := buf.WriteByte(fieldKey)
		if err != nil {
			return err
		}
		data, err := beginBlock.Events[i].MarshalToAmino()
		if err != nil {
			return err
		}
		err = amino.EncodeByteSliceToBuffer(buf, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func MarshalConsensusParamsToAmino(params ConsensusParams) (data []byte, err error) {
	var buf bytes.Buffer
	fieldKeysType := [3]byte{1<<3 | 2, 2<<3 | 2, 3<<3 | 2}
	for pos := 1; pos <= 3; pos++ {
		lBeforeKey := buf.Len()
		var noWrite bool
		err = buf.WriteByte(fieldKeysType[pos-1])
		if err != nil {
			return nil, err
		}
		switch pos {
		case 1:
			if params.Block == nil {
				noWrite = true
				break
			}
			data, err = MarshalBlockParamsToAmino(*params.Block)
			if err != nil {
				return nil, err
			}
			err = amino.EncodeByteSliceToBuffer(&buf, data)
			if err != nil {
				return nil, err
			}
		case 2:
			if params.Evidence == nil {
				noWrite = true
				break
			}
			data, err = MarshalEvidenceParamsToAmino(*params.Evidence)
			if err != nil {
				return nil, err
			}
			err = amino.EncodeByteSliceToBuffer(&buf, data)
			if err != nil {
				return nil, err
			}
		case 3:
			if params.Validator == nil {
				noWrite = true
				break
			}
			data, err = MarshalValidatorParamsToAmino(*params.Validator)
			if err != nil {
				return nil, err
			}
			err = amino.EncodeByteSliceToBuffer(&buf, data)
			if err != nil {
				return nil, err
			}
		default:
			panic("unreachable")
		}

		if noWrite {
			buf.Truncate(lBeforeKey)
		}
	}
	return buf.Bytes(), nil
}

func MarshalResponseEndBlockToAmino(endBlock *ResponseEndBlock) ([]byte, error) {
	if endBlock == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	var err error
	fieldKeysType := [3]byte{1<<3 | 2, 2<<3 | 2, 3<<3 | 2}
	for pos := 1; pos <= 3; pos++ {
		switch pos {
		case 1:
			if len(endBlock.ValidatorUpdates) == 0 {
				break
			}
			for i := 0; i < len(endBlock.ValidatorUpdates); i++ {
				err = buf.WriteByte(fieldKeysType[pos-1])
				if err != nil {
					return nil, err
				}
				data, err := MarshalValidatorUpdateToAmino(endBlock.ValidatorUpdates[0])
				if err != nil {
					return nil, err
				}
				err = amino.EncodeByteSliceToBuffer(&buf, data)
				if err != nil {
					return nil, err
				}
			}
		case 2:
			if endBlock.ConsensusParamUpdates == nil {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return nil, err
			}
			data, err := MarshalConsensusParamsToAmino(*endBlock.ConsensusParamUpdates)
			if err != nil {
				return nil, err
			}
			err = amino.EncodeByteSlice(&buf, data)
			if err != nil {
				return nil, err
			}
		case 3:
			eventsLen := len(endBlock.Events)
			if eventsLen == 0 {
				break
			}
			for i := 0; i < eventsLen; i++ {
				err = buf.WriteByte(fieldKeysType[pos-1])
				if err != nil {
					return nil, err
				}
				data, err := MarshalEventToAmino(endBlock.Events[i])
				if err != nil {
					return nil, err
				}
				err = amino.EncodeByteSlice(&buf, data)
				if err != nil {
					return nil, err
				}
			}
		default:
			panic("unreachable")
		}
	}
	return buf.Bytes(), nil
}

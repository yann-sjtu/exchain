package types

import (
	"testing"
	"time"

	"github.com/okex/exchain/libs/tendermint/libs/kv"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

var testEvents = []Event{
	{},
	{
		Type: "test",
	},
	{
		Attributes: []kv.Pair{
			{Key: []byte("key"), Value: []byte("value")},
			{Key: []byte("key2"), Value: []byte("value2")},
		},
	},
	{
		Type: "test",
		Attributes: []kv.Pair{
			{Key: []byte("key"), Value: []byte("value")},
			{Key: []byte("key2"), Value: []byte("value2")},
			{},
		},
	},
	{
		Type: "longTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongType",
		Attributes: []kv.Pair{
			{
				Key:   []byte("longKeylongKeylongKeylongKeylongKeylongKeylongKeylongKeylongKeylongKeylongKeylongKeylongKeylongKeylongKeylongKeylongKeylongKeylongKey"),
				Value: []byte("longvaluelongvaluelongvaluelongvaluelongvaluelongvaluelongvaluelongvaluelongvaluelongvaluelongvaluelongvaluelongvaluelongvaluelongvaluelongvalue"),
			},
			{},
		},
	},
	{
		Attributes: []kv.Pair{},
	},
	{
		XXX_unrecognized:     []byte{0x01, 0x02, 0x03},
		XXX_sizecache:        10,
		XXX_NoUnkeyedLiteral: struct{}{},
	},
}

func TestEventAmino(t *testing.T) {
	for _, event := range testEvents {
		expect, err := cdc.MarshalBinaryBare(event)
		require.NoError(t, err)

		actual, err := MarshalEventToAmino(event)
		require.NoError(t, err)
		require.EqualValues(t, expect, actual)

		bz1, err := event.marshalToAmino()
		require.NoError(t, err)

		bz2, err := event.marshalToAminoWithPool()
		require.NoError(t, err)

		bz3, err := event.marshalToAminoWithSizeCompute()
		require.NoError(t, err)

		bz4, err := event.marshalToAminoWithSizeCompute2()
		require.NoError(t, err)

		require.EqualValues(t, expect, bz1)
		require.EqualValues(t, expect, bz2)
		require.EqualValues(t, expect, bz3)
		require.EqualValues(t, expect, bz4)

		require.EqualValues(t, len(expect), event.AminoSize())
	}
}

func BenchmarkEventAmino(b *testing.B) {
	b.Run("amino", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, event := range testEvents {
				_, err := cdc.MarshalBinaryBare(event)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshaller", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, event := range testEvents {
				_, err := MarshalEventToAmino(event)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshall", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, event := range testEvents {
				_, err := event.marshalToAmino()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshall with size", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, event := range testEvents {
				_, err := event.marshalToAminoWithSizeCompute()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshall with pool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, event := range testEvents {
				_, err := event.marshalToAminoWithPool()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshall with size 2", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, event := range testEvents {
				_, err := event.marshalToAminoWithSizeCompute2()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

func TestPubKeyAmino(t *testing.T) {
	var pubkeys = []PubKey{
		{},
		{Type: "type"},
		{Data: []byte("testdata")},
		{
			Type: "test",
			Data: []byte("data"),
		},
	}

	for _, pubkey := range pubkeys {
		expect, err := cdc.MarshalBinaryBare(pubkey)
		require.NoError(t, err)

		actual, err := MarshalPubKeyToAmino(pubkey)
		require.NoError(t, err)
		require.EqualValues(t, expect, actual)
	}
}

func TestValidatorUpdateAmino(t *testing.T) {
	var validatorUpdates = []ValidatorUpdate{
		{},
		{
			PubKey: PubKey{
				Type: "test",
			},
		},
		{
			PubKey: PubKey{
				Type: "test",
				Data: []byte("data"),
			},
		},
		{
			Power: 100,
		},
		{
			PubKey: PubKey{
				Type: "test",
				Data: []byte("data"),
			},
			Power: 100,
		},
	}

	for _, validatorUpdate := range validatorUpdates {
		expect, err := cdc.MarshalBinaryBare(validatorUpdate)
		require.NoError(t, err)

		actual, err := MarshalValidatorUpdateToAmino(validatorUpdate)
		require.NoError(t, err)
		require.EqualValues(t, expect, actual)
	}
}

func TestBlockParamsAmino(t *testing.T) {
	tests := []BlockParams{
		{
			MaxBytes: 100,
			MaxGas:   200,
		},
		{
			MaxBytes: -100,
			MaxGas:   -200,
		},
	}

	for _, test := range tests {
		expect, err := cdc.MarshalBinaryBare(test)
		require.NoError(t, err)

		actual, err := MarshalBlockParamsToAmino(test)
		require.NoError(t, err)
		require.EqualValues(t, expect, actual)
	}
}

func TestEvidenceParamsAmino(t *testing.T) {
	tests := []EvidenceParams{
		{
			MaxAgeNumBlocks: 100,
			MaxAgeDuration:  1000 * time.Second,
		},
		{
			MaxAgeNumBlocks: -100,
			MaxAgeDuration:  time.Second,
		},
	}

	for _, test := range tests {
		expect, err := cdc.MarshalBinaryBare(test)
		require.NoError(t, err)

		actual, err := MarshalEvidenceParamsToAmino(test)
		require.NoError(t, err)
		require.EqualValues(t, expect, actual)
	}
}

func TestValidatorParamsAmino(t *testing.T) {
	tests := []ValidatorParams{
		{},
		{
			PubKeyTypes: []string{},
		},
		{
			PubKeyTypes: []string{""},
		},
		{
			PubKeyTypes: []string{"ed25519"},
		},
		{
			PubKeyTypes: []string{"ed25519", "ed25519"},
		},
	}

	for _, test := range tests {
		expect, err := cdc.MarshalBinaryBare(test)
		require.NoError(t, err)

		actual, err := MarshalValidatorParamsToAmino(test)
		require.NoError(t, err)
		require.EqualValues(t, expect, actual)
	}
}

func TestConsensusParamsAmino(t *testing.T) {
	tests := []ConsensusParams{
		{
			Block:     &BlockParams{},
			Evidence:  &EvidenceParams{},
			Validator: &ValidatorParams{},
		},
		{
			Block: &BlockParams{
				MaxBytes: 100,
			},
			Evidence: &EvidenceParams{
				MaxAgeDuration: 5 * time.Second,
			},
			Validator: &ValidatorParams{
				PubKeyTypes: nil,
			},
		},
		{
			Validator: &ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
		},
		{
			Block: &BlockParams{
				MaxBytes: 100,
				MaxGas:   200,
			},
			Evidence: &EvidenceParams{
				MaxAgeNumBlocks: 500,
				MaxAgeDuration:  6 * time.Second,
			},
			Validator: &ValidatorParams{
				PubKeyTypes: []string{},
			},
		},
	}

	for _, test := range tests {
		expect, err := cdc.MarshalBinaryBare(test)
		require.NoError(t, err)

		actual, err := MarshalConsensusParamsToAmino(test)
		require.NoError(t, err)
		require.EqualValues(t, expect, actual)
	}
}

func TestResponseDeliverTxAmino(t *testing.T) {
	var resps = []*ResponseDeliverTx{
		//{},
		{123, nil, "", "", 0, 0, nil, "", struct{}{}, nil, 0},
		{Code: 123, Data: []byte(""), Log: "log123", Info: "123info", GasWanted: 1234445, GasUsed: 98, Events: nil, Codespace: "sssdasf"},
		{Code: 0, Data: []byte("data"), Info: "info"},
		{Events: []Event{{}, {Type: "Event"}}},
	}

	for _, resp := range resps {
		expect, err := cdc.MarshalBinaryBare(resp)
		require.NoError(t, err)

		actual, err := MarshalResponseDeliverTxToAmino(resp)
		require.NoError(t, err)
		require.EqualValues(t, expect, actual)
	}
}

func TestResponseBeginBlockAmino(t *testing.T) {
	var testRespBeginBlock = []*ResponseBeginBlock{
		{},
		{
			Events: []Event{
				{
					Type: "longTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongType",
					Attributes: []kv.Pair{
						{Key: []byte("keykeykeykeykeykeykeykeykeykeykeykeykeykeykeykey"), Value: []byte("valuevaluevaluevaluevaluevaluevaluevaluevaluevaluevaluevaluevalue")},
					},
				},
			},
		},
		{
			Events: []Event{},
		},
		{
			Events: []Event{{}},
		},
		{
			XXX_unrecognized:     []byte{0x01, 0x02, 0x03},
			XXX_sizecache:        10,
			XXX_NoUnkeyedLiteral: struct{}{},
		},
	}
	for _, resp := range testRespBeginBlock {
		expect, err := cdc.MarshalBinaryBare(resp)
		require.NoError(t, err)

		actual, err := MarshalResponseBeginBlockToAmino(resp)
		require.NoError(t, err)
		require.EqualValues(t, expect, actual)

		bz1, err := resp.marshalToAmino()
		require.NoError(t, err)

		bz2, err := resp.marshalToAminoWithSizeCompute()
		require.NoError(t, err)

		bz3, err := resp.marshalToAminoWithSizeCompute2()
		require.NoError(t, err)

		require.EqualValues(t, expect, bz1)
		require.EqualValues(t, expect, bz2)
		require.EqualValues(t, expect, bz3)

		require.EqualValues(t, len(expect), resp.AminoSize())
	}
}

func BenchmarkResponseBeginBlockAmino(b *testing.B) {
	var testRespBeginBlock = []*ResponseBeginBlock{
		{},
		{
			Events: []Event{
				{
					Type: "longTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongTypelongType",
					Attributes: []kv.Pair{
						{Key: []byte("keykeykeykeykeykeykeykeykeykeykeykeykeykeykeykey"), Value: []byte("valuevaluevaluevaluevaluevaluevaluevaluevaluevaluevaluevaluevalue")},
					},
				},
			},
		},
		{
			Events: []Event{},
		},
		{
			Events: []Event{{}},
		},
	}

	b.ResetTimer()

	b.Run("amino", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.ReportAllocs()
			for _, resp := range testRespBeginBlock {
				_, err := cdc.MarshalBinaryBare(resp)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshaller", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.ReportAllocs()
			for _, resp := range testRespBeginBlock {
				_, err := MarshalResponseBeginBlockToAmino(resp)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.ReportAllocs()
			for _, resp := range testRespBeginBlock {
				_, err := resp.marshalToAmino()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshal with size", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.ReportAllocs()
			for _, resp := range testRespBeginBlock {
				_, err := resp.marshalToAminoWithSizeCompute()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshal with size 2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.ReportAllocs()
			for _, resp := range testRespBeginBlock {
				_, err := resp.marshalToAminoWithSizeCompute2()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

func TestResponseEndBlockAmino(t *testing.T) {
	var resps = []*ResponseEndBlock{
		{},
		{
			ValidatorUpdates: []ValidatorUpdate{
				{
					PubKey: PubKey{
						Type: "test",
					},
				},
			},
			ConsensusParamUpdates: &ConsensusParams{},
			Events:                []Event{},
		},
		{
			ValidatorUpdates:      []ValidatorUpdate{},
			ConsensusParamUpdates: &ConsensusParams{},
			Events:                []Event{{}},
		},
		{
			ValidatorUpdates:      []ValidatorUpdate{{}},
			ConsensusParamUpdates: &ConsensusParams{Block: &BlockParams{}, Evidence: &EvidenceParams{}, Validator: &ValidatorParams{}},
			Events:                []Event{{}, {Type: "Event"}, {}},
		},
	}
	for _, resp := range resps {
		expect, err := cdc.MarshalBinaryBare(resp)
		require.NoError(t, err)

		actual, err := MarshalResponseEndBlockToAmino(resp)
		require.NoError(t, err)
		require.EqualValues(t, expect, actual)
	}
}

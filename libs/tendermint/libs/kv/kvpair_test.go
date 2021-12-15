package kv

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

var testPairs = []Pair{
	{},
	{Key: []byte("key")},
	{Value: []byte("value")},
	{Key: []byte("key1"), Value: []byte("value1")},
	{Key: []byte("key1"), Value: []byte("value1"), XXX_NoUnkeyedLiteral: struct{}{}, XXX_sizecache: -10, XXX_unrecognized: []byte("unrecognized")},
	{Key: []byte{}, Value: []byte{}},
	{Key: []byte{}, Value: []byte{}, XXX_sizecache: 10},
}

func TestKvPairAmino(t *testing.T) {
	for _, pair := range testPairs {
		expect, err := cdc.MarshalBinaryBare(pair)
		require.NoError(t, err)

		actual, err := MarshalPairToAmino(pair)
		require.NoError(t, err)
		require.EqualValues(t, expect, actual)

		bz1, err := pair.marshalToAmino()
		require.NoError(t, err)

		bz2, err := pair.marshalToAminoWithPool()
		require.NoError(t, err)

		bz3, err := pair.marshalToAminoWithSizeCompute()
		require.NoError(t, err)

		require.EqualValues(t, expect, bz1)
		require.EqualValues(t, expect, bz2)
		require.EqualValues(t, expect, bz3)

		require.EqualValues(t, len(expect), pair.AminoSize())
	}
}

func BenchmarkKvPairAmino(b *testing.B) {
	b.Run("amino", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, pair := range testPairs {
				_, err := cdc.MarshalBinaryBare(pair)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	b.Run("marshaller", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, pair := range testPairs {
				_, err := MarshalPairToAmino(pair)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	b.Run("marshal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, pair := range testPairs {
				_, err := pair.marshalToAmino()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	b.Run("marshal with pool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, pair := range testPairs {
				_, err := pair.marshalToAminoWithPool()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	b.Run("marshal with size", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, pair := range testPairs {
				_, err := pair.marshalToAminoWithSizeCompute()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

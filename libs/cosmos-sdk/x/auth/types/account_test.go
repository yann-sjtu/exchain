package types

import (
	"errors"
	"math/big"
	"testing"

	"github.com/okex/exchain/libs/tendermint/crypto/sr25519"

	"github.com/okex/exchain/libs/tendermint/crypto/ed25519"

	"github.com/okex/exchain/libs/tendermint/crypto/secp256k1"
	"github.com/stretchr/testify/require"

	"github.com/okex/exchain/libs/cosmos-sdk/codec"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth/exported"
)

func TestBaseAddressPubKey(t *testing.T) {
	_, pub1, addr1 := KeyTestPubAddr()
	_, pub2, addr2 := KeyTestPubAddr()
	acc := NewBaseAccountWithAddress(addr1)

	// check the address (set) and pubkey (not set)
	require.EqualValues(t, addr1, acc.GetAddress())
	require.EqualValues(t, nil, acc.GetPubKey())

	// can't override address
	err := acc.SetAddress(addr2)
	require.NotNil(t, err)
	require.EqualValues(t, addr1, acc.GetAddress())

	// set the pubkey
	err = acc.SetPubKey(pub1)
	require.Nil(t, err)
	require.Equal(t, pub1, acc.GetPubKey())

	// can override pubkey
	err = acc.SetPubKey(pub2)
	require.Nil(t, err)
	require.Equal(t, pub2, acc.GetPubKey())

	//------------------------------------

	// can set address on empty account
	acc2 := BaseAccount{}
	err = acc2.SetAddress(addr2)
	require.Nil(t, err)
	require.EqualValues(t, addr2, acc2.GetAddress())
}

func TestBaseAccountCoins(t *testing.T) {
	_, _, addr := KeyTestPubAddr()
	acc := NewBaseAccountWithAddress(addr)

	someCoins := sdk.Coins{sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 246)}

	err := acc.SetCoins(someCoins)
	require.Nil(t, err)
	require.Equal(t, someCoins, acc.GetCoins())
}

func TestBaseAccountSequence(t *testing.T) {
	_, _, addr := KeyTestPubAddr()
	acc := NewBaseAccountWithAddress(addr)

	seq := uint64(7)

	err := acc.SetSequence(seq)
	require.Nil(t, err)
	require.Equal(t, seq, acc.GetSequence())
}

func TestBaseAccountMarshal(t *testing.T) {
	_, pub, addr := KeyTestPubAddr()
	acc := NewBaseAccountWithAddress(addr)

	someCoins := sdk.Coins{sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 246)}
	seq := uint64(7)

	// set everything on the account
	err := acc.SetPubKey(pub)
	require.Nil(t, err)
	err = acc.SetSequence(seq)
	require.Nil(t, err)
	err = acc.SetCoins(someCoins)
	require.Nil(t, err)

	// need a codec for marshaling
	cdc := codec.New()
	codec.RegisterCrypto(cdc)

	b, err := cdc.MarshalBinaryLengthPrefixed(acc)
	require.Nil(t, err)

	acc2 := BaseAccount{}
	err = cdc.UnmarshalBinaryLengthPrefixed(b, &acc2)
	require.Nil(t, err)
	require.Equal(t, acc, acc2)

	// error on bad bytes
	acc2 = BaseAccount{}
	err = cdc.UnmarshalBinaryLengthPrefixed(b[:len(b)/2], &acc2)
	require.NotNil(t, err)
}

func TestBaseAccountAmino(t *testing.T) {
	_, pub, addr := KeyTestPubAddr()
	acc := NewBaseAccountWithAddress(addr)

	someCoins := sdk.Coins{sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 123456789123456)}
	seq := uint64(7)

	// set everything on the account
	err := acc.SetPubKey(pub)
	require.Nil(t, err)
	err = acc.SetSequence(seq)
	require.Nil(t, err)
	err = acc.SetCoins(someCoins)
	require.Nil(t, err)

	accs := []BaseAccount{
		{},
		{
			Address:       addr,
			PubKey:        pub,
			Coins:         someCoins,
			Sequence:      seq,
			AccountNumber: 512,
		},
		{
			Address:       []byte{},
			PubKey:        ed25519.GenPrivKey().PubKey(),
			Coins:         sdk.Coins{},
			Sequence:      1024,
			AccountNumber: 512,
		},
		{
			Address: ed25519.GenPrivKey().Bytes(),
			PubKey:  sr25519.GenPrivKey().PubKey(),
			Coins: sdk.Coins{
				sdk.NewDecCoinFromDec("test", sdk.Dec{big.NewInt(123456789012345).Mul(big.NewInt(123456789012345), (big.NewInt(123456789012345)))}),
				sdk.DecCoin{"test2", sdk.Dec{nil}},
				sdk.DecCoin{"", sdk.Dec{nil}},
			},
			Sequence:      1024,
			AccountNumber: 512,
		},
		acc,
	}

	// need a codec for marshaling
	cdc := codec.New()
	codec.RegisterCrypto(cdc)

	for _, acc := range accs {
		b, err := cdc.MarshalBinaryBare(acc)
		require.Nil(t, err)

		b2, err := acc.marshalToAmino()
		require.NoError(t, err)

		b3, err := acc.marshalToAminoWithSizeCompute()
		require.NoError(t, err)

		b4, err := acc.marshalToAminoWithPool()
		require.NoError(t, err)

		require.EqualValues(t, b, b2)
		require.EqualValues(t, b, b3)
		require.EqualValues(t, b, b4)
	}
}

func BenchmarkBaseAccountAmino(b *testing.B) {
	accs := []BaseAccount{
		{},
		{
			Address:       []byte{},
			PubKey:        ed25519.GenPrivKey().PubKey(),
			Coins:         sdk.Coins{},
			Sequence:      1024,
			AccountNumber: 512,
		},
		{
			Address: ed25519.GenPrivKey().Bytes(),
			PubKey:  sr25519.GenPrivKey().PubKey(),
			Coins: sdk.Coins{
				sdk.NewDecCoinFromDec("test", sdk.Dec{big.NewInt(123456789012345).Mul(big.NewInt(123456789012345), (big.NewInt(123456789012345)))}),
				sdk.DecCoin{"test2", sdk.Dec{nil}},
			},
			Sequence:      1024,
			AccountNumber: 512,
		},
	}

	// need a codec for marshaling
	cdc := codec.New()
	codec.RegisterCrypto(cdc)

	b.ResetTimer()

	b.Run("amino", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, acc := range accs {
				_, err := cdc.MarshalBinaryBare(acc)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshaller with pool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, acc := range accs {
				_, err := acc.marshalToAminoWithPool()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshaller size compute", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.ReportAllocs()
			for _, acc := range accs {
				_, err := acc.marshalToAminoWithSizeCompute()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("marshaller", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.ReportAllocs()
			for _, acc := range accs {
				_, err := acc.marshalToAmino()
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

func TestGenesisAccountValidate(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	baseAcc := NewBaseAccount(addr, nil, pubkey, 0, 0)
	tests := []struct {
		name   string
		acc    exported.GenesisAccount
		expErr error
	}{
		{
			"valid base account",
			baseAcc,
			nil,
		},
		{
			"invalid base valid account",
			NewBaseAccount(addr, sdk.NewCoins(), secp256k1.GenPrivKey().PubKey(), 0, 0),
			errors.New("pubkey and address pair is invalid"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.acc.Validate()
			require.Equal(t, tt.expErr, err)
		})
	}
}

package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/tendermint/go-amino"

	"gopkg.in/yaml.v2"

	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	sdkerrors "github.com/okex/exchain/libs/cosmos-sdk/types/errors"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth/exported"
	authtypes "github.com/okex/exchain/libs/cosmos-sdk/x/auth/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

var _ exported.Account = (*EthAccount)(nil)
var _ exported.GenesisAccount = (*EthAccount)(nil)

var ethAccountBufferPool = &sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func init() {
	authtypes.RegisterAccountTypeCodec(&EthAccount{}, EthAccountName)
}

// ----------------------------------------------------------------------------
// Main OKExChain account
// ----------------------------------------------------------------------------

// EthAccount implements the auth.Account interface and embeds an
// auth.BaseAccount type. It is compatible with the auth.AccountKeeper.
type EthAccount struct {
	*authtypes.BaseAccount `json:"base_account" yaml:"base_account"`
	CodeHash               []byte `json:"code_hash" yaml:"code_hash"`
}

func (acc EthAccount) MarshalToAmino() ([]byte, error) {
	return acc.marshalToAminoWithSizeCompute()
}

func (acc EthAccount) marshalToAmino() ([]byte, error) {
	var buf = &bytes.Buffer{}
	err := acc.marshalToAminoToBuffer(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (acc EthAccount) marshalToAminoWithPool() ([]byte, error) {
	var buf = ethAccountBufferPool.Get().(*bytes.Buffer)
	defer ethAccountBufferPool.Put(buf)
	buf.Reset()
	err := acc.marshalToAminoToBuffer(buf)
	if err != nil {
		return nil, err
	}
	return amino.GetBytesBufferCopy(buf), nil
}

func (acc EthAccount) marshalToAminoToBuffer(buf *bytes.Buffer) error {
	var fieldKeysType = [2]byte{1<<3 | 2, 2<<3 | 2}
	var err error

	for pos := 1; pos <= 2; pos++ {
		switch pos {
		case 1:
			if acc.BaseAccount == nil {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return err
			}
			data, err := acc.BaseAccount.MarshalToAmino()
			if err != nil {
				return err
			}
			err = amino.EncodeUvarintToBuffer(buf, uint64(len(data)))
			if err != nil {
				return err
			}
			_, err = buf.Write(data)
			if err != nil {
				return err
			}
		case 2:
			codeHashLen := len(acc.CodeHash)
			if codeHashLen == 0 {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return err
			}
			err = amino.EncodeUvarintToBuffer(buf, uint64(codeHashLen))
			if err != nil {
				return err
			}
			_, err = buf.Write(acc.CodeHash)
			if err != nil {
				return err
			}
		default:
			panic("unreachable")
		}
	}
	return nil
}

func (acc EthAccount) marshalToAminoWithSizeCompute() ([]byte, error) {
	var buf = &bytes.Buffer{}
	var fieldKeysType = [2]byte{1<<3 | 2, 2<<3 | 2}
	var err error
	var size int
	var baseAccountBytes []byte

	if acc.BaseAccount != nil {
		baseAccountBytes, err = acc.BaseAccount.MarshalToAmino()
		if err != nil {
			return nil, err
		}
		size += 1 + amino.UvarintSize(uint64(len(baseAccountBytes))) + len(baseAccountBytes)
	}

	if len(acc.CodeHash) != 0 {
		size += 1 + amino.UvarintSize(uint64(len(acc.CodeHash))) + len(acc.CodeHash)
	}

	buf.Grow(size)

	if acc.BaseAccount != nil {
		err = buf.WriteByte(fieldKeysType[0])
		if err != nil {
			return nil, err
		}
		err = amino.EncodeByteSliceToBuffer(buf, baseAccountBytes)
		if err != nil {
			return nil, err
		}
	}

	if len(acc.CodeHash) != 0 {
		err = buf.WriteByte(fieldKeysType[1])
		if err != nil {
			return nil, err
		}
		err = amino.EncodeByteSliceToBuffer(buf, acc.CodeHash)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// ProtoAccount defines the prototype function for BaseAccount used for an
// AccountKeeper.
func ProtoAccount() exported.Account {
	return &EthAccount{
		BaseAccount: &auth.BaseAccount{},
		CodeHash:    ethcrypto.Keccak256(nil),
	}
}

// EthAddress returns the account address ethereum format.
func (acc EthAccount) EthAddress() ethcmn.Address {
	return ethcmn.BytesToAddress(acc.Address.Bytes())
}

// TODO: remove on SDK v0.40

// Balance returns the balance of an account.
func (acc EthAccount) Balance(denom string) sdk.Dec {
	return acc.GetCoins().AmountOf(denom)
}

// SetBalance sets an account's balance of the given coin denomination.
//
// CONTRACT: assumes the denomination is valid.
func (acc *EthAccount) SetBalance(denom string, amt sdk.Dec) {
	coins := acc.GetCoins()
	diff := amt.Sub(coins.AmountOf(denom))
	switch {
	case diff.IsPositive():
		// Increase coins to amount
		coins = coins.Add(sdk.NewCoin(denom, diff))
	case diff.IsNegative():
		// Decrease coins to amount
		coins = coins.Sub(sdk.NewCoins(sdk.NewCoin(denom, diff.Neg())))
	default:
		return
	}

	if err := acc.SetCoins(coins); err != nil {
		panic(fmt.Errorf("could not set %s coins for address %s: %w", denom, acc.EthAddress().String(), err))
	}
}

type ethermintAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	EthAddress    string         `json:"eth_address" yaml:"eth_address"`
	Coins         sdk.Coins      `json:"coins" yaml:"coins"`
	PubKey        string         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
	CodeHash      string         `json:"code_hash" yaml:"code_hash"`
}

// MarshalYAML returns the YAML representation of an account.
func (acc EthAccount) MarshalYAML() (interface{}, error) {
	alias := ethermintAccountPretty{
		Address:       acc.Address,
		EthAddress:    acc.EthAddress().String(),
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
		CodeHash:      ethcmn.Bytes2Hex(acc.CodeHash),
	}

	var err error

	if acc.PubKey != nil {
		alias.PubKey, err = sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, acc.PubKey)
		if err != nil {
			return nil, err
		}
	}

	bz, err := yaml.Marshal(alias)
	if err != nil {
		return nil, err
	}

	return string(bz), err
}

// MarshalJSON returns the JSON representation of an EthAccount.
func (acc EthAccount) MarshalJSON() ([]byte, error) {
	var ethAddress = ""

	if acc.BaseAccount != nil && acc.Address != nil {
		ethAddress = acc.EthAddress().String()
	}

	alias := ethermintAccountPretty{
		Address:       acc.Address,
		EthAddress:    ethAddress,
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
		CodeHash:      ethcmn.Bytes2Hex(acc.CodeHash),
	}

	var err error

	if acc.PubKey != nil {
		alias.PubKey, err = sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, acc.PubKey)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into an EthAccount.
func (acc *EthAccount) UnmarshalJSON(bz []byte) error {
	var (
		alias ethermintAccountPretty
		err   error
	)

	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	switch {
	case !alias.Address.Empty() && alias.EthAddress != "":
		// Both addresses provided. Verify correctness
		ethAddress := ethcmn.HexToAddress(alias.EthAddress)
		ethAddressFromAccAddress := ethcmn.BytesToAddress(alias.Address.Bytes())

		if !bytes.Equal(ethAddress.Bytes(), alias.Address.Bytes()) {
			err = sdkerrors.Wrapf(
				sdkerrors.ErrInvalidAddress,
				"expected %s, got %s",
				ethAddressFromAccAddress.String(), ethAddress.String(),
			)
		}

	case !alias.Address.Empty() && alias.EthAddress == "":
		// unmarshal sdk.AccAddress only. Do nothing here
	case alias.Address.Empty() && alias.EthAddress != "":
		// retrieve sdk.AccAddress from ethereum address
		ethAddress := ethcmn.HexToAddress(alias.EthAddress)
		alias.Address = sdk.AccAddress(ethAddress.Bytes())
	case alias.Address.Empty() && alias.EthAddress == "":
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"account must contain address in Ethereum Hex or Cosmos Bech32 format",
		)
	}

	if err != nil {
		return err
	}

	acc.BaseAccount = &authtypes.BaseAccount{
		Coins:         alias.Coins,
		Address:       alias.Address,
		AccountNumber: alias.AccountNumber,
		Sequence:      alias.Sequence,
	}
	acc.CodeHash = ethcmn.Hex2Bytes(alias.CodeHash)

	if alias.PubKey != "" {
		acc.BaseAccount.PubKey, err = sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, alias.PubKey)
		if err != nil {
			return err
		}
	}
	return nil
}

// String implements the fmt.Stringer interface
func (acc EthAccount) String() string {
	out, _ := yaml.Marshal(acc)
	return string(out)
}

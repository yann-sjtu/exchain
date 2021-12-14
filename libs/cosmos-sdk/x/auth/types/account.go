package types

import (
	"bytes"
	"errors"
	"time"

	"github.com/tendermint/go-amino"

	"github.com/okex/exchain/libs/tendermint/crypto"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth/exported"
	cryptoamino "github.com/okex/exchain/libs/tendermint/crypto/encoding/amino"
)

//-----------------------------------------------------------------------------
// BaseAccount

var _ exported.Account = (*BaseAccount)(nil)
var _ exported.GenesisAccount = (*BaseAccount)(nil)

// BaseAccount - a base account structure.
// This can be extended by embedding within in your AppAccount.
// However one doesn't have to use BaseAccount as long as your struct
// implements Account.
type BaseAccount struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	Coins         sdk.Coins      `json:"coins" yaml:"coins"`
	PubKey        crypto.PubKey  `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
}

var baseAccountBufferPool = amino.NewBufferPool()

func (acc BaseAccount) MarshalToAmino() ([]byte, error) {
	return acc.marshalToAminoWithSizeCompute()
}

func (acc BaseAccount) marshalToAminoWithPool() ([]byte, error) {
	var buf = baseAccountBufferPool.Get()
	defer baseAccountBufferPool.Put(buf)
	err := acc.marshalToAminoBuffer(buf)
	if err != nil {
		return nil, err
	}
	return amino.GetBytesBufferCopy(buf), nil
}

func (acc BaseAccount) marshalToAmino() ([]byte, error) {
	var buf = &bytes.Buffer{}
	err := acc.marshalToAminoBuffer(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var fieldKeysType = [5]byte{1<<3 | 2, 2<<3 | 2, 3<<3 | 2, 4 << 3, 5 << 3}

func (acc BaseAccount) marshalToAminoBuffer(buf *bytes.Buffer) error {
	for pos := 1; pos <= 5; pos++ {
		var err error
		switch pos {
		case 1:
			addressLen := len(acc.Address)
			if addressLen == 0 {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return err
			}
			err = amino.EncodeUvarintToBuffer(buf, uint64(addressLen))
			if err != nil {
				return err
			}
			_, err = buf.Write(acc.Address)
			if err != nil {
				return err
			}
		case 2:
			for _, coin := range acc.Coins {
				err = buf.WriteByte(fieldKeysType[pos-1])
				if err != nil {
					return err
				}
				data, err := coin.MarshalToAmino()
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
			}
		case 3:
			if acc.PubKey == nil {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return err
			}
			data, err := cryptoamino.MarshalPubKeyToAminoWithTypePrefix(acc.PubKey)
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
		case 4:
			if acc.AccountNumber == 0 {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return err
			}
			err := amino.EncodeUvarintToBuffer(buf, acc.AccountNumber)
			if err != nil {
				return err
			}
		case 5:
			if acc.Sequence == 0 {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return err
			}
			err := amino.EncodeUvarintToBuffer(buf, acc.Sequence)
			if err != nil {
				return err
			}
		default:
			panic("unreachable")
		}
	}
	return nil
}

func (acc BaseAccount) marshalToAminoWithSizeCompute() ([]byte, error) {
	var size int
	var err error
	var coinsBytes [][]byte
	var pubkeyaminoSize int

	addressLen := len(acc.Address)
	if addressLen != 0 {
		size += 1 + amino.UvarintSize(uint64(addressLen)) + addressLen
	}
	if len(acc.Coins) <= 6 {
		var coinsBytesArray [6][]byte
		coinsBytes = coinsBytesArray[:0]
	} else {
		coinsBytes = make([][]byte, 0, len(acc.Coins))
	}
	for _, coin := range acc.Coins {
		var data []byte
		data, err = coin.MarshalToAmino()
		if err != nil {
			return nil, err
		}
		coinsBytes = append(coinsBytes, data)
		size += 1 + amino.UvarintSize(uint64(len(data))) + len(data)
	}
	if acc.PubKey != nil {
		pubkeyaminoSize, err = cryptoamino.PubKeyAminoSizeWithTypePrefix(acc.PubKey)
		if err != nil {
			return nil, err
		}
		size += 1 + amino.UvarintSize(uint64(pubkeyaminoSize)) + pubkeyaminoSize
	}
	if acc.AccountNumber != 0 {
		size += 1 + amino.UvarintSize(acc.AccountNumber)
	}
	if acc.Sequence != 0 {
		size += 1 + amino.UvarintSize(acc.Sequence)
	}

	var buf = &bytes.Buffer{}
	buf.Grow(size)

	for pos := 1; pos <= 5; pos++ {
		switch pos {
		case 1:
			if addressLen == 0 {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return nil, err
			}
			err = amino.EncodeByteSliceToBuffer(buf, acc.Address)
			if err != nil {
				return nil, err
			}
		case 2:
			for _, coinBytes := range coinsBytes {
				err = buf.WriteByte(fieldKeysType[pos-1])
				if err != nil {
					return nil, err
				}
				err = amino.EncodeByteSliceToBuffer(buf, coinBytes)
				if err != nil {
					return nil, err
				}
			}
		case 3:
			if acc.PubKey == nil {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return nil, err
			}
			err = amino.EncodeUvarintToBuffer(buf, uint64(pubkeyaminoSize))
			if err != nil {
				return nil, err
			}
			lenBeforeValue := buf.Len()
			err = cryptoamino.MarshalPubKeyToAminoWithTypePrefixToBuffer(acc.PubKey, buf)
			if err != nil {
				return nil, err
			}
			lenAfterValue := buf.Len()
			if lenAfterValue-lenBeforeValue != pubkeyaminoSize {
				return nil, errors.New("pub key marshal error")
			}
		case 4:
			if acc.AccountNumber == 0 {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return nil, err
			}
			err := amino.EncodeUvarintToBuffer(buf, acc.AccountNumber)
			if err != nil {
				return nil, err
			}
		case 5:
			if acc.Sequence == 0 {
				break
			}
			err = buf.WriteByte(fieldKeysType[pos-1])
			if err != nil {
				return nil, err
			}
			err := amino.EncodeUvarintToBuffer(buf, acc.Sequence)
			if err != nil {
				return nil, err
			}
		default:
			panic("unreachable")
		}
	}
	return buf.Bytes(), nil
}

// NewBaseAccount creates a new BaseAccount object
func NewBaseAccount(address sdk.AccAddress, coins sdk.Coins,
	pubKey crypto.PubKey, accountNumber uint64, sequence uint64) *BaseAccount {

	return &BaseAccount{
		Address:       address,
		Coins:         coins,
		PubKey:        pubKey,
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}
}

// ProtoBaseAccount - a prototype function for BaseAccount
func ProtoBaseAccount() exported.Account {
	return &BaseAccount{}
}

// NewBaseAccountWithAddress - returns a new base account with a given address
func NewBaseAccountWithAddress(addr sdk.AccAddress) BaseAccount {
	return BaseAccount{
		Address: addr,
	}
}

// GetAddress - Implements sdk.Account.
func (acc BaseAccount) GetAddress() sdk.AccAddress {
	return acc.Address
}

// SetAddress - Implements sdk.Account.
func (acc *BaseAccount) SetAddress(addr sdk.AccAddress) error {
	if len(acc.Address) != 0 {
		return errors.New("cannot override BaseAccount address")
	}
	acc.Address = addr
	return nil
}

// GetPubKey - Implements sdk.Account.
func (acc BaseAccount) GetPubKey() crypto.PubKey {
	return acc.PubKey
}

// SetPubKey - Implements sdk.Account.
func (acc *BaseAccount) SetPubKey(pubKey crypto.PubKey) error {
	acc.PubKey = pubKey
	return nil
}

// GetCoins - Implements sdk.Account.
func (acc *BaseAccount) GetCoins() sdk.Coins {
	return acc.Coins
}

// SetCoins - Implements sdk.Account.
func (acc *BaseAccount) SetCoins(coins sdk.Coins) error {
	acc.Coins = coins
	return nil
}

// GetAccountNumber - Implements Account
func (acc *BaseAccount) GetAccountNumber() uint64 {
	return acc.AccountNumber
}

// SetAccountNumber - Implements Account
func (acc *BaseAccount) SetAccountNumber(accNumber uint64) error {
	acc.AccountNumber = accNumber
	return nil
}

// GetSequence - Implements sdk.Account.
func (acc *BaseAccount) GetSequence() uint64 {
	return acc.Sequence
}

// SetSequence - Implements sdk.Account.
func (acc *BaseAccount) SetSequence(seq uint64) error {
	acc.Sequence = seq
	return nil
}

// SpendableCoins returns the total set of spendable coins. For a base account,
// this is simply the base coins.
func (acc *BaseAccount) SpendableCoins(_ time.Time) sdk.Coins {
	return acc.GetCoins()
}

// Validate checks for errors on the account fields
func (acc BaseAccount) Validate() error {
	if acc.PubKey != nil && acc.Address != nil &&
		!bytes.Equal(acc.PubKey.Address().Bytes(), acc.Address.Bytes()) {
		return errors.New("pubkey and address pair is invalid")
	}

	return nil
}

type baseAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	Coins         sdk.Coins      `json:"coins" yaml:"coins"`
	PubKey        string         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
}

func (acc BaseAccount) String() string {
	out, _ := acc.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of an account.
func (acc BaseAccount) MarshalYAML() (interface{}, error) {
	alias := baseAccountPretty{
		Address:       acc.Address,
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
	}

	if acc.PubKey != nil {
		pks, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, acc.PubKey)
		if err != nil {
			return nil, err
		}

		alias.PubKey = pks
	}

	bz, err := yaml.Marshal(alias)
	if err != nil {
		return nil, err
	}

	return string(bz), err
}

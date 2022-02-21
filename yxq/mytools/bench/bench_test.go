package bench

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/okex/exchain/libs/tendermint/crypto/etherhash"
	"github.com/tendermint/go-amino"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/okex/exchain/libs/tendermint/crypto/ed25519"
	"github.com/okex/exchain/libs/tendermint/mempool"
	evmtypes "github.com/okex/exchain/x/evm/types"
)

func TestSignature(t *testing.T) {
	rawTx, err := hex.DecodeString("f889568405f5e100832dc6c094456f9c7c7c69ac265626dc0cacca4a6747f0ad4d80a41003e2d2000000000000000000000000000000000000000000000000000000000000006481a9a02408234ee9dd8f4f2bee2771c27071dc3a6d997bee1a76acc76cd132c6e58d27a07d9ed16f18c35886764a698e666f2cdcfa77b575a8347001bf4c98994d937013")
	if err != nil {
		t.Fatal(err)
	}

	var tx evmtypes.MsgEthereumTx
	if err := rlp.DecodeBytes(rawTx, &tx); err != nil {
		t.Fatal(err)
	}

	if _, err := tx.VerifySig(big.NewInt(67), 0, nil); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkRecoverEthSig(b *testing.B) {
	rawTx, err := hex.DecodeString("f889568405f5e100832dc6c094456f9c7c7c69ac265626dc0cacca4a6747f0ad4d80a41003e2d2000000000000000000000000000000000000000000000000000000000000006481a9a02408234ee9dd8f4f2bee2771c27071dc3a6d997bee1a76acc76cd132c6e58d27a07d9ed16f18c35886764a698e666f2cdcfa77b575a8347001bf4c98994d937013")
	if err != nil {
		b.Fatal(err)
	}

	var tx evmtypes.MsgEthereumTx
	if err := rlp.DecodeBytes(rawTx, &tx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i:=0; i<b.N;i++ {
		if _, err := tx.VerifySig(big.NewInt(67), 0, nil); err != nil {
			b.Fatal(err)
		}
		tmp := evmtypes.MsgEthereumTx{
			Data: tx.Data,
		}
		tx = tmp
	}
}

func BenchmarkSetFrom(b *testing.B) {
	rawTx, err := hex.DecodeString("f889568405f5e100832dc6c094456f9c7c7c69ac265626dc0cacca4a6747f0ad4d80a41003e2d2000000000000000000000000000000000000000000000000000000000000006481a9a02408234ee9dd8f4f2bee2771c27071dc3a6d997bee1a76acc76cd132c6e58d27a07d9ed16f18c35886764a698e666f2cdcfa77b575a8347001bf4c98994d937013")
	if err != nil {
		b.Fatal(err)
	}

	var tx evmtypes.MsgEthereumTx
	if err := rlp.DecodeBytes(rawTx, &tx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i:=0; i<b.N;i++ {
		tx.SetFrom("BBE4733D85BC2B90682147779DA49CAB38C0AA1F")
	}
}

func BenchmarkVerify(b *testing.B) {
	privKey := ed25519.GenPrivKey()
	rawTx, err := hex.DecodeString("f889568405f5e100832dc6c094456f9c7c7c69ac265626dc0cacca4a6747f0ad4d80a41003e2d2000000000000000000000000000000000000000000000000000000000000006481a9a02408234ee9dd8f4f2bee2771c27071dc3a6d997bee1a76acc76cd132c6e58d27a07d9ed16f18c35886764a698e666f2cdcfa77b575a8347001bf4c98994d937013")
	if err != nil {
		b.Fatal(err)
	}
	wtx := &mempool.WrappedTx{
		Payload: rawTx,
		From:    "test-address-for-unit-test",
		NodeKey: privKey.PubKey().Bytes(),
	}

	sig, err := privKey.Sign(append(wtx.Payload, wtx.From...))
	if err != nil {
		b.Fatal(err)
	}
	wtx.Signature = sig

	pub := privKey.PubKey()
	b.ResetTimer()
	for i:=0;i<b.N;i++ {
		if !pub.VerifyBytes(append(wtx.Payload, wtx.From...), sig) {
			b.Fatal("invalid signature")
		}
	}
}

func BenchmarkSignWtx(b *testing.B) {
	hexPriv := "d322864e848a3ebbb88cbd45b163db3c479b166937f10a14ab86a3f860b0b0b64506fc928bd335f434691375f63d0baf97968716a20b2ad15463e51ba5cf49fe"
	var privKey ed25519.PrivKeyEd25519
	bs, _ := hex.DecodeString(hexPriv)
	copy(privKey[:], bs)

	s := "f889028405f5e100832dc6c094000000000000000000000000000000000000000080a41003e2d2000000000000000000000000000000000000000000000000000000000000006481aaa019ebb1d813b837bed5648644117e268cb328d9921da0ccc28b2d1c35a26aa4eba0483445cb2518e53708abf5180c257ff1fc46d8b893146bd575c3a78b18fadfdd"
	tx, _ := hex.DecodeString(s)
	wtx := mempool.WrappedTx{
		Payload: tx,
		From: "0xbbE4733d85bc2b90682147779DA49caB38C0aA1F",
		NodeKey: privKey.PubKey().Bytes(),
	}
	b.ResetTimer()
	for i:=0;i<b.N;i++ {
		if _, err := privKey.Sign(append(wtx.Payload, wtx.From...)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSignWtxHash(b *testing.B) {
	hexPriv := "d322864e848a3ebbb88cbd45b163db3c479b166937f10a14ab86a3f860b0b0b64506fc928bd335f434691375f63d0baf97968716a20b2ad15463e51ba5cf49fe"
	var privKey ed25519.PrivKeyEd25519
	bs, _ := hex.DecodeString(hexPriv)
	copy(privKey[:], bs)

	s := "f889028405f5e100832dc6c094000000000000000000000000000000000000000080a41003e2d2000000000000000000000000000000000000000000000000000000000000006481aaa019ebb1d813b837bed5648644117e268cb328d9921da0ccc28b2d1c35a26aa4eba0483445cb2518e53708abf5180c257ff1fc46d8b893146bd575c3a78b18fadfdd"
	tx, _ := hex.DecodeString(s)
	wtx := mempool.WrappedTx{
		Payload: tx,
		From: "0xbbE4733d85bc2b90682147779DA49caB38C0aA1F",
		NodeKey: privKey.PubKey().Bytes(),
	}
	b.ResetTimer()
	for i:=0;i<b.N;i++ {
		if _, err := privKey.Sign(etherhash.Sum(append(wtx.Payload, wtx.From...))); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeTxMessage(b *testing.B) {
	cdc := amino.NewCodec()
	mempool.RegisterMessages(cdc)
	s := "2b06579d0a8b01f889028405f5e100832dc6c094000000000000000000000000000000000000000080a41003e2d2000000000000000000000000000000000000000000000000000000000000006481aaa019ebb1d813b837bed5648644117e268cb328d9921da0ccc28b2d1c35a26aa4eba0483445cb2518e53708abf5180c257ff1fc46d8b893146bd575c3a78b18fadfdd"
	data, _ := hex.DecodeString(s)
	var msg mempool.TxMessage
	for i:=0;i<b.N;i++ {
		if err := cdc.UnmarshalBinaryBare(data, &msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeWtxMessage(b *testing.B) {
	cdc := amino.NewCodec()
	mempool.RegisterMessages(cdc)
	s := "a355fee90aa3020a8b01f889028405f5e100832dc6c094000000000000000000000000000000000000000080a41003e2d2000000000000000000000000000000000000000000000000000000000000006481aaa019ebb1d813b837bed5648644117e268cb328d9921da0ccc28b2d1c35a26aa4eba0483445cb2518e53708abf5180c257ff1fc46d8b893146bd575c3a78b18fadfdd122a3078626245343733336438356263326239303638323134373737394441343963614233384330614131461a4077f565f90b75100a6323dd6d9099c07c5a1f782dc91e9eb4d56609d3e7fc5fca06965a32c78330670ad7b0a836b767d9e86fba801166c49fd1f923cc754ce10c22251624de64204506fc928bd335f434691375f63d0baf97968716a20b2ad15463e51ba5cf49fe"
	data, _ := hex.DecodeString(s)
	var msg mempool.WtxMessage
	for i:=0;i<b.N;i++ {
		if err := cdc.UnmarshalBinaryBare(data, &msg); err != nil {
			b.Fatal(err)
		}
	}
}

func TestA(t *testing.T) {
	name := "/Users/xzavier/go/src/github.com/okex/exchain/yxq/tools/wtx/WtxMessage-0xbbE4733d85bc2b90682147779DA49caB38C0aA1F.txt"
	f, err := os.OpenFile(name, os.O_RDONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}
	r := bufio.NewReader(f)
	var line int
	for {
		data, _, _ := r.ReadLine()
		if len(data) == 0 {
			return
		}
		fmt.Println(line, string(data))
		line++
		if line > 4 {
			break
		}
	}
}

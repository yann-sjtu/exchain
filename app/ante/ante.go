package ante

import (
	"github.com/okex/exchain/app/crypto/ethsecp256k1"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	sdkerrors "github.com/okex/exchain/libs/cosmos-sdk/types/errors"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth"
	authante "github.com/okex/exchain/libs/cosmos-sdk/x/auth/ante"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth/types"
	channelkeeper "github.com/okex/exchain/libs/ibc-go/modules/core/04-channel/keeper"
	ibcante "github.com/okex/exchain/libs/ibc-go/modules/core/ante"
	"github.com/okex/exchain/libs/system/trace"
	tmcrypto "github.com/okex/exchain/libs/tendermint/crypto"
	wasmkeeper "github.com/okex/exchain/x/wasm/keeper"
)

func init() {
	ethsecp256k1.RegisterCodec(types.ModuleCdc)
}

const (
	// TODO: Use this cost per byte through parameter or overriding NewConsumeGasForTxSizeDecorator
	// which currently defaults at 10, if intended
	// memoCostPerByte     sdk.Gas = 3
	secp256k1VerifyCost uint64 = 21000
)

// NewAnteHandler returns an ante handler responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
func NewAnteHandler(ak auth.AccountKeeper, evmKeeper EVMKeeper, sk types.SupplyKeeper, validateMsgHandler ValidateMsgHandler, option wasmkeeper.HandlerOption, ibcChannelKeepr channelkeeper.Keeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler
		switch tx.GetType() {
		case sdk.StdTxType:
			anteHandler = sdk.ChainAnteDecorators(
				authante.NewSetUpContextDecorator(),                                             // outermost AnteDecorator. SetUpContext must be called first
				wasmkeeper.NewLimitSimulationGasDecorator(option.WasmConfig.SimulationGasLimit), // after setup context to enforce limits early
				wasmkeeper.NewCountTXDecorator(option.TXCounterStoreKey),
				NewAccountBlockedVerificationDecorator(evmKeeper), //account blocked check AnteDecorator
				authante.NewMempoolFeeDecorator(),
				authante.NewValidateBasicDecorator(),
				authante.NewValidateMemoDecorator(ak),
				authante.NewConsumeGasForTxSizeDecorator(ak),
				authante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
				authante.NewValidateSigCountDecorator(ak),
				authante.NewDeductFeeDecorator(ak, sk),
				authante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
				authante.NewSigVerificationDecorator(ak),
				authante.NewIncrementSequenceDecorator(ak), // innermost AnteDecorator
				NewValidateMsgHandlerDecorator(validateMsgHandler),
				ibcante.NewAnteDecorator(ibcChannelKeepr),
			)

		case sdk.EvmTxType:
			//if ctx.IsCheckTx() {
			//	var from string
			//	switch tx.GetGasPrice().Uint64() {
			//	case 100000000:
			//		from = "0xbbE4733d85bc2b90682147779DA49caB38C0aA1F"
			//	case 100000001:
			//		from = "0x83D83497431C2D3FEab296a9fba4e5FaDD2f7eD0"
			//	case 100000002:
			//		from = "0x4C12e733e58819A1d3520f1E7aDCc614Ca20De64"
			//	case 100000003:
			//		from = "0x2Bd4AF0C1D0c2930fEE852D07bB9dE87D8C07044"
			//	case 100000004:
			//		from = "0xEb4D58BA0ADBD52AE9dB5a97E5c99289e721b30C"
			//	case 100000005:
			//		from = "0xAdD3292c9479f0D3faffB50490adF3BF90361A74"
			//	case 100000006:
			//		from = "0x37f3BbE03880B3062B24F4a4b6c17F1188a1c8E5"
			//	case 100000007:
			//		from = "0xFE157eB5bf53bdF0CA11C4b8D7467a022b195535"
			//	}
			//	tx.(*types2.MsgEthereumTx).SetFrom(from)
			//	//return ctx, nil
			//}
			if ctx.IsWrappedCheckTx() {
				anteHandler = sdk.ChainAnteDecorators(
					NewNonceVerificationDecorator(ak),
					NewIncrementSenderSequenceDecorator(ak),
				)
			} else {
				anteHandler = sdk.ChainAnteDecorators(
					NewEthSetupContextDecorator(), // outermost AnteDecorator. EthSetUpContext must be called first
					NewGasLimitDecorator(evmKeeper),
					NewEthMempoolFeeDecorator(evmKeeper),
					authante.NewValidateBasicDecorator(),
					NewEthSigVerificationDecorator(),
					NewAccountBlockedVerificationDecorator(evmKeeper), //account blocked check AnteDecorator
					NewAccountAnteDecorator(ak, evmKeeper, sk),
				)
			}

		default:
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", tx)
		}

		return anteHandler(ctx, tx, sim)
	}
}

// sigGasConsumer overrides the DefaultSigVerificationGasConsumer from the x/auth
// module on the SDK. It doesn't allow ed25519 nor multisig thresholds.
func sigGasConsumer(
	meter sdk.GasMeter, _ []byte, pubkey tmcrypto.PubKey, _ types.Params,
) error {
	switch pubkey.(type) {
	case ethsecp256k1.PubKey:
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: secp256k1")
		return nil
	case tmcrypto.PubKey:
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: tendermint secp256k1")
		return nil
	default:
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
	}
}

func pinAnte(trc *trace.Tracer, tag string) {
	if trc != nil {
		trc.RepeatingPin(tag)
	}
}

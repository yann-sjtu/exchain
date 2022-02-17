package ante

import (
	"github.com/okex/exchain/app/crypto/ethsecp256k1"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	sdkerrors "github.com/okex/exchain/libs/cosmos-sdk/types/errors"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth"
	authante "github.com/okex/exchain/libs/cosmos-sdk/x/auth/ante"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth/keeper"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth/types"
	tmcrypto "github.com/okex/exchain/libs/tendermint/crypto"
	evmtypes "github.com/okex/exchain/x/evm/types"
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
func NewAnteHandler(ak auth.AccountKeeper, evmKeeper EVMKeeper, sk types.SupplyKeeper, validateMsgHandler ValidateMsgHandler) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler
		switch msgEthTx := tx.(type) {
		case auth.StdTx:
			anteHandler = sdk.ChainAnteDecorators(
				authante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
				NewAccountSetupDecorator(ak),
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
			)

		case evmtypes.MsgEthereumTx:
			if ctx.IsWrappedCheckTx() {
				if ctx.From() == "" {
					var addr string
					switch msgEthTx.Data.GasLimit {
					case 3000000:
						addr = "0xbbE4733d85bc2b90682147779DA49caB38C0aA1F"
					case 3000001:
						addr = "0x83D83497431C2D3FEab296a9fba4e5FaDD2f7eD0"
					case 3000002:
						addr = "0x4C12e733e58819A1d3520f1E7aDCc614Ca20De64"
					case 3000003:
						addr = "0x2Bd4AF0C1D0c2930fEE852D07bB9dE87D8C07044"
					default:
						panic(msgEthTx.Data.GasLimit)
					}
					ctx = ctx.WithFrom(addr)
				}

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
					NewAccountVerificationDecorator(ak, evmKeeper),
					NewNonceVerificationDecorator(ak),
					NewEthGasConsumeDecorator(ak, sk, evmKeeper),
					NewIncrementSenderSequenceDecorator(ak), // innermost AnteDecorator.
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

// AccountSetupDecorator sets an account to state if it's not stored already. This only applies for MsgEthermint.
type AccountSetupDecorator struct {
	ak auth.AccountKeeper
}

// NewAccountSetupDecorator creates a new AccountSetupDecorator instance
func NewAccountSetupDecorator(ak auth.AccountKeeper) AccountSetupDecorator {
	return AccountSetupDecorator{
		ak: ak,
	}
}

// AnteHandle sets an account for MsgEthermint (evm) if the sender is registered.
// NOTE: Since the account is set without any funds, the message execution will
// fail if the validator requires a minimum fee > 0.
func (asd AccountSetupDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "no messages included in transaction")
	}

	for _, msg := range msgs {
		if msgEthermint, ok := msg.(evmtypes.MsgEthermint); ok {
			setupAccount(asd.ak, ctx, msgEthermint.From)
		}
	}

	return next(ctx, tx, simulate)
}

func setupAccount(ak keeper.AccountKeeper, ctx sdk.Context, addr sdk.AccAddress) {
	acc := ak.GetAccount(ctx, addr)
	if acc != nil {
		return
	}

	acc = ak.NewAccountWithAddress(ctx, addr)
	ak.SetAccount(ctx, acc)
}

// AccountBlockedVerificationDecorator check whether signer is blocked.
type AccountBlockedVerificationDecorator struct {
	evmKeeper EVMKeeper
}

// NewAccountBlockedVerificationDecorator creates a new AccountBlockedVerificationDecorator instance
func NewAccountBlockedVerificationDecorator(evmKeeper EVMKeeper) AccountBlockedVerificationDecorator {
	return AccountBlockedVerificationDecorator{
		evmKeeper: evmKeeper,
	}
}

// AnteHandle check wether signer of tx(contains cosmos-tx and eth-tx) is blocked.
func (abvd AccountBlockedVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	signers, err := getSigners(tx)
	if err != nil {
		return ctx, err
	}
	currentGasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	for _, signer := range signers {
		//TODO it may be optimizate by cache blockedAddressList
		if ok := abvd.evmKeeper.IsAddressBlocked(ctx, signer); ok {
			ctx = ctx.WithGasMeter(currentGasMeter)
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "address: %s has been blocked", signer.String())
		}
	}
	ctx = ctx.WithGasMeter(currentGasMeter)
	return next(ctx, tx, simulate)
}

// getSigners get signers of tx(contains cosmos-tx and eth-tx.
func getSigners(tx sdk.Tx) ([]sdk.AccAddress, error) {
	signers := make([]sdk.AccAddress, 0)
	switch tx.(type) {
	case auth.StdTx:
		sigTx, ok := tx.(authante.SigVerifiableTx)
		if !ok {
			return signers, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
		}
		signers = append(signers, sigTx.GetSigners()...)
	case evmtypes.MsgEthereumTx:
		msgEthTx, ok := tx.(evmtypes.MsgEthereumTx)
		if !ok {
			return signers, sdkerrors.Wrapf(sdkerrors.ErrTxDecode, "invalid transaction type: %T", tx)
		}
		signers = append(signers, msgEthTx.GetSigners()...)

	default:
		return signers, sdkerrors.Wrapf(sdkerrors.ErrTxDecode, "invalid transaction type: %T", tx)
	}
	return signers, nil
}

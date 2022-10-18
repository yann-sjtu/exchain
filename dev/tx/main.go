package main

import (
	"bufio"
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/okex/exchain/libs/tendermint/crypto/ed25519"
	"github.com/okex/exchain/libs/tendermint/mempool"
	"github.com/tendermint/go-amino"
)

const (
	abiFile = "../client/contracts/counter/counter.abi"
	binFile = "../client/contracts/counter/counter.bin"

	ChainId  int64  = 67        //  okc
	GasPrice int64  = 100000000 // 0.1 gwei
	GasLimit uint64 = 3000000
)

const (
	oip20Code       = "608060405260405162001e6938038062001e6983398181016040528101906200002991906200064e565b856001908051906020019062000041929190620002de565b5084600090805190602001906200005a929190620002de565b5083600260006101000a81548160ff021916908360ff160217905550826003819055506200008f8284620000e360201b60201c565b8073ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f19350505050158015620000d6573d6000803e3d6000fd5b505050505050506200093c565b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16141562000156576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016200014d9062000789565b60405180910390fd5b62000172816003546200028060201b620008651790919060201c565b600381905550620001d181600460008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020546200028060201b620008651790919060201c565b600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508173ffffffffffffffffffffffffffffffffffffffff16600073ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051620002749190620007bc565b60405180910390a35050565b600082828462000291919062000808565b9150811015620002d8576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401620002cf90620008b5565b60405180910390fd5b92915050565b828054620002ec9062000906565b90600052602060002090601f0160209004810192826200031057600085556200035c565b82601f106200032b57805160ff19168380011785556200035c565b828001600101855582156200035c579182015b828111156200035b5782518255916020019190600101906200033e565b5b5090506200036b91906200036f565b5090565b5b808211156200038a57600081600090555060010162000370565b5090565b6000604051905090565b600080fd5b600080fd5b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b620003f782620003ac565b810181811067ffffffffffffffff82111715620004195762000418620003bd565b5b80604052505050565b60006200042e6200038e565b90506200043c8282620003ec565b919050565b600067ffffffffffffffff8211156200045f576200045e620003bd565b5b6200046a82620003ac565b9050602081019050919050565b60005b83811015620004975780820151818401526020810190506200047a565b83811115620004a7576000848401525b50505050565b6000620004c4620004be8462000441565b62000422565b905082815260208101848484011115620004e357620004e2620003a7565b5b620004f084828562000477565b509392505050565b600082601f83011262000510576200050f620003a2565b5b815162000522848260208601620004ad565b91505092915050565b600060ff82169050919050565b62000543816200052b565b81146200054f57600080fd5b50565b600081519050620005638162000538565b92915050565b6000819050919050565b6200057e8162000569565b81146200058a57600080fd5b50565b6000815190506200059e8162000573565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000620005d182620005a4565b9050919050565b620005e381620005c4565b8114620005ef57600080fd5b50565b6000815190506200060381620005d8565b92915050565b60006200061682620005a4565b9050919050565b620006288162000609565b81146200063457600080fd5b50565b60008151905062000648816200061d565b92915050565b60008060008060008060c087890312156200066e576200066d62000398565b5b600087015167ffffffffffffffff8111156200068f576200068e6200039d565b5b6200069d89828a01620004f8565b965050602087015167ffffffffffffffff811115620006c157620006c06200039d565b5b620006cf89828a01620004f8565b9550506040620006e289828a0162000552565b9450506060620006f589828a016200058d565b93505060806200070889828a01620005f2565b92505060a06200071b89828a0162000637565b9150509295509295509295565b600082825260208201905092915050565b7f45524332303a206d696e7420746f20746865207a65726f206164647265737300600082015250565b600062000771601f8362000728565b91506200077e8262000739565b602082019050919050565b60006020820190508181036000830152620007a48162000762565b9050919050565b620007b68162000569565b82525050565b6000602082019050620007d36000830184620007ab565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000620008158262000569565b9150620008228362000569565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156200085a5762000859620007d9565b5b828201905092915050565b7f64732d6d6174682d6164642d6f766572666c6f77000000000000000000000000600082015250565b60006200089d60148362000728565b9150620008aa8262000865565b602082019050919050565b60006020820190508181036000830152620008d0816200088e565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b600060028204905060018216806200091f57607f821691505b60208210811415620009365762000935620008d7565b5b50919050565b61151d806200094c6000396000f3fe608060405234801561001057600080fd5b50600436106100b45760003560e01c80634e6ec247116100715780634e6ec247146101a357806370a08231146101bf57806395d89b41146101ef578063a457c2d71461020d578063a9059cbb1461023d578063dd62ed3e1461026d576100b4565b806306fdde03146100b9578063095ea7b3146100d757806318160ddd1461010757806323b872dd14610125578063313ce567146101555780633950935114610173575b600080fd5b6100c161029d565b6040516100ce9190610def565b60405180910390f35b6100f160048036038101906100ec9190610eaa565b61032f565b6040516100fe9190610f05565b60405180910390f35b61010f610346565b60405161011c9190610f2f565b60405180910390f35b61013f600480360381019061013a9190610f4a565b610350565b60405161014c9190610f05565b60405180910390f35b61015d610401565b60405161016a9190610fb9565b60405180910390f35b61018d60048036038101906101889190610eaa565b610418565b60405161019a9190610f05565b60405180910390f35b6101bd60048036038101906101b89190610eaa565b6104bd565b005b6101d960048036038101906101d49190610fd4565b610647565b6040516101e69190610f2f565b60405180910390f35b6101f7610690565b6040516102049190610def565b60405180910390f35b61022760048036038101906102229190610eaa565b610722565b6040516102349190610f05565b60405180910390f35b61025760048036038101906102529190610eaa565b6107c7565b6040516102649190610f05565b60405180910390f35b61028760048036038101906102829190611001565b6107de565b6040516102949190610f2f565b60405180910390f35b6060600080546102ac90611070565b80601f01602080910402602001604051908101604052809291908181526020018280546102d890611070565b80156103255780601f106102fa57610100808354040283529160200191610325565b820191906000526020600020905b81548152906001019060200180831161030857829003601f168201915b5050505050905090565b600061033c3384846108be565b6001905092915050565b6000600354905090565b600061035d848484610a89565b6103f684336103f185600560008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610cfd90919063ffffffff16565b6108be565b600190509392505050565b6000600260009054906101000a900460ff16905090565b60006104b333846104ae85600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205461086590919063ffffffff16565b6108be565b6001905092915050565b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16141561052d576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610524906110ee565b60405180910390fd5b6105428160035461086590919063ffffffff16565b60038190555061059a81600460008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205461086590919063ffffffff16565b600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508173ffffffffffffffffffffffffffffffffffffffff16600073ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef8360405161063b9190610f2f565b60405180910390a35050565b6000600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b60606001805461069f90611070565b80601f01602080910402602001604051908101604052809291908181526020018280546106cb90611070565b80156107185780601f106106ed57610100808354040283529160200191610718565b820191906000526020600020905b8154815290600101906020018083116106fb57829003601f168201915b5050505050905090565b60006107bd33846107b885600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610cfd90919063ffffffff16565b6108be565b6001905092915050565b60006107d4338484610a89565b6001905092915050565b6000600560008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054905092915050565b6000828284610874919061113d565b91508110156108b8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108af906111df565b60405180910390fd5b92915050565b600073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16141561092e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161092590611271565b60405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16141561099e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161099590611303565b60405180910390fd5b80600560008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92583604051610a7c9190610f2f565b60405180910390a3505050565b600073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161415610af9576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610af090611395565b60405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff161415610b69576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b6090611427565b60405180910390fd5b610bbb81600460008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610cfd90919063ffffffff16565b600460008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550610c5081600460008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205461086590919063ffffffff16565b600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051610cf09190610f2f565b60405180910390a3505050565b6000828284610d0c9190611447565b9150811115610d50576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610d47906114c7565b60405180910390fd5b92915050565b600081519050919050565b600082825260208201905092915050565b60005b83811015610d90578082015181840152602081019050610d75565b83811115610d9f576000848401525b50505050565b6000601f19601f8301169050919050565b6000610dc182610d56565b610dcb8185610d61565b9350610ddb818560208601610d72565b610de481610da5565b840191505092915050565b60006020820190508181036000830152610e098184610db6565b905092915050565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610e4182610e16565b9050919050565b610e5181610e36565b8114610e5c57600080fd5b50565b600081359050610e6e81610e48565b92915050565b6000819050919050565b610e8781610e74565b8114610e9257600080fd5b50565b600081359050610ea481610e7e565b92915050565b60008060408385031215610ec157610ec0610e11565b5b6000610ecf85828601610e5f565b9250506020610ee085828601610e95565b9150509250929050565b60008115159050919050565b610eff81610eea565b82525050565b6000602082019050610f1a6000830184610ef6565b92915050565b610f2981610e74565b82525050565b6000602082019050610f446000830184610f20565b92915050565b600080600060608486031215610f6357610f62610e11565b5b6000610f7186828701610e5f565b9350506020610f8286828701610e5f565b9250506040610f9386828701610e95565b9150509250925092565b600060ff82169050919050565b610fb381610f9d565b82525050565b6000602082019050610fce6000830184610faa565b92915050565b600060208284031215610fea57610fe9610e11565b5b6000610ff884828501610e5f565b91505092915050565b6000806040838503121561101857611017610e11565b5b600061102685828601610e5f565b925050602061103785828601610e5f565b9150509250929050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b6000600282049050600182168061108857607f821691505b6020821081141561109c5761109b611041565b5b50919050565b7f45524332303a206d696e7420746f20746865207a65726f206164647265737300600082015250565b60006110d8601f83610d61565b91506110e3826110a2565b602082019050919050565b60006020820190508181036000830152611107816110cb565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061114882610e74565b915061115383610e74565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156111885761118761110e565b5b828201905092915050565b7f64732d6d6174682d6164642d6f766572666c6f77000000000000000000000000600082015250565b60006111c9601483610d61565b91506111d482611193565b602082019050919050565b600060208201905081810360008301526111f8816111bc565b9050919050565b7f45524332303a20617070726f76652066726f6d20746865207a65726f2061646460008201527f7265737300000000000000000000000000000000000000000000000000000000602082015250565b600061125b602483610d61565b9150611266826111ff565b604082019050919050565b6000602082019050818103600083015261128a8161124e565b9050919050565b7f45524332303a20617070726f766520746f20746865207a65726f20616464726560008201527f7373000000000000000000000000000000000000000000000000000000000000602082015250565b60006112ed602283610d61565b91506112f882611291565b604082019050919050565b6000602082019050818103600083015261131c816112e0565b9050919050565b7f45524332303a207472616e736665722066726f6d20746865207a65726f20616460008201527f6472657373000000000000000000000000000000000000000000000000000000602082015250565b600061137f602583610d61565b915061138a82611323565b604082019050919050565b600060208201905081810360008301526113ae81611372565b9050919050565b7f45524332303a207472616e7366657220746f20746865207a65726f206164647260008201527f6573730000000000000000000000000000000000000000000000000000000000602082015250565b6000611411602383610d61565b915061141c826113b5565b604082019050919050565b6000602082019050818103600083015261144081611404565b9050919050565b600061145282610e74565b915061145d83610e74565b9250828210156114705761146f61110e565b5b828203905092915050565b7f64732d6d6174682d7375622d756e646572666c6f770000000000000000000000600082015250565b60006114b1601583610d61565b91506114bc8261147b565b602082019050919050565b600060208201905081810360008301526114e0816114a4565b905091905056fea2646970667358221220206ee5db59571ba825d922503b87562d8c5f4942b30b7e067e14789d9f769f9e64736f6c634300080b003300000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000001200000000000000000000000000000125dfa371a19e6f7cb54395ca0000000000000000000000000000000000bbe4733d85bc2b90682147779da49cab38c0aa1f000000000000000000000000bbe4733d85bc2b90682147779da49cab38c0aa1f00000000000000000000000000000000000000000000000000000000000000054f4950323000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000094f49503230205354440000000000000000000000000000000000000000000000"
	transferPayload = "a9059cbb00000000000000000000000083d83497431c2d3feab296a9fba4e5fadd2f7ed000000000000000000000000000000000000000000000152d02c7e14af6800000"
	oip20Address    = "0x1033796B018B2bf0Fc9CB88c0793b2F275eDB624" //"0x45dD91b0289E60D89Cec94dF0Aac3a2f539c514a"
)

var hexKeys = []string{
	"8ff3ca2d9985c3a52b459e2f6e7822b23e1af845961e22128d5f372fb9aa5f17", //0xbbE4733d85bc2b90682147779DA49caB38C0aA1F
	"171786c73f805d257ceb07206d851eea30b3b41a2170ae55e1225e0ad516ef42", //0x83D83497431C2D3FEab296a9fba4e5FaDD2f7eD0
	"b7700998b973a2cae0cb8e8a328171399c043e57289735aca5f2419bd622297a", //0x4C12e733e58819A1d3520f1E7aDCc614Ca20De64
	"00dcf944648491b3a822d40bf212f359f699ed0dd5ce5a60f1da5e1142855949", //0x2Bd4AF0C1D0c2930fEE852D07bB9dE87D8C07044
}

const hexNodeKey = "d322864e848a3ebbb88cbd45b163db3c479b166937f10a14ab86a3f860b0b0b64506fc928bd335f434691375f63d0baf97968716a20b2ad15463e51ba5cf49fe"

var (
	// flag
	txNum   uint64
	msgType int64

	cdc             *amino.Codec
	counterContract *Contract
	nodePrivKey     ed25519.PrivKeyEd25519
	nodeKey         []byte

	privateKeys []*ecdsa.PrivateKey
	address     []common.Address
)

func init() {
	flag.Uint64Var(&txNum, "num", 1e6, "tx num per account")
	flag.Int64Var(&msgType, "type", 1, "enable wtx to create wtx at same time")
	flag.Parse()

	cdc = amino.NewCodec()
	mempool.RegisterMessages(cdc)
	counterContract = newContract("counter", "0x45dD91b0289E60D89Cec94dF0Aac3a2f539c514a", abiFile, binFile)
	b, _ := hex.DecodeString(hexNodeKey)
	copy(nodePrivKey[:], b)
	nodeKey = nodePrivKey.PubKey().Bytes()

	privateKeys = make([]*ecdsa.PrivateKey, len(hexKeys))
	address = make([]common.Address, len(hexKeys))
	for i := range hexKeys {
		privateKey, err := crypto.HexToECDSA(hexKeys[i])
		if err != nil {
			panic("failed to switch unencrypted private key -> secp256k1 private key:" + err.Error())
		}
		privateKeys[i] = privateKey
		address[i] = crypto.PubkeyToAddress(privateKey.PublicKey)
	}
}

func main() {
	start := time.Now()
	var wg sync.WaitGroup
	for i := range hexKeys {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if err := createOip20Txs(index); err != nil {
				fmt.Println("createTxs error:", err, "index:", index)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("time cost:", time.Since(start))
}

func createTxs(index int) error {
	f, err := os.OpenFile(fmt.Sprintf("TxMessage-%s.txt", address[index]), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	txWriter := bufio.NewWriter(f)
	defer txWriter.Flush()

	var wtxWriter *bufio.Writer
	if msgType&2 == 2 {
		f2, err := os.OpenFile(fmt.Sprintf("WtxMessage-%s.txt", address[index]), os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer f2.Close()
		wtxWriter = bufio.NewWriter(f2)
		defer wtxWriter.Flush()
	}

	tx, nonce, err := createDeploy(index)
	if err != nil {
		return err
	}

	if err = writeTxMessage(txWriter, tx); err != nil {
		panic(err)
	}
	if err = writeWtxMessage(wtxWriter, tx, address[index].String()); err != nil {
		panic(err)
	}

	addData, _ := hex.DecodeString("1003e2d20000000000000000000000000000000000000000000000000000000000000064")
	subtractData, _ := hex.DecodeString("6deebae3")

	for {
		nonce++
		tx = createCall(index, nonce, addData)
		if err = writeTxMessage(txWriter, tx); err != nil {
			panic(err)
		}
		if err = writeWtxMessage(wtxWriter, tx, address[index].String()); err != nil {
			panic(err)
		}

		nonce++
		tx = createCall(index, nonce, subtractData)
		if err = writeTxMessage(txWriter, tx); err != nil {
			panic(err)
		}
		if err = writeWtxMessage(wtxWriter, tx, address[index].String()); err != nil {
			panic(err)
		}
		if nonce > txNum {
			break
		}
	}
	return nil
}

func createOip20Txs(index int) error {
	f, err := os.OpenFile(fmt.Sprintf("TxMessage-%s.txt", address[index]), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	txWriter := bufio.NewWriter(f)
	defer txWriter.Flush()

	tx, nonce, err := createOip20Deploy(index)
	if err != nil {
		return err
	}

	if err = writeTxMessage(txWriter, tx); err != nil {
		panic(err)
	}

	for {
		nonce++
		tx, err = createOip20Transfer(index, nonce, decodeHex(transferPayload))
		if err != nil {
			panic(err)
		}
		if err = writeTxMessage(txWriter, tx); err != nil {
			panic(err)
		}

		if nonce > txNum {
			break
		}
	}
	return nil
}

func createOip20Deploy(index int) (tx []byte, nonce uint64, err error) {
	if hexKeys[index] == "8ff3ca2d9985c3a52b459e2f6e7822b23e1af845961e22128d5f372fb9aa5f17" {
		nonce++
	}
	unsignedTx := types.NewContractCreation(nonce, big.NewInt(0), GasLimit+uint64(index), big.NewInt(GasPrice+int64(index)), decodeHex(oip20Code))
	signedTx, err := types.SignTx(unsignedTx, types.NewEIP155Signer(big.NewInt(ChainId)), privateKeys[index])
	if err != nil {
		return
	}

	tx, err = signedTx.MarshalBinary()
	return
}

func createOip20Transfer(index int, nonce uint64, data []byte) ([]byte, error) {
	unsignedTx := types.NewTransaction(nonce, common.HexToAddress(oip20Address), big.NewInt(0), GasLimit+uint64(index), big.NewInt(GasPrice+int64(index)), decodeHex(oip20Code))
	signedTx, err := types.SignTx(unsignedTx, types.NewEIP155Signer(big.NewInt(ChainId)), privateKeys[index])
	if err != nil {
		return nil, err
	}
	return signedTx.MarshalBinary()
}

func decodeHex(s string) []byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func writeTxMessage(w *bufio.Writer, tx []byte) error {
	if msgType&1 != 1 {
		return nil
	}
	msg := mempool.TxMessage{Tx: tx}
	if _, err := w.WriteString(hex.EncodeToString(cdc.MustMarshalBinaryBare(&msg))); err != nil {
		return err
	}
	return w.WriteByte('\n')
}

func writeWtxMessage(w *bufio.Writer, tx []byte, from string) error {
	if msgType&2 != 2 {
		return nil
	}
	wtx := &mempool.WrappedTx{
		Payload: tx,
		From:    from,
		NodeKey: nodeKey,
	}
	sig, err := nodePrivKey.Sign(append(wtx.Payload, wtx.From...))
	if err != nil {
		return err
	}
	wtx.Signature = sig

	msg := mempool.WtxMessage{Wtx: wtx}
	if _, err := w.WriteString(hex.EncodeToString(cdc.MustMarshalBinaryBare(&msg))); err != nil {
		return err
	}
	return w.WriteByte('\n')
}

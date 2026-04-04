// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// InstructionSenderMetaData contains all meta data concerning the InstructionSender contract.
var InstructionSenderMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_teeExtensionRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_teeMachineRegistry\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"OP_COMMAND_BINANCE_24H_STATS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OP_COMMAND_BINANCE_ACCOUNT_PNL\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OP_COMMAND_BINANCE_ACCOUNT_SUMMARY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OP_COMMAND_BINANCE_FETCH_AND_ATTEST\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OP_COMMAND_BINANCE_USER_PROFILE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OP_COMMAND_SIGN\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OP_COMMAND_UPDATE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OP_TYPE_KEY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OP_TYPE_MARKET\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"_extensionId\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"fetchBinance24hStatsAndAttest\",\"inputs\":[{\"name\":\"_message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"fetchBinanceAccountPnlAndAttest\",\"inputs\":[{\"name\":\"_message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"fetchBinanceAccountSummaryAndAttest\",\"inputs\":[{\"name\":\"_message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"fetchBinanceAndAttest\",\"inputs\":[{\"name\":\"_message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"fetchBinanceUserProfileAndAttest\",\"inputs\":[{\"name\":\"_message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"setExtensionId\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"sign\",\"inputs\":[{\"name\":\"_message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"teeExtensionRegistry\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contract ITeeExtensionRegistry\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"teeMachineRegistry\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contract ITeeMachineRegistry\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"updateKey\",\"inputs\":[{\"name\":\"_encryptedKey\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"payable\"}]",
	Bin: "0x60c060405234801561000f575f5ffd5b506040516122f63803806122f6833981810160405281019061003191906100fe565b8173ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250508073ffffffffffffffffffffffffffffffffffffffff1660a08173ffffffffffffffffffffffffffffffffffffffff1681525050505061013c565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6100cd826100a4565b9050919050565b6100dd816100c3565b81146100e7575f5ffd5b50565b5f815190506100f8816100d4565b92915050565b5f5f60408385031215610114576101136100a0565b5b5f610121858286016100ea565b9250506020610132858286016100ea565b9150509250929050565b60805160a0516121296101cd5f395f818161053301528181610722015281816107d301528181610a2b01528181610c5f0152818161112b01528181611364015261159801525f818161067b0152818161091b01528181610b7301528181610da701528181610e9601528181610efe01528181610fb101528181611273015281816114ac01526116e001526121295ff3fe60806040526004361061011e575f3560e01c8063786ba5131161009f578063c5028bbb11610063578063c5028bbb1461039c578063d313bf4e146103c6578063d473e270146103f6578063e6eb686714610420578063f5193e81146104505761011e565b8063786ba513146102d85780638d34b10514610308578063a15e520814610332578063a435d58a1461035c578063aa5032c6146103865761011e565b806356c4e670116100e657806356c4e670146101fa57806359ba4f3f1461022457806362e3a4401461024e57806376ba7fc61461027e57806376cd7cbc146102a85761011e565b80631b3bba9d146101225780631cd470321461014c57806320fc940714610176578063272a60b9146101a0578063524967d7146101d0575b5f5ffd5b34801561012d575f5ffd5b50610136610480565b60405161014391906117ef565b60405180910390f35b348015610157575f5ffd5b506101606104a4565b60405161016d91906117ef565b60405180910390f35b348015610181575f5ffd5b5061018a6104c8565b60405161019791906117ef565b60405180910390f35b6101ba60048036038101906101b5919061187a565b6104ec565b6040516101c791906117ef565b60405180910390f35b3480156101db575f5ffd5b506101e4610720565b6040516101f1919061193f565b60405180910390f35b348015610205575f5ffd5b5061020e610744565b60405161021b91906117ef565b60405180910390f35b34801561022f575f5ffd5b50610238610768565b60405161024591906117ef565b60405180910390f35b6102686004803603810190610263919061187a565b61078c565b60405161027591906117ef565b60405180910390f35b348015610289575f5ffd5b506102926109c0565b60405161029f91906117ef565b60405180910390f35b6102c260048036038101906102bd919061187a565b6109e4565b6040516102cf91906117ef565b60405180910390f35b6102f260048036038101906102ed919061187a565b610c18565b6040516102ff91906117ef565b60405180910390f35b348015610313575f5ffd5b5061031c610e4c565b60405161032991906117ef565b60405180910390f35b34801561033d575f5ffd5b50610346610e70565b60405161035391906117ef565b60405180910390f35b348015610367575f5ffd5b50610370610e94565b60405161037d9190611978565b60405180910390f35b348015610391575f5ffd5b5061039a610eb8565b005b3480156103a7575f5ffd5b506103b06110c0565b6040516103bd91906117ef565b60405180910390f35b6103e060048036038101906103db919061187a565b6110e4565b6040516103ed91906117ef565b60405180910390f35b348015610401575f5ffd5b5061040a611318565b60405161041791906119a9565b60405180910390f35b61043a6004803603810190610435919061187a565b61131d565b60405161044791906117ef565b60405180910390f35b61046a6004803603810190610465919061187a565b611551565b60405161047791906117ef565b60405180910390f35b7f4d41524b4554000000000000000000000000000000000000000000000000000081565b7f42494e414e43455f4143434f554e545f504e4c0000000000000000000000000081565b7f555044415445000000000000000000000000000000000000000000000000000081565b5f5f5f5403610530576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161052790611a1c565b60405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663feeabcbf5f5460016040518363ffffffff1660e01b815260040161058e929190611a73565b5f60405180830381865afa1580156105a8573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f820116820180604052508101906105d09190611c1d565b90506105da611785565b7f4d41524b45540000000000000000000000000000000000000000000000000000815f0181815250507f42494e414e43455f3234485f535441545300000000000000000000000000000081602001818152505084848080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f8201169050808301925050505050505081604001819052507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663f731df533484846040518463ffffffff1660e01b81526004016106d5929190611ea5565b60206040518083038185885af11580156106f1573d5f5f3e3d5ffd5b50505050506040513d601f19601f820116820180604052508101906107169190611f04565b9250505092915050565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f5349474e0000000000000000000000000000000000000000000000000000000081565b7f42494e414e43455f46455443485f414e445f415454455354000000000000000081565b5f5f5f54036107d0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107c790611a1c565b60405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663feeabcbf5f5460016040518363ffffffff1660e01b815260040161082e929190611a73565b5f60405180830381865afa158015610848573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f820116820180604052508101906108709190611c1d565b905061087a611785565b7f4d41524b45540000000000000000000000000000000000000000000000000000815f0181815250507f42494e414e43455f4143434f554e545f504e4c0000000000000000000000000081602001818152505084848080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f8201169050808301925050505050505081604001819052507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663f731df533484846040518463ffffffff1660e01b8152600401610975929190611ea5565b60206040518083038185885af1158015610991573d5f5f3e3d5ffd5b50505050506040513d601f19601f820116820180604052508101906109b69190611f04565b9250505092915050565b7f42494e414e43455f3234485f535441545300000000000000000000000000000081565b5f5f5f5403610a28576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610a1f90611a1c565b60405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663feeabcbf5f5460016040518363ffffffff1660e01b8152600401610a86929190611a73565b5f60405180830381865afa158015610aa0573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f82011682018060405250810190610ac89190611c1d565b9050610ad2611785565b7f4b45590000000000000000000000000000000000000000000000000000000000815f0181815250507f5349474e0000000000000000000000000000000000000000000000000000000081602001818152505084848080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f8201169050808301925050505050505081604001819052507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663f731df533484846040518463ffffffff1660e01b8152600401610bcd929190611ea5565b60206040518083038185885af1158015610be9573d5f5f3e3d5ffd5b50505050506040513d601f19601f82011682018060405250810190610c0e9190611f04565b9250505092915050565b5f5f5f5403610c5c576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c5390611a1c565b60405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663feeabcbf5f5460016040518363ffffffff1660e01b8152600401610cba929190611a73565b5f60405180830381865afa158015610cd4573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f82011682018060405250810190610cfc9190611c1d565b9050610d06611785565b7f4d41524b45540000000000000000000000000000000000000000000000000000815f0181815250507f42494e414e43455f46455443485f414e445f415454455354000000000000000081602001818152505084848080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f8201169050808301925050505050505081604001819052507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663f731df533484846040518463ffffffff1660e01b8152600401610e01929190611ea5565b60206040518083038185885af1158015610e1d573d5f5f3e3d5ffd5b50505050506040513d601f19601f82011682018060405250810190610e429190611f04565b9250505092915050565b7f42494e414e43455f555345525f50524f46494c4500000000000000000000000081565b7f42494e414e43455f4143434f554e545f53554d4d41525900000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b5f5f5414610efb576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610ef290611f79565b60405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663fad5902b6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610f65573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610f899190611fc1565b90505f600190505b818111611082573073ffffffffffffffffffffffffffffffffffffffff167f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16632c177358836040518263ffffffff1660e01b815260040161100891906119a9565b602060405180830381865afa158015611023573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906110479190611fec565b73ffffffffffffffffffffffffffffffffffffffff160361106f57805f8190555050506110be565b808061107a90612044565b915050610f91565b506040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016110b5906120d5565b60405180910390fd5b565b7f4b4559000000000000000000000000000000000000000000000000000000000081565b5f5f5f5403611128576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161111f90611a1c565b60405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663feeabcbf5f5460016040518363ffffffff1660e01b8152600401611186929190611a73565b5f60405180830381865afa1580156111a0573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f820116820180604052508101906111c89190611c1d565b90506111d2611785565b7f4d41524b45540000000000000000000000000000000000000000000000000000815f0181815250507f42494e414e43455f555345525f50524f46494c4500000000000000000000000081602001818152505084848080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f8201169050808301925050505050505081604001819052507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663f731df533484846040518463ffffffff1660e01b81526004016112cd929190611ea5565b60206040518083038185885af11580156112e9573d5f5f3e3d5ffd5b50505050506040513d601f19601f8201168201806040525081019061130e9190611f04565b9250505092915050565b5f5481565b5f5f5f5403611361576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161135890611a1c565b60405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663feeabcbf5f5460016040518363ffffffff1660e01b81526004016113bf929190611a73565b5f60405180830381865afa1580156113d9573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f820116820180604052508101906114019190611c1d565b905061140b611785565b7f4b45590000000000000000000000000000000000000000000000000000000000815f0181815250507f555044415445000000000000000000000000000000000000000000000000000081602001818152505084848080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f8201169050808301925050505050505081604001819052507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663f731df533484846040518463ffffffff1660e01b8152600401611506929190611ea5565b60206040518083038185885af1158015611522573d5f5f3e3d5ffd5b50505050506040513d601f19601f820116820180604052508101906115479190611f04565b9250505092915050565b5f5f5f5403611595576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161158c90611a1c565b60405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663feeabcbf5f5460016040518363ffffffff1660e01b81526004016115f3929190611a73565b5f60405180830381865afa15801561160d573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f820116820180604052508101906116359190611c1d565b905061163f611785565b7f4d41524b45540000000000000000000000000000000000000000000000000000815f0181815250507f42494e414e43455f4143434f554e545f53554d4d41525900000000000000000081602001818152505084848080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f8201169050808301925050505050505081604001819052507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663f731df533484846040518463ffffffff1660e01b815260040161173a929190611ea5565b60206040518083038185885af1158015611756573d5f5f3e3d5ffd5b50505050506040513d601f19601f8201168201806040525081019061177b9190611f04565b9250505092915050565b6040518060c001604052805f81526020015f815260200160608152602001606081526020015f67ffffffffffffffff1681526020015f73ffffffffffffffffffffffffffffffffffffffff1681525090565b5f819050919050565b6117e9816117d7565b82525050565b5f6020820190506118025f8301846117e0565b92915050565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f5f83601f84011261183a57611839611819565b5b8235905067ffffffffffffffff8111156118575761185661181d565b5b60208301915083600182028301111561187357611872611821565b5b9250929050565b5f5f602083850312156118905761188f611811565b5b5f83013567ffffffffffffffff8111156118ad576118ac611815565b5b6118b985828601611825565b92509250509250929050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f819050919050565b5f6119076119026118fd846118c5565b6118e4565b6118c5565b9050919050565b5f611918826118ed565b9050919050565b5f6119298261190e565b9050919050565b6119398161191f565b82525050565b5f6020820190506119525f830184611930565b92915050565b5f6119628261190e565b9050919050565b61197281611958565b82525050565b5f60208201905061198b5f830184611969565b92915050565b5f819050919050565b6119a381611991565b82525050565b5f6020820190506119bc5f83018461199a565b92915050565b5f82825260208201905092915050565b7f657874656e73696f6e204944206e6f74207365740000000000000000000000005f82015250565b5f611a066014836119c2565b9150611a11826119d2565b602082019050919050565b5f6020820190508181035f830152611a33816119fa565b9050919050565b5f819050919050565b5f611a5d611a58611a5384611a3a565b6118e4565b611991565b9050919050565b611a6d81611a43565b82525050565b5f604082019050611a865f83018561199a565b611a936020830184611a64565b9392505050565b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b611ae082611a9a565b810181811067ffffffffffffffff82111715611aff57611afe611aaa565b5b80604052505050565b5f611b11611808565b9050611b1d8282611ad7565b919050565b5f67ffffffffffffffff821115611b3c57611b3b611aaa565b5b602082029050602081019050919050565b5f611b57826118c5565b9050919050565b611b6781611b4d565b8114611b71575f5ffd5b50565b5f81519050611b8281611b5e565b92915050565b5f611b9a611b9584611b22565b611b08565b90508083825260208201905060208402830185811115611bbd57611bbc611821565b5b835b81811015611be65780611bd28882611b74565b845260208401935050602081019050611bbf565b5050509392505050565b5f82601f830112611c0457611c03611819565b5b8151611c14848260208601611b88565b91505092915050565b5f60208284031215611c3257611c31611811565b5b5f82015167ffffffffffffffff811115611c4f57611c4e611815565b5b611c5b84828501611bf0565b91505092915050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b611c9681611b4d565b82525050565b5f611ca78383611c8d565b60208301905092915050565b5f602082019050919050565b5f611cc982611c64565b611cd38185611c6e565b9350611cde83611c7e565b805f5b83811015611d0e578151611cf58882611c9c565b9750611d0083611cb3565b925050600181019050611ce1565b5085935050505092915050565b611d24816117d7565b82525050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f611d5c82611d2a565b611d668185611d34565b9350611d76818560208601611d44565b611d7f81611a9a565b840191505092915050565b5f82825260208201905092915050565b5f611da482611c64565b611dae8185611d8a565b9350611db983611c7e565b805f5b83811015611de9578151611dd08882611c9c565b9750611ddb83611cb3565b925050600181019050611dbc565b5085935050505092915050565b5f67ffffffffffffffff82169050919050565b611e1281611df6565b82525050565b5f60c083015f830151611e2d5f860182611d1b565b506020830151611e406020860182611d1b565b5060408301518482036040860152611e588282611d52565b91505060608301518482036060860152611e728282611d9a565b9150506080830151611e876080860182611e09565b5060a0830151611e9a60a0860182611c8d565b508091505092915050565b5f6040820190508181035f830152611ebd8185611cbf565b90508181036020830152611ed18184611e18565b90509392505050565b611ee3816117d7565b8114611eed575f5ffd5b50565b5f81519050611efe81611eda565b92915050565b5f60208284031215611f1957611f18611811565b5b5f611f2684828501611ef0565b91505092915050565b7f657874656e73696f6e20494420616c72656164792073657400000000000000005f82015250565b5f611f636018836119c2565b9150611f6e82611f2f565b602082019050919050565b5f6020820190508181035f830152611f9081611f57565b9050919050565b611fa081611991565b8114611faa575f5ffd5b50565b5f81519050611fbb81611f97565b92915050565b5f60208284031215611fd657611fd5611811565b5b5f611fe384828501611fad565b91505092915050565b5f6020828403121561200157612000611811565b5b5f61200e84828501611b74565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61204e82611991565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82036120805761207f612017565b5b600182019050919050565b7f657874656e73696f6e204944206e6f7420666f756e64000000000000000000005f82015250565b5f6120bf6016836119c2565b91506120ca8261208b565b602082019050919050565b5f6020820190508181035f8301526120ec816120b3565b905091905056fea2646970667358221220126fb0eca8961c99a375f56e22dc6c3528d6363341aed42f22ee145a3e640d6364736f6c634300081f0033",
}

// InstructionSenderABI is the input ABI used to generate the binding from.
// Deprecated: Use InstructionSenderMetaData.ABI instead.
var InstructionSenderABI = InstructionSenderMetaData.ABI

// InstructionSenderBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use InstructionSenderMetaData.Bin instead.
var InstructionSenderBin = InstructionSenderMetaData.Bin

// DeployInstructionSender deploys a new Ethereum contract, binding an instance of InstructionSender to it.
func DeployInstructionSender(auth *bind.TransactOpts, backend bind.ContractBackend, _teeExtensionRegistry common.Address, _teeMachineRegistry common.Address) (common.Address, *types.Transaction, *InstructionSender, error) {
	parsed, err := InstructionSenderMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(InstructionSenderBin), backend, _teeExtensionRegistry, _teeMachineRegistry)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &InstructionSender{InstructionSenderCaller: InstructionSenderCaller{contract: contract}, InstructionSenderTransactor: InstructionSenderTransactor{contract: contract}, InstructionSenderFilterer: InstructionSenderFilterer{contract: contract}}, nil
}

// InstructionSender is an auto generated Go binding around an Ethereum contract.
type InstructionSender struct {
	InstructionSenderCaller     // Read-only binding to the contract
	InstructionSenderTransactor // Write-only binding to the contract
	InstructionSenderFilterer   // Log filterer for contract events
}

// InstructionSenderCaller is an auto generated read-only Go binding around an Ethereum contract.
type InstructionSenderCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InstructionSenderTransactor is an auto generated write-only Go binding around an Ethereum contract.
type InstructionSenderTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InstructionSenderFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type InstructionSenderFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InstructionSenderSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type InstructionSenderSession struct {
	Contract     *InstructionSender // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// InstructionSenderCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type InstructionSenderCallerSession struct {
	Contract *InstructionSenderCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// InstructionSenderTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type InstructionSenderTransactorSession struct {
	Contract     *InstructionSenderTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// InstructionSenderRaw is an auto generated low-level Go binding around an Ethereum contract.
type InstructionSenderRaw struct {
	Contract *InstructionSender // Generic contract binding to access the raw methods on
}

// InstructionSenderCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type InstructionSenderCallerRaw struct {
	Contract *InstructionSenderCaller // Generic read-only contract binding to access the raw methods on
}

// InstructionSenderTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type InstructionSenderTransactorRaw struct {
	Contract *InstructionSenderTransactor // Generic write-only contract binding to access the raw methods on
}

// NewInstructionSender creates a new instance of InstructionSender, bound to a specific deployed contract.
func NewInstructionSender(address common.Address, backend bind.ContractBackend) (*InstructionSender, error) {
	contract, err := bindInstructionSender(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &InstructionSender{InstructionSenderCaller: InstructionSenderCaller{contract: contract}, InstructionSenderTransactor: InstructionSenderTransactor{contract: contract}, InstructionSenderFilterer: InstructionSenderFilterer{contract: contract}}, nil
}

// NewInstructionSenderCaller creates a new read-only instance of InstructionSender, bound to a specific deployed contract.
func NewInstructionSenderCaller(address common.Address, caller bind.ContractCaller) (*InstructionSenderCaller, error) {
	contract, err := bindInstructionSender(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &InstructionSenderCaller{contract: contract}, nil
}

// NewInstructionSenderTransactor creates a new write-only instance of InstructionSender, bound to a specific deployed contract.
func NewInstructionSenderTransactor(address common.Address, transactor bind.ContractTransactor) (*InstructionSenderTransactor, error) {
	contract, err := bindInstructionSender(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &InstructionSenderTransactor{contract: contract}, nil
}

// NewInstructionSenderFilterer creates a new log filterer instance of InstructionSender, bound to a specific deployed contract.
func NewInstructionSenderFilterer(address common.Address, filterer bind.ContractFilterer) (*InstructionSenderFilterer, error) {
	contract, err := bindInstructionSender(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &InstructionSenderFilterer{contract: contract}, nil
}

// bindInstructionSender binds a generic wrapper to an already deployed contract.
func bindInstructionSender(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := InstructionSenderMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InstructionSender *InstructionSenderRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InstructionSender.Contract.InstructionSenderCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InstructionSender *InstructionSenderRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InstructionSender.Contract.InstructionSenderTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InstructionSender *InstructionSenderRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InstructionSender.Contract.InstructionSenderTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InstructionSender *InstructionSenderCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InstructionSender.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InstructionSender *InstructionSenderTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InstructionSender.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InstructionSender *InstructionSenderTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InstructionSender.Contract.contract.Transact(opts, method, params...)
}

// ExtensionId is a free data retrieval call binding the contract method 0xd473e270.
//
// Solidity: function _extensionId() view returns(uint256)
func (_InstructionSender *InstructionSenderCaller) ExtensionId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InstructionSender.contract.Call(opts, &out, "_extensionId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ExtensionId is a free data retrieval call binding the contract method 0xd473e270.
//
// Solidity: function _extensionId() view returns(uint256)
func (_InstructionSender *InstructionSenderSession) ExtensionId() (*big.Int, error) {
	return _InstructionSender.Contract.ExtensionId(&_InstructionSender.CallOpts)
}

// ExtensionId is a free data retrieval call binding the contract method 0xd473e270.
//
// Solidity: function _extensionId() view returns(uint256)
func (_InstructionSender *InstructionSenderCallerSession) ExtensionId() (*big.Int, error) {
	return _InstructionSender.Contract.ExtensionId(&_InstructionSender.CallOpts)
}

// TeeExtensionRegistry is a free data retrieval call binding the contract method 0xa435d58a.
//
// Solidity: function teeExtensionRegistry() view returns(address)
func (_InstructionSender *InstructionSenderCaller) TeeExtensionRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InstructionSender.contract.Call(opts, &out, "teeExtensionRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TeeExtensionRegistry is a free data retrieval call binding the contract method 0xa435d58a.
//
// Solidity: function teeExtensionRegistry() view returns(address)
func (_InstructionSender *InstructionSenderSession) TeeExtensionRegistry() (common.Address, error) {
	return _InstructionSender.Contract.TeeExtensionRegistry(&_InstructionSender.CallOpts)
}

// TeeExtensionRegistry is a free data retrieval call binding the contract method 0xa435d58a.
//
// Solidity: function teeExtensionRegistry() view returns(address)
func (_InstructionSender *InstructionSenderCallerSession) TeeExtensionRegistry() (common.Address, error) {
	return _InstructionSender.Contract.TeeExtensionRegistry(&_InstructionSender.CallOpts)
}

// TeeMachineRegistry is a free data retrieval call binding the contract method 0x524967d7.
//
// Solidity: function teeMachineRegistry() view returns(address)
func (_InstructionSender *InstructionSenderCaller) TeeMachineRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InstructionSender.contract.Call(opts, &out, "teeMachineRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TeeMachineRegistry is a free data retrieval call binding the contract method 0x524967d7.
//
// Solidity: function teeMachineRegistry() view returns(address)
func (_InstructionSender *InstructionSenderSession) TeeMachineRegistry() (common.Address, error) {
	return _InstructionSender.Contract.TeeMachineRegistry(&_InstructionSender.CallOpts)
}

// TeeMachineRegistry is a free data retrieval call binding the contract method 0x524967d7.
//
// Solidity: function teeMachineRegistry() view returns(address)
func (_InstructionSender *InstructionSenderCallerSession) TeeMachineRegistry() (common.Address, error) {
	return _InstructionSender.Contract.TeeMachineRegistry(&_InstructionSender.CallOpts)
}

// SetExtensionId is a paid mutator transaction binding the contract method 0xaa5032c6.
//
// Solidity: function setExtensionId() returns()
func (_InstructionSender *InstructionSenderTransactor) SetExtensionId(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InstructionSender.contract.Transact(opts, "setExtensionId")
}

// SetExtensionId is a paid mutator transaction binding the contract method 0xaa5032c6.
//
// Solidity: function setExtensionId() returns()
func (_InstructionSender *InstructionSenderSession) SetExtensionId() (*types.Transaction, error) {
	return _InstructionSender.Contract.SetExtensionId(&_InstructionSender.TransactOpts)
}

// SetExtensionId is a paid mutator transaction binding the contract method 0xaa5032c6.
//
// Solidity: function setExtensionId() returns()
func (_InstructionSender *InstructionSenderTransactorSession) SetExtensionId() (*types.Transaction, error) {
	return _InstructionSender.Contract.SetExtensionId(&_InstructionSender.TransactOpts)
}

// Sign is a paid mutator transaction binding the contract method 0x76cd7cbc.
//
// Solidity: function sign(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactor) Sign(opts *bind.TransactOpts, _message []byte) (*types.Transaction, error) {
	return _InstructionSender.contract.Transact(opts, "sign", _message)
}

// Sign is a paid mutator transaction binding the contract method 0x76cd7cbc.
//
// Solidity: function sign(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderSession) Sign(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.Sign(&_InstructionSender.TransactOpts, _message)
}

// Sign is a paid mutator transaction binding the contract method 0x76cd7cbc.
//
// Solidity: function sign(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactorSession) Sign(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.Sign(&_InstructionSender.TransactOpts, _message)
}

// UpdateKey is a paid mutator transaction binding the contract method 0xe6eb6867.
//
// Solidity: function updateKey(bytes _encryptedKey) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactor) UpdateKey(opts *bind.TransactOpts, _encryptedKey []byte) (*types.Transaction, error) {
	return _InstructionSender.contract.Transact(opts, "updateKey", _encryptedKey)
}

// UpdateKey is a paid mutator transaction binding the contract method 0xe6eb6867.
//
// Solidity: function updateKey(bytes _encryptedKey) payable returns(bytes32)
func (_InstructionSender *InstructionSenderSession) UpdateKey(_encryptedKey []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.UpdateKey(&_InstructionSender.TransactOpts, _encryptedKey)
}

// UpdateKey is a paid mutator transaction binding the contract method 0xe6eb6867.
//
// Solidity: function updateKey(bytes _encryptedKey) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactorSession) UpdateKey(_encryptedKey []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.UpdateKey(&_InstructionSender.TransactOpts, _encryptedKey)
}

// FetchBinanceAndAttest is a paid mutator transaction binding the contract method for fetchBinanceAndAttest.
//
// Solidity: function fetchBinanceAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactor) FetchBinanceAndAttest(opts *bind.TransactOpts, _message []byte) (*types.Transaction, error) {
	return _InstructionSender.contract.Transact(opts, "fetchBinanceAndAttest", _message)
}

// FetchBinanceAndAttest is a paid mutator transaction binding the contract method for fetchBinanceAndAttest.
//
// Solidity: function fetchBinanceAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderSession) FetchBinanceAndAttest(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.FetchBinanceAndAttest(&_InstructionSender.TransactOpts, _message)
}

// FetchBinanceAndAttest is a paid mutator transaction binding the contract method for fetchBinanceAndAttest.
//
// Solidity: function fetchBinanceAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactorSession) FetchBinanceAndAttest(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.FetchBinanceAndAttest(&_InstructionSender.TransactOpts, _message)
}

// FetchBinance24hStatsAndAttest is a paid mutator transaction binding the contract method for fetchBinance24hStatsAndAttest.
//
// Solidity: function fetchBinance24hStatsAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactor) FetchBinance24hStatsAndAttest(opts *bind.TransactOpts, _message []byte) (*types.Transaction, error) {
	return _InstructionSender.contract.Transact(opts, "fetchBinance24hStatsAndAttest", _message)
}

// FetchBinance24hStatsAndAttest is a paid mutator transaction binding the contract method for fetchBinance24hStatsAndAttest.
//
// Solidity: function fetchBinance24hStatsAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderSession) FetchBinance24hStatsAndAttest(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.FetchBinance24hStatsAndAttest(&_InstructionSender.TransactOpts, _message)
}

// FetchBinance24hStatsAndAttest is a paid mutator transaction binding the contract method for fetchBinance24hStatsAndAttest.
//
// Solidity: function fetchBinance24hStatsAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactorSession) FetchBinance24hStatsAndAttest(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.FetchBinance24hStatsAndAttest(&_InstructionSender.TransactOpts, _message)
}

// FetchBinanceAccountPnlAndAttest is a paid mutator transaction binding the contract method for fetchBinanceAccountPnlAndAttest.
//
// Solidity: function fetchBinanceAccountPnlAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactor) FetchBinanceAccountPnlAndAttest(opts *bind.TransactOpts, _message []byte) (*types.Transaction, error) {
	return _InstructionSender.contract.Transact(opts, "fetchBinanceAccountPnlAndAttest", _message)
}

// FetchBinanceAccountPnlAndAttest is a paid mutator transaction binding the contract method for fetchBinanceAccountPnlAndAttest.
//
// Solidity: function fetchBinanceAccountPnlAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderSession) FetchBinanceAccountPnlAndAttest(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.FetchBinanceAccountPnlAndAttest(&_InstructionSender.TransactOpts, _message)
}

// FetchBinanceAccountPnlAndAttest is a paid mutator transaction binding the contract method for fetchBinanceAccountPnlAndAttest.
//
// Solidity: function fetchBinanceAccountPnlAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactorSession) FetchBinanceAccountPnlAndAttest(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.FetchBinanceAccountPnlAndAttest(&_InstructionSender.TransactOpts, _message)
}

// FetchBinanceAccountSummaryAndAttest is a paid mutator transaction binding the contract method for fetchBinanceAccountSummaryAndAttest.
//
// Solidity: function fetchBinanceAccountSummaryAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactor) FetchBinanceAccountSummaryAndAttest(opts *bind.TransactOpts, _message []byte) (*types.Transaction, error) {
	return _InstructionSender.contract.Transact(opts, "fetchBinanceAccountSummaryAndAttest", _message)
}

// FetchBinanceAccountSummaryAndAttest is a paid mutator transaction binding the contract method for fetchBinanceAccountSummaryAndAttest.
//
// Solidity: function fetchBinanceAccountSummaryAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderSession) FetchBinanceAccountSummaryAndAttest(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.FetchBinanceAccountSummaryAndAttest(&_InstructionSender.TransactOpts, _message)
}

// FetchBinanceAccountSummaryAndAttest is a paid mutator transaction binding the contract method for fetchBinanceAccountSummaryAndAttest.
//
// Solidity: function fetchBinanceAccountSummaryAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactorSession) FetchBinanceAccountSummaryAndAttest(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.FetchBinanceAccountSummaryAndAttest(&_InstructionSender.TransactOpts, _message)
}

// FetchBinanceUserProfileAndAttest is a paid mutator transaction binding the contract method for fetchBinanceUserProfileAndAttest.
//
// Solidity: function fetchBinanceUserProfileAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactor) FetchBinanceUserProfileAndAttest(opts *bind.TransactOpts, _message []byte) (*types.Transaction, error) {
	return _InstructionSender.contract.Transact(opts, "fetchBinanceUserProfileAndAttest", _message)
}

// FetchBinanceUserProfileAndAttest is a paid mutator transaction binding the contract method for fetchBinanceUserProfileAndAttest.
//
// Solidity: function fetchBinanceUserProfileAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderSession) FetchBinanceUserProfileAndAttest(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.FetchBinanceUserProfileAndAttest(&_InstructionSender.TransactOpts, _message)
}

// FetchBinanceUserProfileAndAttest is a paid mutator transaction binding the contract method for fetchBinanceUserProfileAndAttest.
//
// Solidity: function fetchBinanceUserProfileAndAttest(bytes _message) payable returns(bytes32)
func (_InstructionSender *InstructionSenderTransactorSession) FetchBinanceUserProfileAndAttest(_message []byte) (*types.Transaction, error) {
	return _InstructionSender.Contract.FetchBinanceUserProfileAndAttest(&_InstructionSender.TransactOpts, _message)
}

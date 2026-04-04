// SPDX-License-Identifier: MIT
pragma solidity >=0.7.6 <0.9;

import { ITeeExtensionRegistry } from "./interface/ITeeExtensionRegistry.sol";
import { ITeeMachineRegistry } from "./interface/ITeeMachineRegistry.sol";

contract InstructionSender {
    ITeeExtensionRegistry public immutable teeExtensionRegistry;
    ITeeMachineRegistry public immutable teeMachineRegistry;

    bytes32 public constant OP_TYPE_KEY = bytes32("KEY");
    bytes32 public constant OP_COMMAND_UPDATE = bytes32("UPDATE");
    bytes32 public constant OP_COMMAND_SIGN = bytes32("SIGN");
    bytes32 public constant OP_TYPE_MARKET = bytes32("MARKET");
    bytes32 public constant OP_COMMAND_BINANCE_FETCH_AND_ATTEST = bytes32("BINANCE_FETCH_AND_ATTEST");
    bytes32 public constant OP_COMMAND_BINANCE_24H_STATS = bytes32("BINANCE_24H_STATS");
    bytes32 public constant OP_COMMAND_BINANCE_ACCOUNT_PNL = bytes32("BINANCE_ACCOUNT_PNL");
    bytes32 public constant OP_COMMAND_BINANCE_ACCOUNT_SUMMARY = bytes32("BINANCE_ACCOUNT_SUMMARY");
    bytes32 public constant OP_COMMAND_BINANCE_USER_PROFILE = bytes32("BINANCE_USER_PROFILE");
    bytes32 public constant OP_COMMAND_BINANCE_FUTURES_DETAILS = bytes32("BINANCE_FUTURES_DETAILS");
    bytes32 public constant OP_COMMAND_BITGET_ACCOUNT_SUMMARY = bytes32("BITGET_ACCOUNT_SUMMARY");

    uint256 public _extensionId;

    constructor(
        address _teeExtensionRegistry,
        address _teeMachineRegistry
    ) {
        teeExtensionRegistry = ITeeExtensionRegistry(_teeExtensionRegistry);
        teeMachineRegistry = ITeeMachineRegistry(_teeMachineRegistry);
    }

    /// @notice Discover and store this contract's extension ID.
    function setExtensionId() external {
        require(_extensionId == 0, "extension ID already set");

        uint256 count = teeExtensionRegistry.extensionsCounter();
        for (uint256 i = 1; i <= count; i++) {
            if (
                teeExtensionRegistry.getTeeExtensionInstructionsSender(i) ==
                address(this)
            ) {
                _extensionId = i;
                return;
            }
        }
        revert("extension ID not found");
    }

    /// @notice Update the stored private key by sending an encrypted key to the TEE.
    function updateKey(bytes calldata _encryptedKey) external payable returns (bytes32) {
        require(_extensionId != 0, "extension ID not set");

        address[] memory teeIds = teeMachineRegistry.getRandomTeeIds(_extensionId, 1);

        ITeeExtensionRegistry.TeeInstructionParams memory params;
        params.opType = OP_TYPE_KEY;
        params.opCommand = OP_COMMAND_UPDATE;
        params.message = _encryptedKey;

        return teeExtensionRegistry.sendInstructions{value: msg.value}(teeIds, params);
    }

    /// @notice Request the TEE to sign a message with the stored private key.
    function sign(bytes calldata _message) external payable returns (bytes32) {
        require(_extensionId != 0, "extension ID not set");

        address[] memory teeIds = teeMachineRegistry.getRandomTeeIds(_extensionId, 1);

        ITeeExtensionRegistry.TeeInstructionParams memory params;
        params.opType = OP_TYPE_KEY;
        params.opCommand = OP_COMMAND_SIGN;
        params.message = _message;

        return teeExtensionRegistry.sendInstructions{value: msg.value}(teeIds, params);
    }

    /// @notice Fetch Binance ticker data in the TEE and return a TEE-signed attestation payload.
    /// @dev _message should be JSON bytes like: {"symbol":"BTCUSDT"}
    function fetchBinanceAndAttest(bytes calldata _message) external payable returns (bytes32) {
        require(_extensionId != 0, "extension ID not set");

        address[] memory teeIds = teeMachineRegistry.getRandomTeeIds(_extensionId, 1);

        ITeeExtensionRegistry.TeeInstructionParams memory params;
        params.opType = OP_TYPE_MARKET;
        params.opCommand = OP_COMMAND_BINANCE_FETCH_AND_ATTEST;
        params.message = _message;

        return teeExtensionRegistry.sendInstructions{value: msg.value}(teeIds, params);
    }

    /// @notice Fetch authenticated Binance futures account PnL and return a TEE-signed attestation payload.
    function fetchBinanceAccountPnlAndAttest(bytes calldata _message) external payable returns (bytes32) {
        require(_extensionId != 0, "extension ID not set");

        address[] memory teeIds = teeMachineRegistry.getRandomTeeIds(_extensionId, 1);

        ITeeExtensionRegistry.TeeInstructionParams memory params;
        params.opType = OP_TYPE_MARKET;
        params.opCommand = OP_COMMAND_BINANCE_ACCOUNT_PNL;
        params.message = _message;

        return teeExtensionRegistry.sendInstructions{value: msg.value}(teeIds, params);
    }

    /// @notice Fetch public Binance 24h market stats and return a TEE-signed attestation payload.
    function fetchBinance24hStatsAndAttest(bytes calldata _message) external payable returns (bytes32) {
        require(_extensionId != 0, "extension ID not set");

        address[] memory teeIds = teeMachineRegistry.getRandomTeeIds(_extensionId, 1);

        ITeeExtensionRegistry.TeeInstructionParams memory params;
        params.opType = OP_TYPE_MARKET;
        params.opCommand = OP_COMMAND_BINANCE_24H_STATS;
        params.message = _message;

        return teeExtensionRegistry.sendInstructions{value: msg.value}(teeIds, params);
    }

    /// @notice Fetch authenticated Binance user profile (UID, account type, permissions, balances)
    ///         and return a TEE-signed payload. The result is ABI-encoded (payload, signature)
    ///         where signature is a 65-byte secp256k1 sig over keccak256(payload) by the TEE key.
    function fetchBinanceUserProfileAndAttest(bytes calldata _message) external payable returns (bytes32) {
        require(_extensionId != 0, "extension ID not set");

        address[] memory teeIds = teeMachineRegistry.getRandomTeeIds(_extensionId, 1);

        ITeeExtensionRegistry.TeeInstructionParams memory params;
        params.opType = OP_TYPE_MARKET;
        params.opCommand = OP_COMMAND_BINANCE_USER_PROFILE;
        params.message = _message;

        return teeExtensionRegistry.sendInstructions{value: msg.value}(teeIds, params);
    }

    /// @notice Fetch authenticated Binance spot account balances and return a TEE-signed account summary payload.
    function fetchBinanceAccountSummaryAndAttest(bytes calldata _message) external payable returns (bytes32) {
        require(_extensionId != 0, "extension ID not set");

        address[] memory teeIds = teeMachineRegistry.getRandomTeeIds(_extensionId, 1);

        ITeeExtensionRegistry.TeeInstructionParams memory params;
        params.opType = OP_TYPE_MARKET;
        params.opCommand = OP_COMMAND_BINANCE_ACCOUNT_SUMMARY;
        params.message = _message;

        return teeExtensionRegistry.sendInstructions{value: msg.value}(teeIds, params);
    }

    /// @notice Fetch detailed Binance futures position data (open and closed) and return a TEE-signed payload.
    function fetchBinanceFuturesDetailsAndAttest(bytes calldata _message) external payable returns (bytes32) {
        require(_extensionId != 0, "extension ID not set");

        address[] memory teeIds = teeMachineRegistry.getRandomTeeIds(_extensionId, 1);

        ITeeExtensionRegistry.TeeInstructionParams memory params;
        params.opType = OP_TYPE_MARKET;
        params.opCommand = OP_COMMAND_BINANCE_FUTURES_DETAILS;
        params.message = _message;

        return teeExtensionRegistry.sendInstructions{value: msg.value}(teeIds, params);
    }

    /// @notice Fetch authenticated Bitget spot account balances and return a TEE-signed account summary payload.
    function fetchBitgetAccountSummaryAndAttest(bytes calldata _message) external payable returns (bytes32) {
        require(_extensionId != 0, "extension ID not set");

        address[] memory teeIds = teeMachineRegistry.getRandomTeeIds(_extensionId, 1);

        ITeeExtensionRegistry.TeeInstructionParams memory params;
        params.opType = OP_TYPE_MARKET;
        params.opCommand = OP_COMMAND_BITGET_ACCOUNT_SUMMARY;
        params.message = _message;

        return teeExtensionRegistry.sendInstructions{value: msg.value}(teeIds, params);
    }
}

// SPDX-License-Identifier: MIT
pragma solidity >=0.7.6 <0.9;

/// @title BinanceAttestationStore
/// @notice Verifies a TEE-signed ABI-encoded user profile payload and publishes
///         it on-chain as a decoded, human-readable event.
/// @dev The TEE signs keccak256(abiPayload) with its secp256k1 key. The contract
///      recovers the signer address and emits individual typed fields alongside the
///      raw payload so consumers can (a) read data directly from the event and
///      (b) verify the attestation came from a trusted TEE via the Flare registry.
contract BinanceAttestationStore {
    struct AssetSummary {
        string asset;
        string total;
        string estimatedUsdt;
    }

    struct UserProfile {
        string   source;
        uint64   uid;
        string   accountType;
        string[] permissions;
        bool     canTrade;
        bool     canDeposit;
        bool     canWithdraw;
        string   estimatedTotalUsdt;
        uint256  unsupportedAssetCount;
        AssetSummary[] assets;
        uint256  fetchedAt;
        string   version;
    }

    event AttestationPublished(
        address  indexed teeAddress,
        uint64   indexed uid,
        string           accountType,
        string           estimatedTotalUsdt,
        uint256          fetchedAt,
        string           version,
        bytes            rawPayload,
        bytes            signature,
        uint256          timestamp
    );

    /// @notice Verify a TEE-signed ABI-encoded UserProfile payload and emit it
    ///         as a permanent, human-readable on-chain record.
    /// @param abiPayload  abi.encode(UserProfile) bytes produced by the TEE extension.
    /// @param signature   65-byte secp256k1 ECDSA signature (r || s || v) over keccak256(abiPayload).
    function publishAttestation(bytes calldata abiPayload, bytes calldata signature) external {
        address teeAddress = _recoverSigner(abiPayload, signature);
        UserProfile memory profile = abi.decode(abiPayload, (UserProfile));

        emit AttestationPublished(
            teeAddress,
            profile.uid,
            profile.accountType,
            profile.estimatedTotalUsdt,
            profile.fetchedAt,
            profile.version,
            abiPayload,
            signature,
            block.timestamp
        );
    }

    function _recoverSigner(bytes memory payload, bytes memory sig) internal pure returns (address) {
        require(sig.length == 65, "invalid signature length");
        bytes32 hash = keccak256(payload);
        bytes32 r;
        bytes32 s;
        uint8 v;
        // solhint-disable-next-line no-inline-assembly
        assembly {
            r := mload(add(sig, 32))
            s := mload(add(sig, 64))
            v := byte(0, mload(add(sig, 96)))
        }
        if (v < 27) v += 27;
        require(v == 27 || v == 28, "invalid v value");
        address recovered = ecrecover(hash, v, r, s);
        require(recovered != address(0), "ecrecover failed");
        return recovered;
    }
}

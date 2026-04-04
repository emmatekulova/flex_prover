// SPDX-License-Identifier: MIT
pragma solidity >=0.7.6 <0.9;

/// @title BinanceAttestationStore
/// @notice Verifies a TEE-signed Binance payload and publishes it on-chain as an event.
/// @dev The TEE signs keccak256(payload) with its secp256k1 key. This contract recovers
///      the signer address and emits it alongside the payload so consumers can verify
///      the attestation came from a trusted TEE by checking the recovered address against
///      the Flare TEE machine registry.
contract BinanceAttestationStore {
    event AttestationPublished(
        address indexed teeAddress,
        bytes payload,
        bytes signature,
        uint256 timestamp
    );

    /// @notice Verify a TEE-signed payload and emit it as a permanent on-chain record.
    /// @param payload  Raw JSON bytes of the attestation (e.g. BinanceUserProfileAttestationPayload).
    /// @param signature 65-byte secp256k1 ECDSA signature (r || s || v) over keccak256(payload).
    /// @dev Anyone can call this. The recovered teeAddress in the event is what consumers
    ///      should verify against their trusted TEE list.
    function publishAttestation(bytes calldata payload, bytes calldata signature) external {
        address teeAddress = _recoverSigner(payload, signature);
        emit AttestationPublished(teeAddress, payload, signature, block.timestamp);
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

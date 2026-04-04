// SPDX-License-Identifier: MIT
pragma solidity >=0.7.6 <0.9;

/// @title BinanceAttestationStore
/// @notice Stores TEE-signed Binance portfolio growth attestations on-chain.
/// @dev The TEE produces a JSON payload and signs keccak256(payload) with its
///      secp256k1 key. This contract recovers the signer and emits the raw payload
///      as a permanent on-chain record. Payload format is UTF-8 JSON:
///      {"source":"binance-profile-growth","wallet":"0x...","windowDays":7,
///       "startSnapshot":{...},"endSnapshot":{...},"growthPercent":"9.00",...}
contract BinanceAttestationStore {
    event AttestationPublished(
        address indexed teeAddress,
        bytes           payload,
        bytes           signature,
        uint256         timestamp
    );

    /// @notice Verify a TEE-signed payload and emit it as a permanent on-chain record.
    /// @param payload    Raw JSON bytes produced by the TEE extension.
    /// @param signature  65-byte secp256k1 ECDSA signature (r || s || v) over keccak256(payload).
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

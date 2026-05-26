package crypto

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

func TestBase58Encoding(t *testing.T) {
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte{0x00}, "1"},
		{[]byte("hello"), "Cn8eVZg"},
		{[]byte("world"), "EUYUqQf"},
	}

	for _, tc := range tests {
		actual := Base58Encode(tc.input)
		if actual != tc.expected {
			t.Errorf("Base58Encode(%v) = '%s'; expected '%s'", tc.input, actual, tc.expected)
		}
	}
}

func TestTransactionSerialization(t *testing.T) {
	// A simple transaction JSON representing a vote operation
	txJSON := `{
		"ref_block_num": 1234,
		"ref_block_prefix": 56789,
		"expiration": "2026-05-25T10:00:00",
		"operations": [
			["vote", {
				"voter": "alice",
				"author": "bob",
				"permlink": "hello-world",
				"weight": 10000
			}]
		],
		"extensions": [],
		"signatures": []
	}`

	var tx Transaction
	if err := json.Unmarshal([]byte(txJSON), &tx); err != nil {
		t.Fatalf("failed to unmarshal test transaction: %v", err)
	}

	bytes, err := tx.Serialize()
	if err != nil {
		t.Fatalf("failed to serialize transaction: %v", err)
	}

	if len(bytes) == 0 {
		t.Error("expected non-empty serialized bytes")
	}
}

func TestSignatureRecovery(t *testing.T) {
	// 1. Generate a random private key
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	// 2. Create a dummy digest
	digestHex := "2b15740a6b7d57fd5a542459a94f3a8153c68c4a92c4a0e7c509b9bd2b642705"
	digest, _ := hex.DecodeString(digestHex)

	// 3. Sign the digest (compact signature)
	sig := ecdsa.SignCompact(privKey, digest, true)

	// ecdsa.SignCompact returns the signature with recovery byte in [27-34] range
	// For testing, we ensure it's in hex format as our RecoverPublicKey expects.
	sigHex := hex.EncodeToString(sig)

	// 4. Recover the public key using our function
	pubKeyStr, err := RecoverPublicKey(sigHex, digest)
	if err != nil {
		t.Fatalf("failed to recover public key: %v", err)
	}

	if len(pubKeyStr) < 3 || pubKeyStr[:3] != "STM" {
		t.Errorf("expected recovered public key to start with 'STM', got '%s'", pubKeyStr)
	}
}

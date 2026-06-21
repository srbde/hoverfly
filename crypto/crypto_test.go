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

func TestTransferSerializationAcceptsAppbaseAsset(t *testing.T) {
	legacy := `{
		"ref_block_num": 1234,
		"ref_block_prefix": 56789,
		"expiration": "2026-05-25T10:00:00",
		"operations": [["transfer", {
			"from": "bob",
			"to": "cambium.vault",
			"amount": "1.000 HIVE",
			"memo": "@bob"
		}]],
		"extensions": [],
		"signatures": []
	}`
	appbase := `{
		"ref_block_num": 1234,
		"ref_block_prefix": 56789,
		"expiration": "2026-05-25T10:00:00",
		"operations": [["transfer", {
			"from": "bob",
			"to": "cambium.vault",
			"amount": {"amount": "1000", "precision": 3, "nai": "@@000000021"},
			"memo": "@bob"
		}]],
		"extensions": [],
		"signatures": []
	}`

	serialize := func(input string) []byte {
		t.Helper()
		var tx Transaction
		if err := json.Unmarshal([]byte(input), &tx); err != nil {
			t.Fatalf("failed to unmarshal transaction: %v", err)
		}
		serialized, err := tx.Serialize()
		if err != nil {
			t.Fatalf("failed to serialize transaction: %v", err)
		}
		return serialized
	}

	legacyBytes := serialize(legacy)
	appbaseBytes := serialize(appbase)
	if hex.EncodeToString(appbaseBytes) != hex.EncodeToString(legacyBytes) {
		t.Fatalf("appbase and legacy assets serialized differently")
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

func TestVerifySignaturesUsesHiveChainID(t *testing.T) {
	txJSON := `{
		"ref_block_num": 36312,
		"ref_block_prefix": 3608149636,
		"expiration": "2026-06-21T15:47:08",
		"operations": [{
			"type": "custom_json_operation",
			"value": {
				"required_auths": [],
				"required_posting_auths": ["bob"],
				"id": "cambium_register",
				"json": "{\"l2_public_key\":\"444fdbe02002855c51b39c3bf7abe72bbf19207aa4de87eb5abdfe82dbc1917d\"}"
			}
		}],
		"extensions": [],
		"signatures": [
			"204c19e3b25c3cc460fc25d806d10aa94dc05e200079f779ef0ece5bc8b855a9de5371499a23deb93e4235d1946030107080337c3130ded4f8765bcfa7e7b9a180"
		]
	}`

	var tx Transaction
	if err := json.Unmarshal([]byte(txJSON), &tx); err != nil {
		t.Fatalf("failed to unmarshal custom_json transaction: %v", err)
	}

	recoveredKeys, err := VerifySignatures(&tx, HiveChainID)
	if err != nil {
		t.Fatalf("failed to verify transaction signature: %v", err)
	}

	const expectedPostingKey = "STM7UpcJ97QRgsXkKVmx8QZZJVHihsJBRmDW57QbYWxMg7m5AVcrB"
	if len(recoveredKeys) != 1 || recoveredKeys[0] != expectedPostingKey {
		t.Fatalf("expected recovered posting key %s, got %v", expectedPostingKey, recoveredKeys)
	}
}

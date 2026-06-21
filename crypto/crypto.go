package crypto

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/ripemd160"
)

const (
	base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	HiveChainID    = "beeab0de00000000000000000000000000000000000000000000000000000000"
)

// Base58Encode encodes a byte slice into a Base58 string.
func Base58Encode(input []byte) string {
	x := new(big.Int).SetBytes(input)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)
	var result []byte
	for x.Cmp(zero) > 0 {
		x.DivMod(x, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}
	for _, b := range input {
		if b == 0x00 {
			result = append(result, base58Alphabet[0])
		} else {
			break
		}
	}
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

type Transaction struct {
	RefBlockNum    uint16            `json:"ref_block_num"`
	RefBlockPrefix uint32            `json:"ref_block_prefix"`
	Expiration     string            `json:"expiration"`
	Operations     []json.RawMessage `json:"operations"`
	Extensions     []any             `json:"extensions"`
	Signatures     []string          `json:"signatures"`
}

func (tx *Transaction) Serialize() ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.LittleEndian, tx.RefBlockNum); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.LittleEndian, tx.RefBlockPrefix); err != nil {
		return nil, err
	}

	t, err := time.Parse("2006-01-02T15:04:05", tx.Expiration)
	if err != nil {
		return nil, fmt.Errorf("invalid expiration time format: %w", err)
	}
	expUnix := uint32(t.Unix())
	if err := binary.Write(&buf, binary.LittleEndian, expUnix); err != nil {
		return nil, err
	}

	writeVarint(&buf, uint64(len(tx.Operations)))
	for _, rawOp := range tx.Operations {
		var tuple []json.RawMessage
		var opName string
		var opBody json.RawMessage
		var parsed bool

		// 1. Try tuple style: [name, body]
		if err := json.Unmarshal(rawOp, &tuple); err == nil && len(tuple) == 2 {
			if err := json.Unmarshal(tuple[0], &opName); err == nil {
				opBody = tuple[1]
				parsed = true
			}
		}

		// 2. Try object style: {"type": "...", "value": ...}
		if !parsed {
			var obj struct {
				Type  string          `json:"type"`
				Value json.RawMessage `json:"value"`
			}
			if err := json.Unmarshal(rawOp, &obj); err == nil && obj.Type != "" {
				opName = obj.Type
				opBody = obj.Value
				parsed = true
			}
		}

		if !parsed {
			return nil, errors.New("invalid operation format (expected [name, body] or {type, value})")
		}

		// Strip _operation suffix if present, e.g. "pow_operation" -> "pow"
		opName = strings.TrimSuffix(opName, "_operation")

		opBytes, err := serializeOperation(opName, opBody)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize operation %s: %w", opName, err)
		}
		buf.Write(opBytes)
	}

	buf.WriteByte(0)

	return buf.Bytes(), nil
}

func writeVarint(buf *bytes.Buffer, val uint64) {
	varintBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(varintBuf, val)
	buf.Write(varintBuf[:n])
}

func serializeString(buf *bytes.Buffer, s string) {
	writeVarint(buf, uint64(len(s)))
	buf.WriteString(s)
}

func serializeAsset(buf *bytes.Buffer, assetStr string) error {
	parts := strings.Fields(assetStr)
	if len(parts) != 2 {
		return fmt.Errorf("invalid asset format: %s", assetStr)
	}
	val, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("invalid asset value: %w", err)
	}
	symbol := parts[1]

	precision := uint8(3)
	switch symbol {
	case "VESTS":
		precision = 6
	}

	satoshis := int64(math.Round(val * math.Pow(10, float64(precision))))

	wireSymbol := symbol
	switch symbol {
	case "HIVE":
		wireSymbol = "STEEM"
	case "HBD":
		wireSymbol = "SBD"
	}

	if len(wireSymbol) > 7 {
		return fmt.Errorf("asset symbol exceeds 7 characters: %s", wireSymbol)
	}

	if err := binary.Write(buf, binary.LittleEndian, satoshis); err != nil {
		return err
	}

	buf.WriteByte(precision)

	symbolBytes := []byte(wireSymbol)
	buf.Write(symbolBytes)
	for i := len(symbolBytes); i < 7; i++ {
		buf.WriteByte(0)
	}

	return nil
}

func serializeOperation(opName string, bodyJSON json.RawMessage) ([]byte, error) {
	var buf bytes.Buffer

	switch opName {
	case "vote":
		writeVarint(&buf, 0)
		var op struct {
			Voter    string `json:"voter"`
			Author   string `json:"author"`
			Permlink string `json:"permlink"`
			Weight   int16  `json:"weight"`
		}
		if err := json.Unmarshal(bodyJSON, &op); err != nil {
			return nil, err
		}
		serializeString(&buf, op.Voter)
		serializeString(&buf, op.Author)
		serializeString(&buf, op.Permlink)
		if err := binary.Write(&buf, binary.LittleEndian, op.Weight); err != nil {
			return nil, err
		}

	case "comment":
		writeVarint(&buf, 1)
		var op struct {
			ParentAuthor   string `json:"parent_author"`
			ParentPermlink string `json:"parent_permlink"`
			Author         string `json:"author"`
			Permlink       string `json:"permlink"`
			Title          string `json:"title"`
			Body           string `json:"body"`
			JSONMetadata   string `json:"json_metadata"`
		}
		if err := json.Unmarshal(bodyJSON, &op); err != nil {
			return nil, err
		}
		serializeString(&buf, op.ParentAuthor)
		serializeString(&buf, op.ParentPermlink)
		serializeString(&buf, op.Author)
		serializeString(&buf, op.Permlink)
		serializeString(&buf, op.Title)
		serializeString(&buf, op.Body)
		serializeString(&buf, op.JSONMetadata)

	case "transfer":
		writeVarint(&buf, 2)
		var op struct {
			From   string `json:"from"`
			To     string `json:"to"`
			Amount string `json:"amount"`
			Memo   string `json:"memo"`
		}
		if err := json.Unmarshal(bodyJSON, &op); err != nil {
			return nil, err
		}
		serializeString(&buf, op.From)
		serializeString(&buf, op.To)
		if err := serializeAsset(&buf, op.Amount); err != nil {
			return nil, err
		}
		serializeString(&buf, op.Memo)

	case "transfer_to_savings":
		writeVarint(&buf, 32)
		var op struct {
			From   string `json:"from"`
			To     string `json:"to"`
			Amount string `json:"amount"`
			Memo   string `json:"memo"`
		}
		if err := json.Unmarshal(bodyJSON, &op); err != nil {
			return nil, err
		}
		serializeString(&buf, op.From)
		serializeString(&buf, op.To)
		if err := serializeAsset(&buf, op.Amount); err != nil {
			return nil, err
		}
		serializeString(&buf, op.Memo)

	case "custom_json":
		writeVarint(&buf, 18)
		var op struct {
			RequiredAuths        []string `json:"required_auths"`
			RequiredPostingAuths []string `json:"required_posting_auths"`
			ID                   string   `json:"id"`
			JSON                 string   `json:"json"`
		}
		if err := json.Unmarshal(bodyJSON, &op); err != nil {
			return nil, err
		}
		writeVarint(&buf, uint64(len(op.RequiredAuths)))
		for _, auth := range op.RequiredAuths {
			serializeString(&buf, auth)
		}
		writeVarint(&buf, uint64(len(op.RequiredPostingAuths)))
		for _, auth := range op.RequiredPostingAuths {
			serializeString(&buf, auth)
		}
		serializeString(&buf, op.ID)
		serializeString(&buf, op.JSON)

	case "pow":
		writeVarint(&buf, 14)
		// Dummy serialization for deprecated PoW operation to allow testing/hex conversion
		buf.Write(make([]byte, 32))

	default:
		return nil, fmt.Errorf("operation '%s' is not supported by hoverfly signature validation", opName)
	}

	return buf.Bytes(), nil
}

func RecoverPublicKey(sigHex string, digest []byte) (string, error) {
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return "", err
	}
	if len(sigBytes) != 65 {
		return "", fmt.Errorf("invalid signature length: expected 65, got %d", len(sigBytes))
	}

	recByte := sigBytes[0]
	if recByte < 27 || recByte > 34 {
		return "", fmt.Errorf("invalid recovery byte: %d", recByte)
	}

	pub, _, err := ecdsa.RecoverCompact(sigBytes, digest)
	if err != nil {
		return "", err
	}

	pubBytes := pub.SerializeCompressed()

	hasher := ripemd160.New()
	hasher.Write(pubBytes)
	checksum := hasher.Sum(nil)[:4]

	payload := append(pubBytes, checksum...)
	return "STM" + Base58Encode(payload), nil
}

func VerifySignatures(tx *Transaction, chainID string) ([]string, error) {
	if len(tx.Signatures) == 0 {
		return nil, errors.New("transaction has no signatures")
	}

	txBytes, err := tx.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	chainBytes, err := hex.DecodeString(chainID)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID hex: %w", err)
	}

	hasher := sha256.New()
	hasher.Write(chainBytes)
	hasher.Write(txBytes)
	digest := hasher.Sum(nil)

	var recoveredKeys []string
	for _, sigHex := range tx.Signatures {
		pubKey, err := RecoverPublicKey(sigHex, digest)
		if err != nil {
			return nil, fmt.Errorf("failed to recover public key from signature: %w", err)
		}
		recoveredKeys = append(recoveredKeys, pubKey)
	}

	return recoveredKeys, nil
}

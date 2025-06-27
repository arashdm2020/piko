package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/piko/piko/models"
)

// calculateBlockHash calculates the hash of a block
func calculateBlockHash(previousHash interface{}, timestamp time.Time, merkleRoot string, nonce int64) string {
	var prevHash string
	if previousHash == nil {
		prevHash = "0000000000000000000000000000000000000000000000000000000000000000"
	} else {
		prevHash = previousHash.(string)
	}

	data := fmt.Sprintf("%s%d%s%d", prevHash, timestamp.UnixNano(), merkleRoot, nonce)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// calculateTransactionHash calculates the hash of a transaction
func calculateTransactionHash(txType models.TransactionType, dataID string, blockID string) string {
	data := fmt.Sprintf("%s%s%s%d", txType, dataID, blockID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// calculateMerkleRoot calculates the merkle root of transactions
func calculateMerkleRoot(transactions []*MempoolTransaction) string {
	if len(transactions) == 0 {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}

	// Calculate hashes of transactions
	hashes := make([]string, len(transactions))
	for i, tx := range transactions {
		data := fmt.Sprintf("%s%s%d", tx.Type, tx.DataID, tx.Timestamp.UnixNano())
		hash := sha256.Sum256([]byte(data))
		hashes[i] = hex.EncodeToString(hash[:])
	}

	// Calculate merkle root
	for len(hashes) > 1 {
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		var newHashes []string
		for i := 0; i < len(hashes); i += 2 {
			data := hashes[i] + hashes[i+1]
			hash := sha256.Sum256([]byte(data))
			newHashes = append(newHashes, hex.EncodeToString(hash[:]))
		}
		hashes = newHashes
	}

	return hashes[0]
}

// calculateNonce calculates a nonce for a block (simplified proof of work)
func calculateNonce(previousHash string, timestamp time.Time, merkleRoot string) int64 {
	var nonce int64
	for {
		hash := calculateBlockHash(previousHash, timestamp, merkleRoot, nonce)
		// Simplified proof of work: hash must end with "00"
		if hash[len(hash)-2:] == "00" {
			return nonce
		}
		nonce++
	}
} 
package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/piko/piko/database"
)

var (
	// ErrBlockNotFound is returned when a block is not found
	ErrBlockNotFound = errors.New("block not found")
	// ErrTransactionNotFound is returned when a transaction is not found
	ErrTransactionNotFound = errors.New("transaction not found")
)

// TransactionType represents the type of a transaction
type TransactionType string

const (
	// TransactionTypeMessage indicates a direct message transaction
	TransactionTypeMessage TransactionType = "message"
	// TransactionTypeChannelMessage indicates a channel message transaction
	TransactionTypeChannelMessage TransactionType = "channel_message"
	// TransactionTypeChannelCreate indicates a channel creation transaction
	TransactionTypeChannelCreate TransactionType = "channel_create"
	// TransactionTypeChannelJoin indicates a channel join transaction
	TransactionTypeChannelJoin TransactionType = "channel_join"
)

// Block represents a block in the blockchain
type Block struct {
	ID           string    `json:"id"`
	PreviousHash *string   `json:"previous_hash,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	MerkleRoot   string    `json:"merkle_root"`
	Nonce        int64     `json:"nonce"`
	Height       int       `json:"height"`
	Transactions []*Transaction `json:"transactions,omitempty"`
}

// Transaction represents a transaction in the blockchain
type Transaction struct {
	Hash      string         `json:"hash"`
	BlockID   string         `json:"block_id"`
	Type      TransactionType `json:"type"`
	DataID    string         `json:"data_id"`
	Timestamp time.Time      `json:"timestamp"`
}

// CreateBlock creates a new block in the database
func CreateBlock(block *Block) error {
	_, err := database.DB.Exec(
		"INSERT INTO blocks (id, previous_hash, merkle_root, nonce, height) VALUES (?, ?, ?, ?, ?)",
		block.ID, block.PreviousHash, block.MerkleRoot, block.Nonce, block.Height,
	)
	return err
}

// GetBlockByID retrieves a block by its ID
func GetBlockByID(id string) (*Block, error) {
	block := &Block{}
	err := database.DB.QueryRow(
		"SELECT id, previous_hash, timestamp, merkle_root, nonce, height FROM blocks WHERE id = ?",
		id,
	).Scan(
		&block.ID, &block.PreviousHash, &block.Timestamp, &block.MerkleRoot, &block.Nonce, &block.Height,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}

	// Get transactions for this block
	transactions, err := GetTransactionsByBlockID(id)
	if err != nil {
		return nil, err
	}
	block.Transactions = transactions

	return block, nil
}

// GetBlockByHeight retrieves a block by its height
func GetBlockByHeight(height int) (*Block, error) {
	block := &Block{}
	err := database.DB.QueryRow(
		"SELECT id, previous_hash, timestamp, merkle_root, nonce, height FROM blocks WHERE height = ?",
		height,
	).Scan(
		&block.ID, &block.PreviousHash, &block.Timestamp, &block.MerkleRoot, &block.Nonce, &block.Height,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}

	// Get transactions for this block
	transactions, err := GetTransactionsByBlockID(block.ID)
	if err != nil {
		return nil, err
	}
	block.Transactions = transactions

	return block, nil
}

// GetLatestBlock retrieves the latest block in the blockchain
func GetLatestBlock() (*Block, error) {
	block := &Block{}
	err := database.DB.QueryRow(
		"SELECT id, previous_hash, timestamp, merkle_root, nonce, height FROM blocks ORDER BY height DESC LIMIT 1",
	).Scan(
		&block.ID, &block.PreviousHash, &block.Timestamp, &block.MerkleRoot, &block.Nonce, &block.Height,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}

	// Get transactions for this block
	transactions, err := GetTransactionsByBlockID(block.ID)
	if err != nil {
		return nil, err
	}
	block.Transactions = transactions

	return block, nil
}

// CreateTransaction creates a new transaction in the database
func CreateTransaction(transaction *Transaction) error {
	_, err := database.DB.Exec(
		"INSERT INTO transactions (hash, block_id, type, data_id) VALUES (?, ?, ?, ?)",
		transaction.Hash, transaction.BlockID, transaction.Type, transaction.DataID,
	)
	return err
}

// GetTransactionByHash retrieves a transaction by its hash
func GetTransactionByHash(hash string) (*Transaction, error) {
	transaction := &Transaction{}
	var txType string
	err := database.DB.QueryRow(
		"SELECT hash, block_id, type, data_id, timestamp FROM transactions WHERE hash = ?",
		hash,
	).Scan(
		&transaction.Hash, &transaction.BlockID, &txType, &transaction.DataID, &transaction.Timestamp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	transaction.Type = TransactionType(txType)
	return transaction, nil
}

// GetTransactionsByBlockID retrieves all transactions for a block
func GetTransactionsByBlockID(blockID string) ([]*Transaction, error) {
	rows, err := database.DB.Query(
		"SELECT hash, block_id, type, data_id, timestamp FROM transactions WHERE block_id = ? ORDER BY timestamp",
		blockID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []*Transaction{}
	for rows.Next() {
		transaction := &Transaction{}
		var txType string
		err := rows.Scan(
			&transaction.Hash, &transaction.BlockID, &txType, &transaction.DataID, &transaction.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		transaction.Type = TransactionType(txType)
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransactionsByAddress retrieves all transactions related to an address
func GetTransactionsByAddress(address string) ([]*Transaction, error) {
	// This query joins the transactions table with messages and channel_messages
	// to find all transactions related to the given address
	rows, err := database.DB.Query(`
		SELECT t.hash, t.block_id, t.type, t.data_id, t.timestamp 
		FROM transactions t
		LEFT JOIN messages m ON t.data_id = m.id AND t.type = 'message'
		LEFT JOIN channel_messages cm ON t.data_id = cm.id AND t.type = 'channel_message'
		LEFT JOIN channels c ON t.data_id = c.id AND t.type = 'channel_create'
		LEFT JOIN channel_members cmem ON t.data_id = CONCAT(cmem.channel_id, ':', cmem.user_address) AND t.type = 'channel_join'
		WHERE m.sender_address = ? OR m.recipient_address = ? OR cm.sender_address = ? OR c.admin_address = ? OR cmem.user_address = ?
		ORDER BY t.timestamp DESC
	`, address, address, address, address, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []*Transaction{}
	for rows.Next() {
		transaction := &Transaction{}
		var txType string
		err := rows.Scan(
			&transaction.Hash, &transaction.BlockID, &txType, &transaction.DataID, &transaction.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		transaction.Type = TransactionType(txType)
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetBlockchainStats retrieves statistics about the blockchain
func GetBlockchainStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total number of blocks
	var blockCount int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM blocks").Scan(&blockCount)
	if err != nil {
		return nil, err
	}
	stats["block_count"] = blockCount

	// Get total number of transactions
	var txCount int
	err = database.DB.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&txCount)
	if err != nil {
		return nil, err
	}
	stats["transaction_count"] = txCount

	// Get transaction counts by type
	rows, err := database.DB.Query("SELECT type, COUNT(*) FROM transactions GROUP BY type")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txTypes := make(map[string]int)
	for rows.Next() {
		var txType string
		var count int
		err := rows.Scan(&txType, &count)
		if err != nil {
			return nil, err
		}
		txTypes[txType] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	stats["transaction_types"] = txTypes

	// Get latest block timestamp
	var latestTimestamp time.Time
	err = database.DB.QueryRow("SELECT timestamp FROM blocks ORDER BY height DESC LIMIT 1").Scan(&latestTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err != sql.ErrNoRows {
		stats["latest_block_time"] = latestTimestamp
	}

	return stats, nil
} 
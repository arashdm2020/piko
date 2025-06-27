package blockchain

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/piko/piko/models"
)

// Initialize initializes the blockchain
func (bc *Blockchain) Initialize() error {
	// Get the latest block from the database
	latestBlock, err := models.GetLatestBlock()
	if err != nil {
		if errors.Is(err, models.ErrBlockNotFound) {
			// Create genesis block
			if err := bc.createGenesisBlock(); err != nil {
				return fmt.Errorf("failed to create genesis block: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get latest block: %w", err)
		}
	} else {
		bc.mu.Lock()
		bc.LatestBlock = latestBlock
		bc.mu.Unlock()
	}

	// Start block creation goroutine
	go bc.startBlockCreation()

	return nil
}

// createGenesisBlock creates the genesis block
func (bc *Blockchain) createGenesisBlock() error {
	log.Println("Creating genesis block...")

	// Create genesis block
	genesisBlock := &models.Block{
		ID:         calculateBlockHash(nil, time.Now(), "genesis", 0),
		Timestamp:  time.Now(),
		MerkleRoot: "genesis",
		Nonce:      0,
		Height:     0,
	}

	// Save genesis block to database
	if err := models.CreateBlock(genesisBlock); err != nil {
		return err
	}

	// Set as latest block
	bc.mu.Lock()
	bc.LatestBlock = genesisBlock
	bc.mu.Unlock()

	log.Println("Genesis block created")
	return nil
}

// startBlockCreation starts the block creation process
func (bc *Blockchain) startBlockCreation() {
	ticker := time.NewTicker(bc.Config.BlockTime)
	defer ticker.Stop()

	for range ticker.C {
		if err := bc.createBlock(); err != nil {
			if errors.Is(err, ErrEmptyMempool) {
				// Skip block creation if mempool is empty
				continue
			}
			log.Printf("Failed to create block: %v", err)
		}
	}
}

// createBlock creates a new block
func (bc *Blockchain) createBlock() error {
	// Get transactions from mempool
	transactions, err := bc.Mempool.GetTransactions()
	if err != nil {
		return err
	}

	if len(transactions) == 0 {
		return ErrEmptyMempool
	}

	// Get latest block
	bc.mu.RLock()
	latestBlock := bc.LatestBlock
	bc.mu.RUnlock()

	// Calculate new block height
	height := latestBlock.Height + 1

	// Calculate merkle root
	merkleRoot := calculateMerkleRoot(transactions)

	// Create new block
	timestamp := time.Now()
	nonce := calculateNonce(latestBlock.ID, timestamp, merkleRoot)
	blockID := calculateBlockHash(latestBlock.ID, timestamp, merkleRoot, nonce)

	// Create block
	block := &models.Block{
		ID:           blockID,
		PreviousHash: &latestBlock.ID,
		Timestamp:    timestamp,
		MerkleRoot:   merkleRoot,
		Nonce:        nonce,
		Height:       height,
	}

	// Save block to database
	if err := models.CreateBlock(block); err != nil {
		return err
	}

	// Create transactions in database
	for _, tx := range transactions {
		transaction := &models.Transaction{
			Hash:    calculateTransactionHash(tx.Type, tx.DataID, blockID),
			BlockID: blockID,
			Type:    tx.Type,
			DataID:  tx.DataID,
		}
		if err := models.CreateTransaction(transaction); err != nil {
			log.Printf("Failed to create transaction: %v", err)
			continue
		}

		// Update message or channel message with block ID
		switch tx.Type {
		case models.TransactionTypeMessage:
			if err := models.UpdateMessageBlockID(tx.DataID, blockID); err != nil {
				log.Printf("Failed to update message block ID: %v", err)
			}
		case models.TransactionTypeChannelMessage:
			if err := models.UpdateChannelMessageBlockID(tx.DataID, blockID); err != nil {
				log.Printf("Failed to update channel message block ID: %v", err)
			}
		}
	}

	// Set as latest block
	bc.mu.Lock()
	bc.LatestBlock = block
	bc.mu.Unlock()

	// Clear mempool
	bc.Mempool.Clear()

	log.Printf("Block created: %s (height: %d, transactions: %d)", blockID, height, len(transactions))
	return nil
}

// AddToMempool adds a transaction to the mempool
func (bc *Blockchain) AddToMempool(txType models.TransactionType, dataID string) error {
	return bc.Mempool.AddTransaction(&MempoolTransaction{
		Type:      txType,
		DataID:    dataID,
		Timestamp: time.Now(),
	})
}

// GetTransactions gets transactions from the mempool
func (m *Mempool) GetTransactions() ([]*MempoolTransaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Transactions) == 0 {
		return nil, ErrEmptyMempool
	}

	// Return a copy of the transactions
	transactions := make([]*MempoolTransaction, len(m.Transactions))
	copy(transactions, m.Transactions)
	return transactions, nil
}

// AddTransaction adds a transaction to the mempool
func (m *Mempool) AddTransaction(tx *MempoolTransaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if mempool is full
	if len(m.Transactions) >= m.Capacity {
		// Remove oldest transaction
		m.Transactions = m.Transactions[1:]
	}

	// Add transaction
	m.Transactions = append(m.Transactions, tx)
	return nil
}

// Clear clears the mempool
func (m *Mempool) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Transactions = make([]*MempoolTransaction, 0)
} 
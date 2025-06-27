package blockchain

import (
	"errors"
	"sync"
	"time"

	"github.com/piko/piko/config"
	"github.com/piko/piko/models"
)

var (
	// ErrEmptyMempool is returned when the mempool is empty
	ErrEmptyMempool = errors.New("mempool is empty")
)

// Blockchain represents the blockchain
type Blockchain struct {
	Config      *config.BlockchainConfig
	Mempool     *Mempool
	LatestBlock *models.Block
	mu          sync.RWMutex
}

// Mempool represents the mempool (pending transactions)
type Mempool struct {
	Transactions []*MempoolTransaction
	Capacity     int
	mu           sync.RWMutex
}

// MempoolTransaction represents a transaction in the mempool
type MempoolTransaction struct {
	Type      models.TransactionType
	DataID    string
	Timestamp time.Time
}

// NewBlockchain creates a new blockchain
func NewBlockchain(cfg *config.BlockchainConfig) *Blockchain {
	return &Blockchain{
		Config: cfg,
		Mempool: &Mempool{
			Transactions: make([]*MempoolTransaction, 0),
			Capacity:     cfg.MempoolCapacity,
		},
	}
}

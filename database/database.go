package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/piko/piko/config"
)

var (
	// DB is the global database connection
	DB *sql.DB

	// ErrNotInitialized is returned when the database is not initialized
	ErrNotInitialized = errors.New("database not initialized")
)

// Initialize initializes the database connection
func Initialize(cfg config.DatabaseConfig) error {
	var err error

	// Connect to the database
	DB, err = sql.Open(cfg.Driver, cfg.ConnectionString)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(cfg.MaxOpenConns)
	DB.SetMaxIdleConns(cfg.MaxIdleConns)
	DB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Verify the connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize database schema
	if err := initSchema(); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return ErrNotInitialized
	}
	return DB.Close()
}

// dropTables drops all tables if they exist
func dropTables() error {
	if DB == nil {
		return ErrNotInitialized
	}

	// Drop tables in reverse order of dependencies
	tables := []string{
		"transactions",
		"blocks",
		"channel_messages",
		"channel_members",
		"channels",
		"messages",
		"users",
		"secret_chat_messages",
		"secret_chat_participants",
		"secret_chats",
	}

	for _, table := range tables {
		_, err := DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// initSchema initializes the database schema
func initSchema() error {
	if DB == nil {
		return ErrNotInitialized
	}

	// Set file per table to ON
	_, err := DB.Exec("SET GLOBAL innodb_file_per_table=ON")
	if err != nil {
		fmt.Printf("Warning: Could not set innodb_file_per_table: %v\n", err)
	}

	// Drop all existing tables to ensure clean schema
	if err = dropTables(); err != nil {
		return fmt.Errorf("failed to drop existing tables: %w", err)
	}

	// Create users table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(255) UNIQUE,
			phone VARCHAR(20) UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			public_key BLOB NOT NULL,
			address VARCHAR(46) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create messages table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id VARCHAR(64) PRIMARY KEY,
			sender_address VARCHAR(46) NOT NULL,
			recipient_address VARCHAR(46) NOT NULL,
			encrypted_content BLOB NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			status ENUM('pending', 'delivered', 'read') DEFAULT 'pending',
			expiration_time TIMESTAMP NULL,
			block_id VARCHAR(64) NULL,
			INDEX (sender_address(32)),
			INDEX (recipient_address(32)),
			INDEX (block_id(32))
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create channels table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS channels (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			admin_address VARCHAR(46) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX (admin_address(32))
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create channel_members table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS channel_members (
			channel_id VARCHAR(64) NOT NULL,
			user_address VARCHAR(46) NOT NULL,
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (channel_id(32), user_address(32)),
			INDEX (user_address(32))
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create channel_messages table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS channel_messages (
			id VARCHAR(64) PRIMARY KEY,
			channel_id VARCHAR(64) NOT NULL,
			sender_address VARCHAR(46) NOT NULL,
			encrypted_content BLOB NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			block_id VARCHAR(64) NULL,
			INDEX (channel_id(32)),
			INDEX (sender_address(32)),
			INDEX (block_id(32))
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create blocks table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS blocks (
			id VARCHAR(64) PRIMARY KEY,
			previous_hash VARCHAR(64) NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			merkle_root VARCHAR(64) NOT NULL,
			nonce BIGINT NOT NULL,
			height INT NOT NULL,
			INDEX (height)
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create transactions table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			hash VARCHAR(64) PRIMARY KEY,
			block_id VARCHAR(64) NOT NULL,
			type ENUM('message', 'channel_message', 'channel_create', 'channel_join') NOT NULL,
			data_id VARCHAR(64) NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX (block_id(32)),
			INDEX (data_id(32))
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create secret_chats table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS secret_chats (
			channel_id VARCHAR(12) PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			INDEX (expires_at)
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create secret_chat_participants table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS secret_chat_participants (
			session_id VARCHAR(32) PRIMARY KEY,
			channel_id VARCHAR(12) NOT NULL,
			display_name VARCHAR(64) NOT NULL,
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_active_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX (channel_id),
			INDEX (last_active_at)
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create secret_chat_messages table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS secret_chat_messages (
			id VARCHAR(64) PRIMARY KEY,
			channel_id VARCHAR(12) NOT NULL,
			session_id VARCHAR(32) NOT NULL,
			display_name VARCHAR(64) NOT NULL,
			encrypted_content BLOB NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX (channel_id),
			INDEX (session_id)
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	return nil
}

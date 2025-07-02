package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

	// For MySQL, try to create the database first if it doesn't exist
	if cfg.Driver == "mysql" {
		// Parse MySQL connection string to extract database name
		// Format: username:password@protocol(address)/dbname?param=value
		connString := cfg.ConnectionString

		// Find the database name
		dbNameStart := strings.LastIndex(connString, "/")
		if dbNameStart == -1 {
			return fmt.Errorf("invalid MySQL connection string format")
		}

		dbNameEnd := strings.Index(connString[dbNameStart+1:], "?")
		var dbName string
		if dbNameEnd == -1 {
			dbName = connString[dbNameStart+1:]
		} else {
			dbName = connString[dbNameStart+1 : dbNameStart+1+dbNameEnd]
		}

		// Create connection string without database name
		baseConnString := connString[:dbNameStart+1]

		// Connect without specifying a database
		tempDB, tempErr := sql.Open(cfg.Driver, baseConnString)
		if tempErr == nil {
			defer tempDB.Close()

			// Create the database if it doesn't exist
			_, execErr := tempDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName))
			if execErr != nil {
				fmt.Printf("Warning: Could not create database: %v\n", execErr)
			} else {
				fmt.Printf("Database '%s' created or already exists\n", dbName)
			}
		} else {
			fmt.Printf("Warning: Could not connect to MySQL server: %v\n", tempErr)
		}
	}

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
		"group_messages",
		"group_members",
		"chat_groups",
		"channel_messages",
		"channel_members",
		"channels",
		"messages",
		"user_avatars",
		"user_settings",
		"users",
		"otp",
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
			phone VARCHAR(20) UNIQUE NOT NULL,
			username VARCHAR(30) UNIQUE,
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

	// Create OTP table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS otp (
			id INT AUTO_INCREMENT PRIMARY KEY,
			phone VARCHAR(20) NOT NULL,
			code VARCHAR(6) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			verified BOOLEAN DEFAULT FALSE,
			failed_attempts INT DEFAULT 0,
			INDEX (phone)
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
			type ENUM('message', 'channel_message', 'channel_create', 'channel_join', 'group_message', 'group_create', 'group_join') NOT NULL,
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

	// Create groups table for group chats
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS chat_groups (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			creator_address VARCHAR(46) NOT NULL,
			photo_url VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX (creator_address)
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create group_members table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS group_members (
			group_id VARCHAR(64) NOT NULL,
			user_address VARCHAR(46) NOT NULL,
			role ENUM('admin', 'member') NOT NULL DEFAULT 'member',
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (group_id, user_address),
			INDEX (user_address),
			FOREIGN KEY (group_id) REFERENCES chat_groups(id) ON DELETE CASCADE
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create group_messages table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS group_messages (
			id VARCHAR(64) PRIMARY KEY,
			group_id VARCHAR(64) NOT NULL,
			sender_address VARCHAR(46) NOT NULL,
			content BLOB NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			block_id VARCHAR(64) NULL,
			INDEX (group_id),
			INDEX (sender_address),
			INDEX (block_id),
			FOREIGN KEY (group_id) REFERENCES chat_groups(id) ON DELETE CASCADE
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create user_settings table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS user_settings (
			user_id INT PRIMARY KEY,
			nickname VARCHAR(50),
			theme ENUM('light', 'dark', 'system') DEFAULT 'system',
			notification_enabled BOOLEAN DEFAULT TRUE,
			sound_enabled BOOLEAN DEFAULT TRUE,
			language VARCHAR(10) DEFAULT 'en',
			auto_download_media BOOLEAN DEFAULT TRUE,
			privacy_last_seen ENUM('everyone', 'contacts', 'nobody') DEFAULT 'everyone',
			privacy_profile_photo ENUM('everyone', 'contacts', 'nobody') DEFAULT 'everyone',
			privacy_status ENUM('everyone', 'contacts', 'nobody') DEFAULT 'everyone',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	// Create user_avatars table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS user_avatars (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			file_path VARCHAR(255) NOT NULL,
			file_name VARCHAR(255) NOT NULL,
			file_size INT NOT NULL,
			mime_type VARCHAR(100) NOT NULL,
			width INT NOT NULL,
			height INT NOT NULL,
			is_active BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX (user_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB ROW_FORMAT=DYNAMIC
	`)
	if err != nil {
		return err
	}

	return nil
}

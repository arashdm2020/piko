-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS piko DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE piko;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    public_key BLOB NOT NULL,
    address VARCHAR(46) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC;

-- Create messages table
CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(64) PRIMARY KEY,
    sender_address VARCHAR(46) NOT NULL,
    recipient_address VARCHAR(46) NOT NULL,
    encrypted_content BLOB NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status ENUM('pending', 'delivered', 'read') DEFAULT 'pending',
    expiration_time TIMESTAMP NULL,
    block_id VARCHAR(64) NULL,
    INDEX (sender_address),
    INDEX (recipient_address),
    INDEX (block_id)
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC;

-- Create channels table
CREATE TABLE IF NOT EXISTS channels (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    admin_address VARCHAR(46) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX (admin_address)
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC;

-- Create channel_members table
CREATE TABLE IF NOT EXISTS channel_members (
    channel_id VARCHAR(64) NOT NULL,
    user_address VARCHAR(46) NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (channel_id, user_address),
    INDEX (user_address)
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC;

-- Create channel_messages table
CREATE TABLE IF NOT EXISTS channel_messages (
    id VARCHAR(64) PRIMARY KEY,
    channel_id VARCHAR(64) NOT NULL,
    sender_address VARCHAR(46) NOT NULL,
    encrypted_content BLOB NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    block_id VARCHAR(64) NULL,
    INDEX (channel_id),
    INDEX (sender_address),
    INDEX (block_id)
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC;

-- Create blocks table
CREATE TABLE IF NOT EXISTS blocks (
    id VARCHAR(64) PRIMARY KEY,
    previous_hash VARCHAR(64) NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    merkle_root VARCHAR(64) NOT NULL,
    nonce BIGINT NOT NULL,
    height INT NOT NULL,
    INDEX (height)
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC;

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    hash VARCHAR(64) PRIMARY KEY,
    block_id VARCHAR(64) NOT NULL,
    type ENUM('message', 'channel_message', 'channel_create', 'channel_join') NOT NULL,
    data_id VARCHAR(64) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX (block_id),
    INDEX (data_id)
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC; 
package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

// GenerateRandomBytes generates a random byte slice of the specified length
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// HashSHA256 creates a SHA-256 hash of the input data
func HashSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HashSHA256String creates a SHA-256 hash of the input string and returns it as a hex string
func HashSHA256String(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateUniqueID generates a unique ID based on timestamp and random data
func GenerateUniqueID() string {
	// Combine current timestamp with random data
	timestamp := time.Now().UnixNano()
	random, _ := GenerateRandomBytes(8) // Ignore error for simplicity
	
	// Create a hash of the combined data
	data := fmt.Sprintf("%d%x", timestamp, random)
	return HashSHA256String(data)
}

// GenerateMessageID generates a unique ID for a message
func GenerateMessageID(senderAddress, recipientAddress string) string {
	// Combine sender, recipient, and timestamp with random data
	timestamp := time.Now().UnixNano()
	random, _ := GenerateRandomBytes(4) // Ignore error for simplicity
	
	// Create a hash of the combined data
	data := fmt.Sprintf("%s%s%d%x", senderAddress, recipientAddress, timestamp, random)
	return HashSHA256String(data)
}

// GenerateChannelID generates a unique ID for a channel
func GenerateChannelID(adminAddress, name string) string {
	// Combine admin address, channel name, and timestamp with random data
	timestamp := time.Now().UnixNano()
	random, _ := GenerateRandomBytes(4) // Ignore error for simplicity
	
	// Create a hash of the combined data
	data := fmt.Sprintf("%s%s%d%x", adminAddress, name, timestamp, random)
	return HashSHA256String(data)
} 
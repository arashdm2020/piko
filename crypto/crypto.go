package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/argon2"
)

var (
	// ErrInvalidPublicKey is returned when a public key is invalid
	ErrInvalidPublicKey = errors.New("invalid public key")
	// ErrInvalidPrivateKey is returned when a private key is invalid
	ErrInvalidPrivateKey = errors.New("invalid private key")
	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errors.New("invalid address")
)

// KeyPair represents a public/private key pair
type KeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

// GenerateKeyPair generates a new Ed25519 key pair
func GenerateKeyPair() (*KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// Sign signs a message using the private key
func Sign(privateKey []byte, message []byte) ([]byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, ErrInvalidPrivateKey
	}
	return ed25519.Sign(privateKey, message), nil
}

// Verify verifies a signature using the public key
func Verify(publicKey []byte, message []byte, signature []byte) (bool, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return false, ErrInvalidPublicKey
	}
	return ed25519.Verify(publicKey, message, signature), nil
}

// GenerateAddress generates a Base58 address from a public key
func GenerateAddress(publicKey []byte, length int) (string, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return "", ErrInvalidPublicKey
	}

	// Hash the public key using SHA-256
	hash := sha256.Sum256(publicKey)

	// Encode the hash using Base58
	address := base58.Encode(hash[:])

	// Truncate or pad the address to the desired length
	if len(address) > length {
		address = address[:length]
	} else if len(address) < length {
		// This should not happen with SHA-256 and Base58, but just in case
		padding := length - len(address)
		for i := 0; i < padding; i++ {
			address += "1" // Use "1" for padding (common in Base58)
		}
	}

	return address, nil
}

// ValidateAddress validates a Base58 address
func ValidateAddress(address string, length int) bool {
	// Check length
	if len(address) != length {
		return false
	}

	// Check if it's a valid Base58 string
	_, err := base58.Decode(address)
	return err == nil
}

// HashPassword hashes a password using Argon2id
func HashPassword(password string, salt []byte, time, memory uint32, threads uint8, keyLen uint32) ([]byte, error) {
	if salt == nil {
		var err error
		salt = make([]byte, 16)
		if _, err = rand.Read(salt); err != nil {
			return nil, err
		}
	}

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	// Combine salt and hash
	result := make([]byte, len(salt)+len(hash))
	copy(result, salt)
	copy(result[len(salt):], hash)

	return result, nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password string, encodedHash string, time, memory uint32, threads uint8, keyLen uint32) (bool, error) {
	// Decode the hash from base64
	hashBytes, err := base64.StdEncoding.DecodeString(encodedHash)
	if err != nil {
		return false, err
	}

	// Extract salt (first 16 bytes) and hash
	if len(hashBytes) < 16 {
		return false, fmt.Errorf("invalid hash format")
	}
	salt := hashBytes[:16]

	// Compute hash with the same parameters and salt
	newHash, err := HashPassword(password, salt, time, memory, threads, keyLen)
	if err != nil {
		return false, err
	}

	// Compare hashes
	return hex.EncodeToString(hashBytes) == hex.EncodeToString(newHash), nil
}

// GenerateRandomBytes generates random bytes of the specified length
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// EncodeBase58 encodes bytes to a Base58 string
func EncodeBase58(bytes []byte) string {
	return base58.Encode(bytes)
}

// DecodeBase58 decodes a Base58 string to bytes
func DecodeBase58(str string) ([]byte, error) {
	return base58.Decode(str)
}

// EncodeBase64 encodes bytes to a Base64 string
func EncodeBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// DecodeBase64 decodes a Base64 string to bytes
func DecodeBase64(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

// HashSHA256 computes the SHA-256 hash of data
func HashSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HashToHex computes the SHA-256 hash of data and returns it as a hex string
func HashToHex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

package config

import (
	"encoding/json"
	"os"
	"time"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `json:"server"`
	Database   DatabaseConfig   `json:"database"`
	Auth       AuthConfig       `json:"auth"`
	CORS       CORSConfig       `json:"cors"`
	Crypto     CryptoConfig     `json:"crypto"`
	Blockchain BlockchainConfig `json:"blockchain"`
	SMS        SMSConfig        `json:"sms"`
}

// ServerConfig represents server-specific configuration
type ServerConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	ReadTimeout     time.Duration `json:"readTimeout"`
	WriteTimeout    time.Duration `json:"writeTimeout"`
	ShutdownTimeout time.Duration `json:"shutdownTimeout"`
}

// DatabaseConfig represents database-specific configuration
type DatabaseConfig struct {
	Driver           string `json:"driver"`
	ConnectionString string `json:"connectionString"`
	MaxOpenConns     int    `json:"maxOpenConns"`
	MaxIdleConns     int    `json:"maxIdleConns"`
	ConnMaxLifetime  int    `json:"connMaxLifetime"`
}

// AuthConfig represents authentication-specific configuration
type AuthConfig struct {
	JWTSecret            string        `json:"jwtSecret"`
	JWTExpirationTime    time.Duration `json:"jwtExpirationTime"`
	RefreshTokenDuration time.Duration `json:"refreshTokenDuration"`
	Argon2Time           uint32        `json:"argon2Time"`
	Argon2Memory         uint32        `json:"argon2Memory"`
	Argon2Threads        uint8         `json:"argon2Threads"`
	Argon2KeyLength      uint32        `json:"argon2KeyLength"`
	OTPExpiryMinutes     int           `json:"otpExpiryMinutes"`
}

// CORSConfig represents CORS-specific configuration
type CORSConfig struct {
	AllowOrigins     string `json:"allowOrigins"`
	AllowMethods     string `json:"allowMethods"`
	AllowHeaders     string `json:"allowHeaders"`
	AllowCredentials bool   `json:"allowCredentials"`
	MaxAge           int    `json:"maxAge"`
}

// CryptoConfig represents cryptography-specific configuration
type CryptoConfig struct {
	KeyAlgorithm     string `json:"keyAlgorithm"`
	AddressAlgorithm string `json:"addressAlgorithm"`
	AddressLength    int    `json:"addressLength"`
}

// BlockchainConfig represents blockchain-specific configuration
type BlockchainConfig struct {
	BlockTime       time.Duration `json:"blockTime"`
	DataDir         string        `json:"dataDir"`
	StorageType     string        `json:"storageType"`
	MempoolCapacity int           `json:"mempoolCapacity"`
}

// SMSConfig represents SMS service configuration
type SMSConfig struct {
	Provider    string `json:"provider"`
	APIKey      string `json:"apiKey"`
	SenderID    string `json:"senderId"`
	BaseURL     string `json:"baseUrl"`
	IsEnabled   bool   `json:"isEnabled"`
	PatternCode string `json:"patternCode"`
}

// LoadConfig loads the configuration from the specified file path
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			ReadTimeout:     time.Second * 15,
			WriteTimeout:    time.Second * 15,
			ShutdownTimeout: time.Second * 30,
		},
		Database: DatabaseConfig{
			Driver:           "mysql",
			ConnectionString: "root@tcp(localhost:3306)/piko?parseTime=true",
			MaxOpenConns:     25,
			MaxIdleConns:     25,
			ConnMaxLifetime:  300,
		},
		Auth: AuthConfig{
			JWTSecret:            "change-me-in-production",
			JWTExpirationTime:    time.Hour * 24 * 30, // Extended to 30 days for persistent login
			RefreshTokenDuration: time.Hour * 24 * 7,
			Argon2Time:           1,
			Argon2Memory:         64 * 1024,
			Argon2Threads:        4,
			Argon2KeyLength:      32,
			OTPExpiryMinutes:     5,
		},
		CORS: CORSConfig{
			AllowOrigins:     "*",
			AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
			AllowCredentials: true,
			MaxAge:           86400,
		},
		Crypto: CryptoConfig{
			KeyAlgorithm:     "ed25519",
			AddressAlgorithm: "base58",
			AddressLength:    46,
		},
		Blockchain: BlockchainConfig{
			BlockTime:       time.Second * 10,
			DataDir:         "./data",
			StorageType:     "badger",
			MempoolCapacity: 10000,
		},
		SMS: SMSConfig{
			Provider:    "ippanel",
			APIKey:      "OWVmNGI4MTctODhkMi00OWIxLWI4ZGUtMDhjZTg2NGE1MTAxMjc0ZDAwZjIyYTZkNjA2ODNiNDg1Y2QwZjhkODk4Mjk=",
			SenderID:    "+983000505",
			BaseURL:     "https://edge.ippanel.com/v1",
			IsEnabled:   true,
			PatternCode: "9muuwhyyw2s1ag5",
		},
	}
}

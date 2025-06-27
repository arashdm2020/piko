# Piko - Decentralized Encrypted Messaging with Blockchain Integration

Piko is a secure, scalable backend API for a decentralized encrypted messaging application with blockchain integration. It provides end-to-end encrypted messaging capabilities with the security and immutability of blockchain technology.

## Features

- **End-to-End Encrypted Messaging**: Secure communication between users with Ed25519 cryptography
- **Blockchain Integration**: Messages are stored and verified on a blockchain for immutability
- **Channel/Group Messaging**: Support for group conversations
- **Secret Chat**: Temporary, anonymous chat rooms that require no authentication
- **User Authentication**: JWT-based authentication with Argon2 password hashing
- **WebSocket Support**: Real-time communication between users
- **RESTful API**: Well-structured API for easy integration with frontend applications

## Technology Stack

- **Backend**: Go (Fiber framework)
- **Database**: MySQL/SQLite
- **Authentication**: JWT tokens with Argon2 password hashing
- **Cryptography**: Ed25519 for keys and signatures
- **Blockchain**: Custom implementation with proof-of-work consensus
- **Real-time Communication**: WebSockets

## Project Structure

```
piko/
├── api/            # API routes and error handling
├── blockchain/     # Blockchain implementation
├── config/         # Configuration structures and loading
├── crypto/         # Cryptographic utilities
├── database/       # Database connection and schema
├── handlers/       # API endpoint handlers
├── middleware/     # Authentication middleware
├── models/         # Data models
├── utils/          # Utility functions
├── websocket/      # WebSocket implementation
├── main.go         # Application entry point
└── Dockerfile      # Docker configuration
```

## API Endpoints

### Authentication
- `POST /api/auth/register`: Register a new user
- `POST /api/auth/login`: Login and get JWT token

### User Profile
- `GET /api/profile`: Get user profile
- `PUT /api/profile`: Update user profile

### Messages
- `POST /api/messages`: Send a message
- `GET /api/messages/inbox`: Get received messages
- `GET /api/messages/sent`: Get sent messages
- `GET /api/messages/:id`: Get a specific message
- `DELETE /api/messages/:id`: Delete a message

### Channels
- `POST /api/channels`: Create a channel
- `GET /api/channels`: Get all channels
- `GET /api/channels/:id`: Get a specific channel
- `PUT /api/channels/:id`: Update a channel
- `DELETE /api/channels/:id`: Delete a channel
- `POST /api/channels/:id/members`: Add a member to a channel
- `GET /api/channels/:id/members`: Get channel members
- `DELETE /api/channels/:id/members/:address`: Remove a member from a channel
- `POST /api/channels/:id/messages`: Send a message to a channel
- `GET /api/channels/:id/messages`: Get channel messages
- `DELETE /api/channels/:channel_id/messages/:message_id`: Delete a channel message

### Blockchain
- `GET /api/blocks/:id`: Get a block by ID
- `GET /api/blocks/height/:height`: Get a block by height
- `GET /api/transactions/:hash`: Get a transaction by hash
- `GET /api/explore/:address`: Explore transactions for an address
- `GET /api/proof/:message_id`: Get proof for a message
- `GET /api/blockchain/stats`: Get blockchain statistics

### Secret Chat (No Authentication Required)
- `POST /api/secret-chat/create`: Create a new secret chat
- `POST /api/secret-chat/join`: Join an existing secret chat
- `POST /api/secret-chat/send`: Send a message in a secret chat
- `GET /api/secret-chat/messages/:channel_id`: Get messages from a secret chat
- `DELETE /api/secret-chat/:channel_id`: Delete a secret chat
- `GET /ws/secret/:session_id`: WebSocket connection for real-time secret chat updates

### WebSocket
- `GET /ws`: WebSocket connection for real-time updates

## Getting Started

### Prerequisites

- Go 1.18 or higher
- MySQL or SQLite
- Docker (optional)

### Configuration

Configuration is stored in `config/config.json`. You can modify this file to change database settings, JWT secret, blockchain parameters, etc.

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "database": {
    "driver": "mysql",
    "connectionString": "user:password@tcp(localhost:3306)/piko?parseTime=true"
  }
}
```

### Running Locally

1. Clone the repository
2. Configure `config/config.json`
3. Run the application:

```bash
go run main.go
```

### Running with Docker

1. Build the Docker image:

```bash
docker build -t piko .
```

2. Run the container:

```bash
docker run -p 8080:8080 piko
```

## Security Considerations

- Change the JWT secret in production
- Use HTTPS in production
- Regularly rotate encryption keys
- Monitor blockchain integrity

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
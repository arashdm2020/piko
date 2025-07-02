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
- **Message Status Tracking**: Track message delivery and read status (one tick, two ticks, blue ticks)
- **Group Chats**: Create and manage group conversations with multiple participants and role-based permissions
- **User Settings**: Customizable user settings including theme, language, and privacy options
- **User Profiles**: Customizable profiles with nicknames and avatar photos

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
- `POST /api/auth/register`: Register a new user - Step 1: Send OTP to phone
- `POST /api/auth/verify-register`: Register a new user - Step 2: Verify OTP and create account
- `POST /api/auth/login`: Login - Step 1: Send OTP to phone
- `POST /api/auth/verify-login`: Login - Step 2: Verify OTP and get JWT token

### User Profile
- `GET /api/profile`: Get user profile
- `PUT /api/profile`: Update user profile
- `PUT /api/profile/username`: Set or update username

### User Settings
- `GET /api/settings`: Get user settings
- `PUT /api/settings`: Update user settings
- `PUT /api/settings/nickname`: Update user nickname

### User Avatars
- `POST /api/avatars`: Upload a new avatar
- `GET /api/avatars`: Get all user avatars
- `GET /api/avatars/active`: Get active avatar
- `PUT /api/avatars/:id/active`: Set an avatar as active
- `DELETE /api/avatars/:id`: Delete an avatar
- `GET /api/avatars/:id/file`: Serve avatar file

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

### Group Chats
- `POST /api/groups`: Create a new group
- `GET /api/groups`: Get all groups the user is a member of
- `GET /api/groups/:id`: Get details of a specific group
- `PUT /api/groups/:id`: Update group information
- `DELETE /api/groups/:id`: Delete a group
- `GET /api/groups/:id/members`: Get all members of a group
- `POST /api/groups/:id/members`: Add a member to a group
- `DELETE /api/groups/:id/members/:address`: Remove a member from a group
- `POST /api/groups/:id/messages`: Send a message to a group
- `GET /api/groups/:id/messages`: Get messages from a group

## Phone Authentication

Piko now uses phone-based OTP (One-Time Password) authentication instead of traditional password-based authentication. This provides a more secure and user-friendly authentication experience:

### Registration Process
1. User provides their phone number
2. System sends a 6-digit OTP via SMS to the provided phone number
3. User verifies their phone number by entering the OTP
4. Upon successful verification, a new account is created with a unique blockchain address
5. The user's private key is returned ONLY during this initial registration and must be stored securely by the client

**Important**: Only users who have correctly verified their OTP code will be stored in the database. The system will create a unique blockchain address for each verified user. Failed verification attempts will not result in user account creation.

### Login Process
1. User provides their phone number
2. System sends a 6-digit OTP via SMS to the provided phone number
3. User verifies their phone number by entering the OTP
4. Upon successful verification, a JWT token is issued

### Persistent Login
- JWT tokens are valid for 30 days, providing a persistent login experience
- No password is required for authentication
- Users can securely access their account from any device by receiving an OTP

### SMS Configuration
SMS delivery is configured in the `config.json` file. The application uses IPPanel's Pattern SMS service for sending OTP codes:

```json
"sms": {
  "provider": "ippanel",
  "apiKey": "your-api-key",
  "senderId": "+983000505",
  "baseUrl": "https://edge.ippanel.com/v1",
  "isEnabled": true,
  "patternCode": "your-pattern-code"
}
```

The application uses IPPanel's Pattern SMS API to send verification codes. The pattern code is configured to use the "verfication-code" variable in the pattern template. For testing or development purposes, you can set `isEnabled` to `false` to use the mock SMS provider that simply logs the OTP to the console.

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

## Group Chat Features

Piko provides comprehensive group chat functionality similar to Telegram groups:

### Group Management
- Create groups with name, description, and optional photo URL
- Update group information (admins only)
- Delete groups (creator only)

### Member Management
- Add new members to groups (admins only)
- Remove members from groups (admins only or self-removal)
- Assign admin roles to members

### Messaging
- Send encrypted messages to groups
- Retrieve message history with pagination
- Real-time message delivery via WebSocket

### Roles and Permissions
- Creator: Has full control over the group, including deletion
- Admin: Can manage group settings and members
- Member: Can send and receive messages

### Real-time Updates
- New message notifications via WebSocket
- Group membership changes

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
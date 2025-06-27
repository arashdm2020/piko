# Piko API Documentation

This document provides detailed information about the Piko API endpoints with examples.

## Authentication

### Register a New User

**Endpoint**: `POST /api/auth/register`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response**:
```json
{
  "user_id": 1,
  "address": "PikoXYZ123...",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Login

**Endpoint**: `POST /api/auth/login`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response**:
```json
{
  "user_id": 1,
  "address": "PikoXYZ123...",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## User Profile

### Get User Profile

**Endpoint**: `GET /api/profile`

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**:
```json
{
  "id": 1,
  "email": "user@example.com",
  "address": "PikoXYZ123...",
  "public_key": "base64_encoded_public_key",
  "created_at": "2023-06-15T10:30:00Z"
}
```

### Update User Profile

**Endpoint**: `PUT /api/profile`

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body**:
```json
{
  "email": "newemail@example.com",
  "phone": "1234567890"
}
```

**Response**:
```json
{
  "message": "Profile updated successfully"
}
```

## Messages

### Send a Message

**Endpoint**: `POST /api/messages`

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body**:
```json
{
  "recipient_address": "PikoABC456...",
  "content": "encrypted_message_content",
  "expiration_time": "2023-06-20T00:00:00Z"
}
```

**Response**:
```json
{
  "message_id": "msg123456",
  "timestamp": "2023-06-15T11:45:00Z"
}
```

### Get Inbox Messages

**Endpoint**: `GET /api/messages/inbox`

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Query Parameters**:
- `page`: Page number (default: 1)
- `limit`: Number of messages per page (default: 20)

**Response**:
```json
{
  "messages": [
    {
      "id": "msg123456",
      "sender_address": "PikoABC456...",
      "encrypted_content": "encrypted_message_content",
      "timestamp": "2023-06-15T11:45:00Z",
      "status": "delivered",
      "block_id": "block789012"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

### Get a Specific Message

**Endpoint**: `GET /api/messages/:id`

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**:
```json
{
  "id": "msg123456",
  "sender_address": "PikoABC456...",
  "recipient_address": "PikoXYZ123...",
  "encrypted_content": "encrypted_message_content",
  "timestamp": "2023-06-15T11:45:00Z",
  "status": "read",
  "block_id": "block789012",
  "proof": {
    "block_id": "block789012",
    "merkle_path": ["hash1", "hash2"]
  }
}
```

## Channels

### Create a Channel

**Endpoint**: `POST /api/channels`

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body**:
```json
{
  "name": "My Channel"
}
```

**Response**:
```json
{
  "id": "channel123",
  "name": "My Channel",
  "admin_address": "PikoXYZ123...",
  "created_at": "2023-06-15T14:00:00Z"
}
```

### Send a Message to a Channel

**Endpoint**: `POST /api/channels/:id/messages`

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body**:
```json
{
  "content": "encrypted_channel_message_content"
}
```

**Response**:
```json
{
  "message_id": "cmsg456789",
  "timestamp": "2023-06-15T14:15:00Z"
}
```

## Blockchain

### Get Block by ID

**Endpoint**: `GET /api/blocks/:id`

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**:
```json
{
  "id": "block789012",
  "previous_hash": "block654321",
  "timestamp": "2023-06-15T12:00:00Z",
  "merkle_root": "merkle_root_hash",
  "nonce": 12345,
  "height": 42,
  "transactions": [
    {
      "hash": "tx987654",
      "block_id": "block789012",
      "type": "message",
      "data_id": "msg123456",
      "timestamp": "2023-06-15T11:45:00Z"
    }
  ]
}
```

### Get Blockchain Statistics

**Endpoint**: `GET /api/blockchain/stats`

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**:
```json
{
  "block_count": 42,
  "transaction_count": 156,
  "transaction_types": {
    "message": 98,
    "channel_message": 45,
    "channel_create": 8,
    "channel_join": 5
  },
  "latest_block_time": "2023-06-15T12:00:00Z"
}
```

## WebSocket

### Connect to WebSocket

**Endpoint**: `GET /ws`

**Query Parameters**:
- `token`: JWT token

**Events**:

1. New Message:
```json
{
  "type": "new_message",
  "data": {
    "id": "msg123456",
    "sender_address": "PikoABC456...",
    "timestamp": "2023-06-15T11:45:00Z"
  }
}
```

2. Message Status Update:
```json
{
  "type": "message_status",
  "data": {
    "id": "msg123456",
    "status": "read",
    "timestamp": "2023-06-15T11:46:00Z"
  }
}
```

3. New Block:
```json
{
  "type": "new_block",
  "data": {
    "id": "block789012",
    "height": 42,
    "timestamp": "2023-06-15T12:00:00Z",
    "transaction_count": 5
  }
}
```

## Secret Chat (No Authentication Required)

### Create a Secret Chat

**Endpoint**: `POST /api/secret-chat/create`

**Request Body**:
```json
{}
```

**Response**:
```json
{
  "channel_id": "c52-13gtr3",
  "expires_at": "2023-06-16T14:00:00Z"
}
```

### Join a Secret Chat

**Endpoint**: `POST /api/secret-chat/join`

**Request Body**:
```json
{
  "channel_id": "c52-13gtr3",
  "display_name": "Anonymous User"
}
```

**Response**:
```json
{
  "session_id": "a1b2c3d4e5f6...",
  "channel_id": "c52-13gtr3",
  "expires_at": "2023-06-16T14:00:00Z",
  "websocket_url": "ws://example.com/ws/secret/a1b2c3d4e5f6..."
}
```

### Send a Secret Chat Message

**Endpoint**: `POST /api/secret-chat/send`

**Request Body**:
```json
{
  "session_id": "a1b2c3d4e5f6...",
  "encrypted_content": "base64_encoded_encrypted_content"
}
```

**Response**:
```json
{
  "id": "msg789012"
}
```

### Get Secret Chat Messages

**Endpoint**: `GET /api/secret-chat/messages/:channel_id`

**Query Parameters**:
- `session_id`: Session ID from join response
- `limit`: Number of messages to retrieve (default: 50)
- `offset`: Offset for pagination (default: 0)

**Response**:
```json
[
  {
    "id": "msg789012",
    "channel_id": "c52-13gtr3",
    "display_name": "Anonymous User",
    "encrypted_content": "base64_encoded_encrypted_content",
    "timestamp": "2023-06-15T14:30:00Z"
  }
]
```

### Delete a Secret Chat

**Endpoint**: `DELETE /api/secret-chat/:channel_id`

**Query Parameters**:
- `session_id`: Session ID from join response

**Response**:
```json
{
  "success": true
}
```

### Connect to Secret Chat WebSocket

**Endpoint**: `GET /ws/secret/:session_id`

**Events**:

1. New Secret Chat Message:
```json
{
  "type": "secret_chat_message",
  "payload": {
    "id": "msg789012",
    "channel_id": "c52-13gtr3",
    "display_name": "Anonymous User",
    "encrypted_content": "base64_encoded_encrypted_content",
    "timestamp": "2023-06-15T14:30:00Z"
  }
}
```

2. Secret Chat Deleted:
```json
{
  "type": "secret_chat_deleted",
  "payload": {
    "channel_id": "c52-13gtr3"
  }
}
``` 
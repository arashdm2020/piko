package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/piko/piko/middleware"
	"github.com/piko/piko/models"
)

// GetBlock handles retrieving a block by its ID
func GetBlock() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get block ID from URL parameter
		blockID := c.Params("id")
		if blockID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Block ID is required",
			})
		}

		// Get block from database
		block, err := models.GetBlockByID(blockID)
		if err != nil {
			if errors.Is(err, models.ErrBlockNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Block not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get block",
			})
		}

		// Return block
		return c.Status(fiber.StatusOK).JSON(block)
	}
}

// GetBlockByHeight handles retrieving a block by its height
func GetBlockByHeight() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get block height from URL parameter
		heightStr := c.Params("height")
		if heightStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Block height is required",
			})
		}

		// Parse height
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid block height",
			})
		}

		// Get block from database
		block, err := models.GetBlockByHeight(height)
		if err != nil {
			if errors.Is(err, models.ErrBlockNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Block not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get block",
			})
		}

		// Return block
		return c.Status(fiber.StatusOK).JSON(block)
	}
}

// GetTransaction handles retrieving a transaction by its hash
func GetTransaction() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get transaction hash from URL parameter
		hash := c.Params("hash")
		if hash == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Transaction hash is required",
			})
		}

		// Get transaction from database
		transaction, err := models.GetTransactionByHash(hash)
		if err != nil {
			if errors.Is(err, models.ErrTransactionNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Transaction not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get transaction",
			})
		}

		// Return transaction
		return c.Status(fiber.StatusOK).JSON(transaction)
	}
}

// ExploreAddress handles retrieving all transactions related to an address
func ExploreAddress() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get address from URL parameter
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Address is required",
			})
		}

		// Get transactions from database
		transactions, err := models.GetTransactionsByAddress(address)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get transactions",
			})
		}

		// Return transactions
		return c.Status(fiber.StatusOK).JSON(transactions)
	}
}

// GetProof handles retrieving a Merkle proof for a message
func GetProof() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get message ID from URL parameter
		messageID := c.Params("message_id")
		if messageID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Message ID is required",
			})
		}

		// Get message from database
		message, err := models.GetMessageByID(messageID)
		if err != nil {
			if errors.Is(err, models.ErrMessageNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Message not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get message",
			})
		}

		// Check if user is sender or recipient
		if message.SenderAddress != userAddress && message.RecipientAddress != userAddress {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Check if message is in a block
		if message.BlockID == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Message is not in a block yet",
			})
		}

		// TODO: Implement Merkle proof generation
		// For now, return a placeholder
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message_id": messageID,
			"block_id":   *message.BlockID,
			"proof":      []string{"placeholder_proof"},
			"root_hash":  "placeholder_root_hash",
		})
	}
}

// GetBlockchainStats handles retrieving statistics about the blockchain
func GetBlockchainStats() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get blockchain stats from database
		stats, err := models.GetBlockchainStats()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get blockchain stats",
			})
		}

		// Return stats
		return c.Status(fiber.StatusOK).JSON(stats)
	}
} 
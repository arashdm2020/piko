package utils

import (
	"github.com/gofiber/fiber/v2"
)

// Response types
const (
	SUCCESS        = "success"
	RESPONSE_ERROR = "error"
)

// SuccessResponse sends a standardized success response
func SuccessResponse(c *fiber.Ctx, statusCode int, data interface{}) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status": SUCCESS,
		"data":   data,
	})
}

// ErrorResponse sends a standardized error response
func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status": RESPONSE_ERROR,
		"error":  message,
	})
}

// ValidationErrorResponse sends a standardized validation error response
func ValidationErrorResponse(c *fiber.Ctx, errors map[string]string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"status": RESPONSE_ERROR,
		"error":  "Validation failed",
		"errors": errors,
	})
}

// NotFoundResponse sends a standardized not found response
func NotFoundResponse(c *fiber.Ctx, resource string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"status": RESPONSE_ERROR,
		"error":  resource + " not found",
	})
}

// UnauthorizedResponse sends a standardized unauthorized response
func UnauthorizedResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"status": RESPONSE_ERROR,
		"error":  "Unauthorized access",
	})
}

// ForbiddenResponse sends a standardized forbidden response
func ForbiddenResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"status": RESPONSE_ERROR,
		"error":  "Access forbidden",
	})
}

// InternalErrorResponse sends a standardized internal server error response
func InternalErrorResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"status": RESPONSE_ERROR,
		"error":  "Internal server error",
	})
}

// CreatedResponse sends a standardized resource created response
func CreatedResponse(c *fiber.Ctx, data interface{}) error {
	return SuccessResponse(c, fiber.StatusCreated, data)
}

// OKResponse sends a standardized OK response
func OKResponse(c *fiber.Ctx, data interface{}) error {
	return SuccessResponse(c, fiber.StatusOK, data)
}

package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page  int
	Limit int
}

// GetPaginationParams extracts pagination parameters from the request
func GetPaginationParams(c *fiber.Ctx) PaginationParams {
	// Default values
	page := 1
	limit := 20
	
	// Try to get page from query parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if pageVal, err := strconv.Atoi(pageStr); err == nil && pageVal > 0 {
			page = pageVal
		}
	}
	
	// Try to get limit from query parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitVal, err := strconv.Atoi(limitStr); err == nil && limitVal > 0 {
			// Cap the limit to prevent excessive queries
			if limitVal > 100 {
				limitVal = 100
			}
			limit = limitVal
		}
	}
	
	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

// CalculateOffset calculates the offset for SQL queries based on page and limit
func (p *PaginationParams) CalculateOffset() int {
	return (p.Page - 1) * p.Limit
}

// GeneratePaginationResponse creates a standardized pagination response
func GeneratePaginationResponse(total int, params PaginationParams, data interface{}) fiber.Map {
	totalPages := (total + params.Limit - 1) / params.Limit
	
	return fiber.Map{
		"data":        data,
		"total":       total,
		"page":        params.Page,
		"limit":       params.Limit,
		"total_pages": totalPages,
		"has_next":    params.Page < totalPages,
		"has_prev":    params.Page > 1,
	}
} 
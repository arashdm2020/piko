package utils

import (
	"net/mail"
	"regexp"
	"strings"
)

// IsValidEmail checks if the provided string is a valid email address
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsValidPhone checks if the provided string is a valid phone number
// This is a simple validation, you might want to use a more sophisticated library for production
func IsValidPhone(phone string) bool {
	if phone == "" {
		return false
	}
	re := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	return re.MatchString(phone)
}

// IsValidPassword checks if the provided password meets security requirements
// Password must be at least 8 characters long and contain at least one uppercase letter,
// one lowercase letter, one number, and one special character
func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// SanitizeString removes potentially harmful characters from a string
func SanitizeString(input string) string {
	// Remove any HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	sanitized := re.ReplaceAllString(input, "")
	
	// Remove any script tags and content
	re = regexp.MustCompile(`(?i)<script[\s\S]*?</script>`)
	sanitized = re.ReplaceAllString(sanitized, "")
	
	// Trim spaces
	sanitized = strings.TrimSpace(sanitized)
	
	return sanitized
}

// IsValidAddress checks if the provided string is a valid blockchain address
func IsValidAddress(address string) bool {
	if address == "" {
		return false
	}
	
	// Check length (adjust based on your address format)
	if len(address) != 46 {
		return false
	}
	
	// Check for valid base58 characters
	re := regexp.MustCompile(`^[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$`)
	return re.MatchString(address)
} 
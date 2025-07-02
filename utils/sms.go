package utils

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	ippanel "github.com/ippanel/go-rest-sdk/v2"
	"github.com/piko/piko/config"
)

var (
	// ErrSMSFailed is returned when SMS sending fails
	ErrSMSFailed = errors.New("failed to send SMS")
)

// SMSConfig represents SMS service configuration
type SMSConfig struct {
	Provider    string
	APIKey      string
	SenderID    string
	BaseURL     string
	IsEnabled   bool
	PatternCode string
}

// FromConfigSMS converts config.SMSConfig to utils.SMSConfig
func FromConfigSMS(cfg *config.SMSConfig) *SMSConfig {
	return &SMSConfig{
		Provider:    cfg.Provider,
		APIKey:      cfg.APIKey,
		SenderID:    cfg.SenderID,
		BaseURL:     cfg.BaseURL,
		IsEnabled:   cfg.IsEnabled,
		PatternCode: cfg.PatternCode,
	}
}

// DefaultSMSConfig returns a default SMS configuration
// In a real application, you would configure this with actual SMS provider details
func DefaultSMSConfig() *SMSConfig {
	return &SMSConfig{
		Provider:    "mock", // Using mock provider by default
		APIKey:      "OWVmNGI4MTctODhkMi00OWIxLWI4ZGUtMDhjZTg2NGE1MTAxMjc0ZDAwZjIyYTZkNjA2ODNiNDg1Y2QwZjhkODk4Mjk=",
		SenderID:    "+983000505", // Default sender number for IPPanel
		BaseURL:     "https://edge.ippanel.com/v1",
		IsEnabled:   false, // Disabled by default
		PatternCode: "9muuwhyyw2s1ag5",
	}
}

// SendSMS sends an SMS message to the specified phone number
func SendSMS(config *SMSConfig, phone, message string) error {
	// If SMS service is disabled, just log the message and return success
	if !config.IsEnabled || config.Provider == "mock" {
		log.Printf("[MOCK SMS] To: %s, Message: %s", phone, message)
		return nil
	}

	// Different providers have different APIs
	switch config.Provider {
	case "ippanel":
		return sendIPPanelSMS(config, phone, message)
	case "twilio":
		return sendTwilioSMS(config, phone, message)
	case "nexmo":
		return sendNexmoSMS(config, phone, message)
	default:
		return fmt.Errorf("unsupported SMS provider: %s", config.Provider)
	}
}

// SendOTP sends an OTP code to the specified phone number
func SendOTP(config *SMSConfig, phone, code string) error {
	log.Printf("SendOTP called with phone=%s, code=%s, provider=%s, isEnabled=%v\n",
		phone, code, config.Provider, config.IsEnabled)

	// If SMS service is disabled or provider is mock, just log the message and return success
	if !config.IsEnabled || config.Provider == "mock" {
		log.Printf("[MOCK SMS] To: %s, OTP Code: %s", phone, code)
		return nil
	}

	// Always use pattern SMS for OTP with IPPanel
	if config.Provider == "ippanel" {
		return sendIPPanelPatternSMS(config, phone, code)
	}

	// For other providers, use regular SMS
	message := fmt.Sprintf("Your PIKO verification code is: %s", code)
	return SendSMS(config, phone, message)
}

// sendIPPanelPatternSMS sends an OTP using IPPanel's pattern SMS API with SDK
func sendIPPanelPatternSMS(config *SMSConfig, phone, code string) error {
	log.Printf("Sending pattern SMS via IPPanel to %s with code %s", phone, code)

	// Format phone number (ensure it starts with country code)
	formattedPhone := formatPhoneNumber(phone)
	log.Printf("Formatted phone number: %s", formattedPhone)

	// Create IPPanel client
	smsClient := ippanel.New(config.APIKey)

	// Prepare pattern values - note the spelling "verfication-code" as per the pattern
	patternValues := map[string]string{
		"verfication-code": code,
	}

	log.Printf("Using pattern code: %s, sender: %s", config.PatternCode, config.SenderID)

	// Send pattern SMS
	messageID, err := smsClient.SendPattern(
		config.PatternCode,
		config.SenderID,
		formattedPhone,
		patternValues,
	)

	if err != nil {
		log.Printf("IPPanel SDK error: %v", err)
		// If it's an IPPanel specific error, log more details
		if ipErr, ok := err.(ippanel.Error); ok {
			log.Printf("IPPanel error code: %d, message: %v", ipErr.Code, ipErr.Message)
		}
		return ErrSMSFailed
	}

	log.Printf("Pattern SMS sent successfully, message ID: %v", messageID)
	return nil
}

// sendIPPanelSMS sends a regular SMS using IPPanel API with SDK
func sendIPPanelSMS(config *SMSConfig, phone, message string) error {
	log.Printf("Sending regular SMS via IPPanel to %s", phone)

	// Format phone number (ensure it starts with country code)
	formattedPhone := formatPhoneNumber(phone)
	log.Printf("Formatted phone number: %s", formattedPhone)

	// Create IPPanel client
	smsClient := ippanel.New(config.APIKey)

	// Send SMS
	messageID, err := smsClient.Send(
		config.SenderID,
		[]string{formattedPhone},
		message,
		"", // Empty summary parameter
	)

	if err != nil {
		log.Printf("IPPanel SDK error: %v", err)
		return ErrSMSFailed
	}

	log.Printf("SMS sent successfully, message ID: %v", messageID)
	return nil
}

// formatPhoneNumber ensures the phone number is in the correct format for IPPanel
func formatPhoneNumber(phone string) string {
	// Remove any spaces
	phone = strings.ReplaceAll(phone, " ", "")

	// Remove any plus sign
	phone = strings.ReplaceAll(phone, "+", "")

	// If the number starts with 0, remove it and add country code
	if strings.HasPrefix(phone, "0") {
		phone = "98" + phone[1:]
	}

	// If the number doesn't start with country code, add it
	if !strings.HasPrefix(phone, "98") {
		phone = "98" + phone
	}

	return phone
}

// sendTwilioSMS sends an SMS using Twilio
func sendTwilioSMS(config *SMSConfig, phone, message string) error {
	// Prepare the form data
	formData := url.Values{}
	formData.Set("To", phone)
	formData.Set("From", config.SenderID)
	formData.Set("Body", message)

	// Create the request
	req, err := http.NewRequest("POST", config.BaseURL+"/2010-04-01/Accounts/"+config.APIKey+"/Messages.json", strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	// Set headers
	req.SetBasicAuth(config.APIKey, config.APIKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusCreated {
		return ErrSMSFailed
	}

	return nil
}

// sendNexmoSMS sends an SMS using Nexmo/Vonage
func sendNexmoSMS(config *SMSConfig, phone, message string) error {
	// Prepare the form data
	formData := url.Values{}
	formData.Set("to", phone)
	formData.Set("from", config.SenderID)
	formData.Set("text", message)
	formData.Set("api_key", config.APIKey)
	formData.Set("api_secret", config.APIKey)

	// Create the request
	req, err := http.NewRequest("POST", config.BaseURL+"/sms/json", strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		return ErrSMSFailed
	}

	return nil
}

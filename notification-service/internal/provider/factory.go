package provider

import (
	"log"
	"os"
)

func NewEmailSender() EmailSender {
	mode := os.Getenv("PROVIDER_MODE")
	if mode == "" {
		mode = "SIMULATED"
	}

	switch mode {
	case "REAL":
		log.Println("[Provider] Using REAL SMTP email sender")
		return NewSMTPEmailSender()
	case "SIMULATED":
		log.Println("[Provider] Using SIMULATED (mock) email sender")
		return NewMockEmailSender()
	default:
		log.Printf("[Provider] Unknown PROVIDER_MODE=%q, defaulting to SIMULATED", mode)
		return NewMockEmailSender()
	}
}

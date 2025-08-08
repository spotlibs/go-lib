package ctx

import (
	"crypto/rand"
	"fmt"
	"time"
)

func GenerateTimeBasedID() string {
	// Get current time with microsecond precision
	now := time.Now()
	timestamp := now.UnixMicro()

	// Generate random component for uniqueness
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return "0000"
	}
	randomHex := fmt.Sprintf("%x", randomBytes)

	// Create the full ID (timestamp + random component)
	id := fmt.Sprintf("%d%s", timestamp, randomHex)

	return id
}

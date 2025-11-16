package shared

import (
	"os"
	"time"
)

type AnalyticsEvent struct {
	EventType    string    `json:"event_type"`
	Page         string    `json:"page"`
	Timestamp    time.Time `json:"timestamp"`
	SessionID    string    `json:"session_id"`
	UserAgent    string    `json:"user_agent,omitempty"`
	ScreenWidth  int       `json:"screen_width,omitempty"`
	ScreenHeight int       `json:"screen_height,omitempty"`
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

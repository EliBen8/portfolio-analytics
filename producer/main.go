package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq"

	. "shared"
)

var db *sql.DB
var kafkaProducer sarama.SyncProducer

func createTable() {
	query := `
    CREATE TABLE IF NOT EXISTS analytics_events (
        id SERIAL PRIMARY KEY,
        event_type VARCHAR(50) NOT NULL,
        page VARCHAR(255) NOT NULL,
        timestamp TIMESTAMP NOT NULL,
        session_id VARCHAR(100) NOT NULL,
        user_agent TEXT,
        screen_width INTEGER,
        screen_height INTEGER,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	log.Println("✅ Database table ready!")
}

func main() {
	// Connect to Kafka
	var err error
	kafkaProducer, err = NewKafkaProducer()
	if err != nil {
		log.Fatal("Failed to create Kafka producer:", err)
	}
	defer kafkaProducer.Close()

	// Connect to the database
	connStr := GetEnv("DATABASE_URL", "postgres://localhost/portfolio_analytics?sslmode=disable")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test the connection
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("✅ Connected to database!")
			break
		}
		log.Printf("Failed to ping database: %v", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("Could not connect to database after retries")
	}

	createTable()

	// Routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/api/analytics", handleAnalytics)
	http.HandleFunc("/api/health", handleHealth)
	http.HandleFunc("/api/stats", handleStats)

	// Start server
	port := GetEnv("PORT", "8080")
	fmt.Printf("Server starting on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Home route
func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Portfolio Analytics API",
		"version": "1.0.0",
		"status":  "running",
	})
}

// Health route
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check DB connection
	err := db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":   "healthy",
		"database": "connected",
	})
}

// Analytics endpoints
func handleAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON body
	var event AnalyticsEvent
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Add user agent if not provided
	if event.UserAgent == "" {
		event.UserAgent = r.Header.Get("User-Agent")
	}

	// Marshal event to JSON for Kafka
	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Publish to Kafka instead of DB
	message := &sarama.ProducerMessage{
		Topic: "portfolio-analytics-events",
		Value: sarama.ByteEncoder(eventBytes),
	}

	partition, offset, err := kafkaProducer.SendMessage(message)
	if err != nil {
		log.Printf("Failed to send to Kafka: %v", err)
		http.Error(w, "Failed to record event", http.StatusInternalServerError)
		return
	}

	log.Printf("📤 Event sent to Kafka! Partition: %d, Offset: %d, Type: %s, Page: %s",
		partition, offset, event.EventType, event.Page)

	// Send success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"message":   "Event queued",
		"partition": partition,
		"offset":    offset,
	})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get the total events
	var totalEvents int
	err := db.QueryRow("SELECT COUNT(*) FROM analytics_events").Scan(&totalEvents)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get events by type
	rows, err := db.Query(`
		SELECT event_type, COUNT(*) as count
		FROM analytics_events
		GROUP BY event_type
	`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	eventsByType := make(map[string]int)
	for rows.Next() {
		var eventType string
		var count int
		rows.Scan(&eventType, &count)
		eventsByType[eventType] = count
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_events":   totalEvents,
		"events_by_type": eventsByType,
	})
}

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq"

	. "shared"
)

var consumerDB *sql.DB

func main() {
	// Connect to database
	var err error
	connStr := GetEnv("DATABASE_URL", "postgres://localhost/portfolio_analytics?sslmode=disable")
	consumerDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer consumerDB.Close()

	// Test connection with retries
	for i := 0; i < 10; i++ {
		err = consumerDB.Ping()
		if err == nil {
			log.Println("✅ Consumer connected to database!")
			break
		}
		log.Printf("Database ping failed (attempt %d/10): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to database after retries:", err)
	}

	// Create Kafka consumer
	consumer, err := NewKafkaConsumer()
	if err != nil {
		log.Fatal("Failed to create Kafka consumer:", err)
	}
	defer consumer.Close()

	// Subscribe to topic
	partitionConsumer, err := consumer.ConsumePartition("portfolio-analytics-events", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatal("Failed to start partition consumer:", err)
	}
	defer partitionConsumer.Close()

	log.Println("🎧 Consumer is listening for events...")

	// Handle graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Consume messages
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			handleKafkaMessage(msg)
		case err := <-partitionConsumer.Errors():
			log.Printf("Error: %v", err)
		case <-signals:
			log.Println("Shutting down consumer...")
			return
		}
	}
}

func handleKafkaMessage(msg *sarama.ConsumerMessage) {
	var event AnalyticsEvent
	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		return
	}

	// Insert into database
	query := `
		INSERT INTO analytics_events
		(event_type, page, session_id, user_agent, screen_width, screen_height, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id int
	err = consumerDB.QueryRow(
		query,
		event.EventType,
		event.Page,
		event.SessionID,
		event.UserAgent,
		event.ScreenWidth,
		event.ScreenHeight,
		event.Timestamp,
	).Scan(&id)

	if err != nil {
		log.Printf("❌ Database error: %v", err)
		return
	}

	log.Printf("💾 Event saved to DB! ID: %d, Type: %s, Page: %s (Offset: %d)",
		id, event.EventType, event.Page, msg.Offset)
}

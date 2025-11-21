# Portfolio Analytics Backend

Real-time analytics system built with microservices architecture and event-driven design using Kafka for reliable event streaming.

## 🏗️ Architecture
```
┌─────────────┐      ┌──────────┐      ┌──────────┐      ┌────────────┐
│   Frontend  │─────▶│ Producer │─────▶│  Kafka   │─────▶│  Consumer  │
│  (Website)  │      │   (API)  │      │(Redpanda)│      │  (Worker)  │
└─────────────┘      └──────────┘      └──────────┘      └────────────┘
                           │                                     │
                           │                                     │
                           └─────────────────┬───────────────────┘
                                             ▼
                                      ┌──────────────┐
                                      │  PostgreSQL  │
                                      └──────────────┘
```

### Components

- **Producer**: HTTP API server that receives analytics events and publishes them to Kafka
- **Consumer**: Background worker that consumes events from Kafka and persists them to PostgreSQL
- **Shared**: Common types, utilities, and Kafka configuration used by both services
- **Kafka (Redpanda)**: Message queue for reliable, scalable event streaming
- **PostgreSQL**: Data persistence layer

## 🚀 Features

- ✅ Event-driven architecture with Kafka
- ✅ Horizontal scalability (services scale independently)
- ✅ Reliable message delivery and processing
- ✅ Real-time analytics tracking
- ✅ GDPR compliant (consent-based tracking)
- ✅ RESTful API for event ingestion and stats retrieval
- ✅ Production deployment on Railway

## 📊 Tracked Events

- **Page Views**: Track user navigation across the site
- **Button Clicks**: Monitor user interactions
- **Time Spent**: Measure engagement duration
- **Navigation**: Track internal link clicks

## 🛠️ Tech Stack

- **Language**: Go 1.25.4
- **Message Queue**: Kafka (Redpanda Cloud)
- **Database**: PostgreSQL
- **Framework**: Standard library (net/http)
- **Libraries**:
  - `github.com/IBM/sarama` - Kafka client
  - `github.com/lib/pq` - PostgreSQL driver
  - `github.com/xdg-go/scram` - SASL authentication

## 📁 Project Structure
```
portfolio-analytics/
├── producer/          # HTTP API service
│   ├── main.go       # API routes and handlers
│   ├── go.mod
│   └── go.sum
├── consumer/          # Background worker
│   ├── main.go       # Kafka consumer and DB writer
│   ├── go.mod
│   └── go.sum
├── shared/            # Shared code
│   ├── types.go      # Common data structures
│   ├── kafka.go      # Kafka configuration
│   ├── go.mod
│   └── go.sum
├── go.work           # Go workspace configuration
└── README.md
```

## 🔧 Local Development

### Prerequisites

- Go 1.25.4+
- PostgreSQL 14+
- Kafka/Redpanda instance

### Setup

1. **Clone the repository**
```bash
git clone https://github.com/EliBen8/portfolio-analytics.git
cd portfolio-analytics
```

2. **Set environment variables**
```bash
export DATABASE_URL="postgres://localhost/portfolio_analytics?sslmode=disable"
```

3. **Install dependencies**
```bash
go work sync
cd producer && go mod tidy
cd ../consumer && go mod tidy
cd ../shared && go mod tidy
```

4. **Run the services**

Terminal 1 (Producer):
```bash
cd producer
go run main.go
```

Terminal 2 (Consumer):
```bash
cd consumer
go run main.go
```

## 🚢 Deployment

Deployed on Railway with the following configuration:

### Producer Service
```bash
# Build Command
go build -o producer_bin ./producer

# Start Command
./producer_bin
```

### Consumer Service
```bash
# Build Command
go build -o consumer_bin ./consumer

# Start Command
./consumer_bin
```

### Environment Variables
- `DATABASE_URL`: PostgreSQL connection string
- Kafka credentials configured in `shared/kafka.go`

## 📡 API Endpoints

### POST `/api/analytics`
Track an analytics event

**Request:**
```json
{
  "event_type": "page_view",
  "page": "/",
  "session_id": "session-123",
  "timestamp": "2025-11-16T05:38:00Z",
  "screen_width": 1920,
  "screen_height": 1080
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Event queued",
  "partition": 0,
  "offset": 42
}
```

### GET `/api/stats`
Retrieve analytics statistics

**Response:**
```json
{
  "total_events": 150,
  "events_by_type": {
    "page_view": 75,
    "click": 60,
    "time_spent": 15
  }
}
```

### GET `/api/health`
Health check endpoint

**Response:**
```json
{
  "status": "healthy",
  "database": "connected"
}
```

## 🔐 Security

- SASL/SCRAM-SHA-256 authentication for Kafka
- TLS encryption for Kafka connections
- Environment-based configuration (no hardcoded credentials)
- CORS configured for frontend domain

## 📈 Performance

- **Reliability**: Guaranteed message delivery with Kafka
- **Scalability**: Horizontally scalable (add more consumer instances)

## 🤝 Contributing

This is a personal project, but feedback and suggestions are welcome!

## 👤 Author

**Eli Bendavid**
- GitHub: [@EliBen8](https://github.com/EliBen8)
- Portfolio: [eliben8.github.io](https://eliben8.github.io)
- Email: elirbendavid@gmail.com

## 🙏 Acknowledgments

- Built as part of my portfolio to demonstrate microservices architecture
- Inspired by modern event-driven design patterns
- Uses Redpanda Cloud for managed Kafka
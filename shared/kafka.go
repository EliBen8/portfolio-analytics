package shared

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
)

var SHA256 = scram.SHA256

func sha256HashGenerator() scram.HashGeneratorFcn {
	return scram.SHA256
}

type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}

func (x *XDGSCRAMClient) Step(challenge string) (response string, err error) {
	response, err = x.ClientConversation.Step(challenge)
	return
}

func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}

func NewKafkaProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	// SASL authentication - FROM ENVIRONMENT VARIABLES
	config.Net.SASL.Enable = true
	config.Net.SASL.User = os.Getenv("KAFKA_USER")
	config.Net.SASL.Password = os.Getenv("KAFKA_PASSWORD")
	config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
		return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
	}

	// TLS configuration
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = &tls.Config{
		InsecureSkipVerify: false,
	}

	brokers := []string{os.Getenv("KAFKA_BROKER")}
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Printf("Failed to create Kafka producer: %v", err)
		return nil, err
	}

	log.Println("✅ Connected to Kafka!")
	return producer, nil
}

func NewKafkaConsumer() (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// SASL authentication - FROM ENVIRONMENT VARIABLES
	config.Net.SASL.Enable = true
	config.Net.SASL.User = os.Getenv("KAFKA_USER")
	config.Net.SASL.Password = os.Getenv("KAFKA_PASSWORD")
	config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
		return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
	}

	// TLS configuration
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = &tls.Config{
		InsecureSkipVerify: false,
	}

	brokers := []string{os.Getenv("KAFKA_BROKER")}
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Printf("Failed to create Kafka consumer: %v", err)
		return nil, err
	}

	log.Println("✅ Connected to Kafka as consumer!")
	return consumer, nil
}

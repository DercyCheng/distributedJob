package kafka

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

// MessageHandler is a function that processes Kafka messages
type MessageHandler func(msg *sarama.ConsumerMessage) error

// Producer handles message production to Kafka
type Producer struct {
	producer sarama.SyncProducer
	brokers  []string
}

// Consumer handles message consumption from Kafka
type Consumer struct {
	consumer      sarama.ConsumerGroup
	topics        []string
	handler       MessageHandler
	consumerGroup string
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// Manager handles Kafka operations
type Manager struct {
	producer *Producer
	consumer *Consumer
	brokers  []string
}

// NewManager creates a new Kafka manager
func NewManager(brokers []string) *Manager {
	return &Manager{
		brokers: brokers,
	}
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &Producer{
		producer: producer,
		brokers:  brokers,
	}, nil
}

// SendMessage sends a message to a Kafka topic
func (p *Producer) SendMessage(topic string, key []byte, value []byte) (int32, int64, error) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	return p.producer.SendMessage(msg)
}

// Close closes the producer connection
func (p *Producer) Close() error {
	return p.producer.Close()
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, topics []string, groupID string, handler MessageHandler) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		consumer:      consumerGroup,
		topics:        topics,
		handler:       handler,
		consumerGroup: groupID,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Start starts consuming messages
func (c *Consumer) Start() error {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.consumer.Consume(c.ctx, c.topics, &consumerHandler{handler: c.handler}); err != nil {
				log.Printf("Error from consumer: %v", err)
			}

			if c.ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

// Stop stops the consumer
func (c *Consumer) Stop() error {
	c.cancel()
	c.wg.Wait()
	return c.consumer.Close()
}

// consumerHandler implements sarama.ConsumerGroupHandler
type consumerHandler struct {
	handler MessageHandler
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (ch *consumerHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (ch *consumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (ch *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if err := ch.handler(message); err != nil {
			log.Printf("Error handling message: %v", err)
		}
		session.MarkMessage(message, "")
	}
	return nil
}

// InitializeProducer initializes the Kafka producer
func (m *Manager) InitializeProducer() error {
	producer, err := NewProducer(m.brokers)
	if err != nil {
		return err
	}
	m.producer = producer
	return nil
}

// InitializeConsumer initializes the Kafka consumer
func (m *Manager) InitializeConsumer(topics []string, groupID string, handler MessageHandler) error {
	consumer, err := NewConsumer(m.brokers, topics, groupID, handler)
	if err != nil {
		return err
	}
	m.consumer = consumer
	return nil
}

// StartConsumer starts the consumer
func (m *Manager) StartConsumer() error {
	if m.consumer == nil {
		return fmt.Errorf("consumer not initialized")
	}
	return m.consumer.Start()
}

// GetProducer returns the producer instance
func (m *Manager) GetProducer() *Producer {
	return m.producer
}

// Close closes both producer and consumer connections
func (m *Manager) Close() error {
	var producerErr, consumerErr error

	if m.producer != nil {
		producerErr = m.producer.Close()
	}

	if m.consumer != nil {
		consumerErr = m.consumer.Stop()
	}

	if producerErr != nil {
		return producerErr
	}
	return consumerErr
}

package kafka

import (
	"context"
	"crypto/sha1"
	"log"
	"os"
	"strings"
	"time"

	"restapi/logger"
	helpers "restapi/util"

	"github.com/IBM/sarama"
)

type ConsumerGroup struct {
	Client sarama.ConsumerGroup
}

type Consumer struct {
	Ready     chan bool
	Processor Processor
}
type Params struct {
	SessionTimeout    time.Duration
	RebalanceTimeout  time.Duration
	HeartBeatInterval time.Duration
	OffsetInitial     int64
}

func NewConsumer(ready chan bool, processor Processor) Consumer {
	return Consumer{
		Ready:     ready,
		Processor: processor,
	}
}

func NewConsumerConfig(env string, prefix string, param Params) *sarama.Config {
	// load env if not already loaded
	if len(os.Getenv(prefix+"_KAFKA_VERSION")) == 0 {
		helpers.LoadEnv(env)
	}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer is initialized.
	 */

	version, err := sarama.ParseKafkaVersion(os.Getenv(prefix + "_KAFKA_VERSION"))
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	config := sarama.NewConfig()
	config.Version = version

	/**
	 * ChannelBufferSize is set to 1, so that session timeout is applicable for
	 * processing of one message only (to be extra sure)
	 */
	config.ChannelBufferSize = 1
	/**
	 * Session Timeout is set to 5 minutes (high value) because sarama does not send heartbeats
	 * indicating the liveliness of consumer while rebalancing and consumption of last message on a partition
	 * might take time because consumer might have to wait for jdUrl
	 * (the aforementioned scenario caused frequent generation of new consumers during each rebalancing)
	 */

	// Setting Default safe values
	config.Consumer.Group.Session.Timeout = 1 * time.Minute
	config.Consumer.Group.Rebalance.Timeout = 1 * time.Minute
	config.Consumer.Group.Heartbeat.Interval = 10 * time.Second

	if param.SessionTimeout != 0 {
		config.Consumer.Group.Session.Timeout = param.SessionTimeout
	}

	// By definition Rebalance Timeout should at least be equal to Session Timeout or greater
	if param.RebalanceTimeout != 0 {
		config.Consumer.Group.Rebalance.Timeout = param.RebalanceTimeout
	}

	if param.HeartBeatInterval != 0 {
		config.Consumer.Group.Heartbeat.Interval = param.HeartBeatInterval
	}

	if param.OffsetInitial != 0 {
		config.Consumer.Offsets.Initial = param.OffsetInitial
	}

	return config
}

func NewConsumerGroup(env string, prefix string, param Params) *ConsumerGroup {

	instance := &ConsumerGroup{}

	// load env if not already loaded
	if len(os.Getenv(prefix+"_KAFKA_BROKER")) == 0 {
		helpers.LoadEnv(env)
	}

	config := NewConsumerConfig(env, prefix, param)

	brokers := os.Getenv(prefix + "_KAFKA_BROKER")
	group := os.Getenv(prefix + "_KAFKA_GROUP")

	logger.Info(nil, "Setting up new kafka consumergroup", logger.Z{"group": group, "brokers": brokers})

	client, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), group, config)
	if err != nil {
		log.Fatalf("Could not set up kafka consumer group ERR: %v", err)
	}
	instance.Client = client
	logger.Info(nil, "Setup new kafka consumergroup successfully!!", nil)
	return instance
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	logger.Info(nil, "Setting up Kafka Consumer", nil)
	if consumer.Processor == nil {
		log.Panic("Please Define processor for this consumer")
	}
	close(consumer.Ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {

		tid := hash(string(message.Value))
		ctx := context.WithValue(context.Background(), logger.TransactionIDKey, tid)

		logData := logData{
			Message:          string(message.Value),
			Topic:            message.Topic,
			Partition:        message.Partition,
			Offset:           message.Offset,
			MessageTimestamp: &message.Timestamp,
		}

		err := consumer.Processor.Process(ctx, string(message.Value), message.Timestamp, message.Topic)
		if err != nil {
			logData.Error = err.Error()
			logger.Error(ctx, "Could not Process message", logger.Z{"log_data": logData})
		}

		session.MarkMessage(message, "")
	}

	return nil
}

func hash(s string) []byte {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return bs
}

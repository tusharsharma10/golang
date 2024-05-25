package kafka

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"restapi/helpers"

	"github.com/IBM/sarama"
)

type Processor interface {
	Process(context.Context, string, time.Time, string) error
}

type logData struct {
	Topic            string     `json:"topic,omitempty"`
	Message          string     `json:"message,omitempty"`
	Offset           int64      `json:"offset,omitempty"`
	MessageTimestamp *time.Time `json:"message_timestamp,omitempty"`
	Partition        int32      `json:"partition,omitempty"`
	Error            string     `json:"error,omitempty"`
}

type ConfigParams struct {
	Version      sarama.KafkaVersion
	Partitioner  sarama.PartitionerConstructor
	RequiredAcks sarama.RequiredAcks
	Successes    bool
}
type Producer struct {
	Client sarama.SyncProducer
	Topic  string
}

func NewProducer(
	env string,
	prefix string,
	configParams ConfigParams,
) (*Producer, error) {
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
		return nil, fmt.Errorf("error parsing Kafka version: %w", err)
	}

	config := sarama.NewConfig()

	// Setting Default Values for config
	config.Version = version
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = true

	if configParams.Version.String() != "" {
		config.Version = configParams.Version
	}

	if configParams.Partitioner != nil {
		config.Producer.Partitioner = configParams.Partitioner
	}

	if configParams.RequiredAcks != 0 {
		config.Producer.RequiredAcks = configParams.RequiredAcks
	}

	if configParams.Successes {
		config.Producer.Return.Successes = configParams.Successes
	}

	brokers := os.Getenv(prefix + "_KAFKA_BROKER")

	producer, err := sarama.NewSyncProducer(strings.Split(brokers, ","), config)

	if err != nil {
		return nil, fmt.Errorf("error intializing Kafka producer: %w", err)
	}

	return &Producer{Client: producer, Topic: os.Getenv(prefix + "_KAFKA_TOPIC")}, nil
}

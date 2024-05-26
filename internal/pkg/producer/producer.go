package producer

import (
	"TransportLayer/internal/models"
	"TransportLayer/internal/utils"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
)

func Produce(segment models.MessageChannel) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	// создание producer-а
	producer, err := sarama.NewSyncProducer([]string{utils.KafkaAddress}, config)
	if err != nil {
		return fmt.Errorf("error creating producer: %w", err)
	}
	defer producer.Close()

	segmentString, _ := json.Marshal(segment)
	message := &sarama.ProducerMessage{
		Topic: "segments",
		Value: sarama.StringEncoder(segmentString),
	}

	_, _, err = producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

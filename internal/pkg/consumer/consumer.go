package consumer

import (
	"TransportLayer/internal/models"
	"TransportLayer/internal/utils"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"sync"
	"time"
)

type Storage map[int]models.Message
type sendFunc func(body models.MessageApp)

var storage = Storage{}

func addMessage(segment models.MessageChannel) {
	storage[segment.MessageId] = models.Message{
		Received: 0,
		Total:    segment.TotalSegments,
		Last:     time.Now().UTC(),
		Username: segment.Sender,
		Date:     segment.Date,
		Segments: make([][]byte, segment.TotalSegments), // заранее выделяем память, это важно!
	}
}

func AddSegment(segment models.MessageChannel) {
	// используем мьютекс, чтобы избежать конкуретного доступа к хранилищу
	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	// если это первый сегмент сообщения, создаем пустое сообщение
	id := segment.MessageId
	_, found := storage[id]
	if !found {
		addMessage(segment)
	}

	// добавляем в сообщение информацию о сегменте
	message, _ := storage[id]
	message.Received++
	message.Last = time.Now().UTC()
	message.Segments[segment.NumOfSegment-1] = segment.Message // сохраняем правильный порядок сегментов
	storage[id] = message
	//fmt.Println(storage[id])
}

func getMessageText(messageId int) string {
	result := ""
	message, _ := storage[messageId]
	for _, segment := range message.Segments {
		result += string(segment[:])
		fmt.Println(messageId, segment)
	}
	return result
}

func ScanStorage(sender sendFunc) {
	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	payload := models.MessageApp{}
	for id, message := range storage {
		if message.Received == message.Total {
			payload = models.MessageApp{
				Sender:  message.Username,
				Content: getMessageText(id),
				Date:    message.Date,
				IsError: false,
			}
			delete(storage, id)
			fmt.Println("f", payload)
			go sender(payload)
		} else if time.Since(message.Last) > utils.N*time.Second+time.Second {
			payload = models.MessageApp{
				Sender:  message.Username,
				Content: "",
				Date:    message.Date,
				IsError: true, // ошибка
			}
			delete(storage, id)
			fmt.Printf("sent error: %+v\n", payload)
			go sender(payload) // запускаем горутину с отправкой на прикладной уровень, не будем дожидаться результата ее выполнения
		}
	}
	// storage = make(Storage)
}

func ReadFromKafka() error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{utils.KafkaAddress}, config)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error creating consumer: %w", err)
	}
	defer consumer.Close()

	// подключение consumer-а к топика
	partitionConsumer, err := consumer.ConsumePartition("segments", 0, sarama.OffsetNewest)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error opening topic: %w", err)
	}
	defer partitionConsumer.Close()

	for {
		select {
		case message := <-partitionConsumer.Messages():
			segment := models.MessageChannel{}
			if err := json.Unmarshal(message.Value, &segment); err != nil {
				fmt.Printf("Error reading from kafka: %v", err)
			}
			fmt.Printf("%+v\n", segment)
			AddSegment(segment)
		case err := <-partitionConsumer.Errors():
			fmt.Printf("Error: %s\n", err.Error())
		}
	}
}

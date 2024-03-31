package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	x              = 100
	n              = 2
	channelAddress = "http://localhost:8081"
	frontAddress   = "http://localhost:3000"
)

type MessageApp struct {
	Sender  string
	Content string
	Date    time.Time
}

type MessageChannel struct {
	NumOfSegment  int
	TotalSegments int
	Message       []byte
}

// @title Transport Layer
// @version 1.0
// @description Emulation of transport layer

// @host localhost:8080
// @schemes http
// @BasePath /
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/send", SendMessage).Methods("POST")
	r.HandleFunc("/transfer", TransferMessage).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go consume()
	log.Fatal(srv.ListenAndServe())
}

// TransferMessage godoc
// @Summary      Transfer message
// @Description  Transfer message to app layer through transport layer
// @Accept       json
// @Param        message  body  MessageChannel  true  "Decoded message from channel layer"
// @Success      200
// @Failure      400
// @Router       /transfer [post]
func TransferMessage(writer http.ResponseWriter, request *http.Request) {
	bytesMessage, _ := io.ReadAll(request.Body)
	fmt.Println(bytesMessage)
	err := produce(bytesMessage)
	if err != nil {
		return
	}
}

func consume() {
	topic := "my-topic"
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max

	b := make([]byte, 10e3)
	for {
		time.Sleep(n * time.Second)
		n, err := batch.Read(b)
		if err != nil {
			break
		}
		fmt.Println(string(b[:n]))
	}

	if err := batch.Close(); err != nil {
		log.Fatal("failed to close batch:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close connection:", err)
	}
}

func produce(message []byte) error {
	topic := "my-topic"
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(kafka.Message{Value: message})
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
	return err
}

// SendMessage godoc
// @Summary      Receive message from app layer
// @Description  Receive message from app layer and broke o segments for channel layer
// @Accept       json
// @Param        message  body  MessageApp  true  "Message from chat"
// @Success      200
// @Failure      400
// @Router       /send [post]
func SendMessage(writer http.ResponseWriter, request *http.Request) {
	bytesMessage, _ := io.ReadAll(request.Body)
	fmt.Println(bytesMessage)
	total := len(bytesMessage) / x
	for i := 0; i < len(bytesMessage); i += x {
		if err := sendToChannelLayer(bytesMessage[i:min(len(bytesMessage), i+x)], i, total); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

	}
	writer.WriteHeader(http.StatusOK)
}

func sendToChannelLayer(message []byte, numMessage, total int) error {
	bodyReader := bytes.NewReader(message)
	req, err := http.NewRequest(http.MethodPost, channelAddress, bodyReader)

	if err != nil {
		return err
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	_, err = client.Do(req)
	return err
}

package api

import (
	"TransportLayer/internal/models"
	"TransportLayer/internal/pkg/producer"
	"TransportLayer/internal/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func splitContent(content []byte) [][]byte {
	total := (len(content) + utils.X - 1) / utils.X
	fmt.Println(total)
	result := make([][]byte, total)
	for i := 0; i < total; i++ {
		fmt.Println(i*utils.X, i*utils.X+utils.X, content[i*utils.X:min(len(content), i*utils.X+utils.X)])
		result[i] = content[i*utils.X : min(len(content), i*utils.X+utils.X)]
	}
	return result
}

func sendToChannelLayer(message models.MessageChannel) {
	reqBody, _ := json.Marshal(message)

	req, _ := http.NewRequest("POST", utils.ChannelAddress, bytes.NewBuffer(reqBody))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
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
	message := &models.MessageApp{}
	err := json.Unmarshal(bytesMessage, message)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
	}

	writer.WriteHeader(http.StatusOK)

	contentBytes := []byte(message.Content)
	fmt.Println(contentBytes)
	segments := splitContent(contentBytes)
	fmt.Println(segments)
	for i, segment := range segments {
		payload := models.MessageChannel{
			NumOfSegment:  i + 1,
			TotalSegments: len(segments),
			Message:       segment,
			Date:          message.Date,
			Sender:        message.Sender,
			MessageId:     models.MessageId + 1,
		}
		go sendToChannelLayer(payload)
		// fmt.Printf("sent segment: %+v\n", payload)
	}
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
	body, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	segment := models.MessageChannel{}
	if err = json.Unmarshal(body, &segment); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = producer.Produce(segment); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func SendReceiveRequest(body models.MessageApp) {
	fmt.Println("OOOH")
	reqBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", utils.FrontAddress, bytes.NewBuffer(reqBody))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
}

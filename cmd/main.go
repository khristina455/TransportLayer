package main

import (
	"TransportLayer/internal/pkg/api"
	"TransportLayer/internal/pkg/consumer"
	"TransportLayer/internal/utils"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

// @title Transport Layer
// @version 1.0
// @description Emulation of transport layer

// @host localhost:8083
// @schemes http
// @BasePath /
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/send", api.SendMessage).Methods("POST")
	r.HandleFunc("/transfer", api.TransferMessage).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8083",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go consumer.ReadFromKafka()
	go func() {
		ticker := time.NewTicker(utils.N * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				consumer.ScanStorage(api.SendReceiveRequest)
			}
		}
	}()
	log.Fatal(srv.ListenAndServe())
}

package models

import "time"

var MessageId int

type MessageApp struct {
	Sender  string    `json:"sender"`
	Content string    `json:"content"`
	Date    time.Time `json:"date"`
	IsError bool      `json:"isError"`
}

type MessageChannel struct {
	NumOfSegment  int       `json:"numOfSegment"`
	TotalSegments int       `json:"totalSegments"`
	Message       []byte    `json:"message"`
	Date          time.Time `json:"sendDate"`
	Sender        string    `json:"sender"`
	MessageId     int       `json:"messageId"`
}

type Message struct {
	Received int
	Total    int
	IsUsed   bool
	Last     time.Time
	Username string
	Date     time.Time
	Segments [][]byte
}

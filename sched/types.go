package sched

import (
	"encoding/json"
	"time"
)

type Message struct {
	Id      string `json:"id"`
	Type    int    `json:"type"`
	Payload []byte `json:"payload"`
}

type DelayMessage struct {
	Message
	Time time.Time `json:"time"`
}

func toJson(message Message) []byte {
	b, _ := json.Marshal(message)
	return b
}

func fromJson(s string) Message {
	var m Message
	_ = json.Unmarshal([]byte(s), &m)
	return m
}

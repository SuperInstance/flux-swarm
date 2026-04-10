package main

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of A2A message
type MessageType string

const (
	TELL       MessageType = "TELL"
	ASK        MessageType = "ASK"
	DELEGATE   MessageType = "DELEGATE"
	BROADCAST  MessageType = "BROADCAST"
)

// A2AMessage represents a message between agents
type A2AMessage struct {
	SenderID   string      `json:"sender_id"`
	ReceiverID string      `json:"receiver_id"`
	Type       MessageType `json:"type"`
	Payload    string      `json:"payload"`
	Trust      float64     `json:"trust"`
	Timestamp  time.Time   `json:"timestamp"`
}

// Serialize converts the message to JSON
func (m *A2AMessage) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// DeserializeMessage parses JSON into an A2AMessage
func DeserializeMessage(data []byte) (*A2AMessage, error) {
	var msg A2AMessage
	err := json.Unmarshal(data, &msg)
	return &msg, err
}

package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// ========== A2A MESSAGE SERIALIZATION TESTS ==========

func TestMessageSerialize(t *testing.T) {
	msg := A2AMessage{
		SenderID:   "agent-001",
		ReceiverID: "agent-002",
		Type:       TELL,
		Payload:    "hello world",
		Trust:      1.5,
		Timestamp:  time.Now().UTC().Truncate(time.Second),
	}

	data, err := msg.Serialize()
	if err != nil {
		t.Fatalf("Serialize() failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Serialize() returned empty data")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Serialized data is not valid JSON: %v", err)
	}

	if parsed["sender_id"] != "agent-001" {
		t.Errorf("Expected sender_id 'agent-001', got %v", parsed["sender_id"])
	}
	if parsed["type"] != "TELL" {
		t.Errorf("Expected type 'TELL', got %v", parsed["type"])
	}
	if parsed["payload"] != "hello world" {
		t.Errorf("Expected payload 'hello world', got %v", parsed["payload"])
	}
}

func TestMessageSerializeRoundtrip(t *testing.T) {
	original := A2AMessage{
		SenderID:   "agent-A",
		ReceiverID: "agent-B",
		Type:       DELEGATE,
		Payload:    "do task X",
		Trust:      2.7,
		Timestamp:  time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() failed: %v", err)
	}

	recovered, err := DeserializeMessage(data)
	if err != nil {
		t.Fatalf("DeserializeMessage() failed: %v", err)
	}

	if recovered.SenderID != original.SenderID {
		t.Errorf("SenderID mismatch: got %s, want %s", recovered.SenderID, original.SenderID)
	}
	if recovered.ReceiverID != original.ReceiverID {
		t.Errorf("ReceiverID mismatch: got %s, want %s", recovered.ReceiverID, original.ReceiverID)
	}
	if recovered.Type != original.Type {
		t.Errorf("Type mismatch: got %s, want %s", recovered.Type, original.Type)
	}
	if recovered.Payload != original.Payload {
		t.Errorf("Payload mismatch: got %s, want %s", recovered.Payload, original.Payload)
	}
	if recovered.Trust != original.Trust {
		t.Errorf("Trust mismatch: got %f, want %f", recovered.Trust, original.Trust)
	}
}

func TestDeserializeMessage(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name:    "valid message",
			data:    `{"sender_id":"a","receiver_id":"b","type":"TELL","payload":"hi","trust":1.0,"timestamp":"2025-01-01T00:00:00Z"}`,
			wantErr: false,
		},
		{
			name:    "empty payload",
			data:    `{"sender_id":"a","receiver_id":"b","type":"ASK","payload":"","trust":0.5,"timestamp":"2025-01-01T00:00:00Z"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    `{not valid json}`,
			wantErr: true,
		},
		{
			name:    "empty bytes",
			data:    ``,
			wantErr: true,
		},
		{
			name:    "non-JSON bytes",
			data:    `\x00\x01\x02`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := DeserializeMessage([]byte(tt.data))
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if msg == nil {
				t.Fatal("Expected non-nil message")
			}
		})
	}
}

func TestMessageTypes(t *testing.T) {
	types := []MessageType{TELL, ASK, DELEGATE, BROADCAST}
	expected := []string{"TELL", "ASK", "DELEGATE", "BROADCAST"}

	for i, mt := range types {
		if string(mt) != expected[i] {
			t.Errorf("MessageType[%d] = %q, want %q", i, mt, expected[i])
		}
	}
}

func TestMessageEmptyFieldsSerialize(t *testing.T) {
	msg := A2AMessage{}
	data, err := msg.Serialize()
	if err != nil {
		t.Fatalf("Serialize() failed for empty message: %v", err)
	}

	// Verify JSON contains empty string fields
	str := string(data)
	if !strings.Contains(str, `"sender_id":""`) {
		t.Error("Expected empty sender_id in serialized empty message")
	}
}

func TestSendMessageFields(t *testing.T) {
	agent := NewAgent(Worker, []byte{HALT})
	msg := agent.SendMessage("dest-agent", ASK, "status?")

	if msg.SenderID != agent.ID {
		t.Errorf("Expected sender %s, got %s", agent.ID, msg.SenderID)
	}
	if msg.ReceiverID != "dest-agent" {
		t.Errorf("Expected receiver 'dest-agent', got %s", msg.ReceiverID)
	}
	if msg.Type != ASK {
		t.Errorf("Expected type ASK, got %s", msg.Type)
	}
	if msg.Payload != "status?" {
		t.Errorf("Expected payload 'status?', got %s", msg.Payload)
	}
	if msg.Trust != 1.0 {
		t.Errorf("Expected trust 1.0, got %f", msg.Trust)
	}
	if msg.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

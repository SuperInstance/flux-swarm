package main

import "testing"

// ========== AGENT TESTS ==========

func TestAgentProcess(t *testing.T) {
	tests := []struct {
		name      string
		bytecode  []byte
		wantErr   bool
		wantPanic bool
	}{
		{
			name:     "halt immediately",
			bytecode: []byte{HALT},
			wantErr:  false,
		},
		{
			name:     "single movi then halt",
			bytecode: []byte{MOVI, 0, 42, HALT},
			wantErr:  false,
		},
		{
			name:     "not running returns nil",
			bytecode: []byte{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(Worker, tt.bytecode)
			if !tt.wantPanic {
				// If bytecode is empty, the VM starts running but will halt on first step
				// Actually if bytecode is empty, PC >= len(Bytecode) so ExecuteStep errors
				if len(tt.bytecode) == 0 {
					agent.VM.Running = false
				}
				err := agent.Process()
				if tt.wantErr && err == nil {
					t.Error("Expected error, got nil")
				}
				if !tt.wantErr && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAgentGetNextMessage(t *testing.T) {
	agent := NewAgent(Scout, []byte{HALT})

	// Empty queue returns nil
	msg := agent.GetNextMessage()
	if msg != nil {
		t.Error("Expected nil from empty queue, got non-nil")
	}

	// Add messages and consume them
	agent.ReceiveMessage(A2AMessage{Payload: "msg1"})
	agent.ReceiveMessage(A2AMessage{Payload: "msg2"})

	msg1 := agent.GetNextMessage()
	if msg1 == nil || msg1.Payload != "msg1" {
		t.Errorf("Expected 'msg1', got %v", msg1)
	}

	msg2 := agent.GetNextMessage()
	if msg2 == nil || msg2.Payload != "msg2" {
		t.Errorf("Expected 'msg2', got %v", msg2)
	}

	// Queue should be empty now
	msg3 := agent.GetNextMessage()
	if msg3 != nil {
		t.Error("Expected nil after consuming all messages")
	}
}

func TestAgentGetResult(t *testing.T) {
	agent := NewAgent(Specialist, []byte{MOVI, 0, 99, MOVI, 5, 42, HALT})

	// Execute all steps
	for agent.VM.Running {
		agent.Process()
	}

	result := agent.GetResult()
	if result[0] != 99 {
		t.Errorf("Expected R0=99, got %d", result[0])
	}
	if result[5] != 42 {
		t.Errorf("Expected R5=42, got %d", result[5])
	}
}

func TestAgentRoles(t *testing.T) {
	tests := []struct {
		role    Role
		wantStr string
	}{
		{Worker, "worker"},
		{Scout, "scout"},
		{Coordinator, "coordinator"},
		{Specialist, "specialist"},
	}

	for _, tt := range tests {
		agent := NewAgent(tt.role, []byte{HALT})
		if string(agent.Role) != tt.wantStr {
			t.Errorf("Expected role %q, got %q", tt.wantStr, agent.Role)
		}
	}
}

func TestAgentNewAgentDefaults(t *testing.T) {
	agent := NewAgent(Coordinator, []byte{HALT})

	if agent.ID == "" {
		t.Error("Expected non-empty agent ID")
	}
	if agent.Trust != 1.0 {
		t.Errorf("Expected default trust 1.0, got %f", agent.Trust)
	}
	if agent.VM == nil {
		t.Error("Expected non-nil VM")
	}
	if len(agent.MessageQueue) != 0 {
		t.Errorf("Expected empty message queue, got %d", len(agent.MessageQueue))
	}
	if agent.LastActive.IsZero() {
		t.Error("Expected non-zero LastActive time")
	}
}

func TestAgentReceiveMessage(t *testing.T) {
	agent := NewAgent(Worker, []byte{HALT})

	msg := A2AMessage{
		SenderID:   "sender-1",
		ReceiverID: agent.ID,
		Type:       BROADCAST,
		Payload:    "test payload",
	}

	agent.ReceiveMessage(msg)

	if len(agent.MessageQueue) != 1 {
		t.Errorf("Expected 1 message in queue, got %d", len(agent.MessageQueue))
	}

	queued := agent.MessageQueue[0]
	if queued.Payload != "test payload" {
		t.Errorf("Expected payload 'test payload', got %s", queued.Payload)
	}
	if queued.Type != BROADCAST {
		t.Errorf("Expected type BROADCAST, got %s", queued.Type)
	}
}

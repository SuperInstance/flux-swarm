package main

import (
	"fmt"
	"sync"
	"time"

	
)

// Role type
type Role string

const (
	Worker      Role = "worker"
	Scout       Role = "scout"
	Coordinator Role = "coordinator"
	Specialist  Role = "specialist"
)

// Agent represents a FLUX agent
type Agent struct {
	ID           string
	Role         Role
	Trust        float64
	VM           *FluxVM
	MessageQueue []A2AMessage
	mu           sync.Mutex
	LastActive   time.Time
}

// NewAgent creates a new agent with given role and bytecode
func NewAgent(role Role, bytecode []byte) *Agent {
	return &Agent{
		ID:           fmt.Sprintf("agent-%04d", time.Now().UnixNano()%10000),
		Role:         role,
		Trust:        1.0,
		VM:           NewFluxVM(bytecode),
		MessageQueue: make([]A2AMessage, 0),
		LastActive:   time.Now(),
	}
}

// Process executes one timestep of the agent's VM
func (a *Agent) Process() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if !a.VM.Running {
		return nil
	}
	
	err := a.VM.ExecuteStep()
	if err != nil {
		return err
	}
	
	a.LastActive = time.Now()
	return nil
}

// SendMessage prepares a message to be sent
func (a *Agent) SendMessage(receiverID string, msgType MessageType, payload string) A2AMessage {
	return A2AMessage{
		SenderID:   a.ID,
		ReceiverID: receiverID,
		Type:       msgType,
		Payload:    payload,
		Trust:      a.Trust,
		Timestamp:  time.Now(),
	}
}

// ReceiveMessage adds a message to the agent's queue
func (a *Agent) ReceiveMessage(msg A2AMessage) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.MessageQueue = append(a.MessageQueue, msg)
}

// GetResult returns the current register state
func (a *Agent) GetResult() [16]int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.VM.GetRegisters()
}

// GetNextMessage retrieves and removes the next message from the queue
func (a *Agent) GetNextMessage() *A2AMessage {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if len(a.MessageQueue) == 0 {
		return nil
	}
	
	msg := a.MessageQueue[0]
	a.MessageQueue = a.MessageQueue[1:]
	return &msg
}

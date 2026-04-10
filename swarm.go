package main

import (
	"sync"
	"time"
)

// SwarmCoordinator manages the fleet of agents
type SwarmCoordinator struct {
	Agents      map[string]*Agent
	mu          sync.RWMutex
	MessageLog  []A2AMessage
	TrustMatrix map[string]map[string]float64 // agentID -> agentID -> trust
}

// NewSwarmCoordinator creates a new swarm coordinator
func NewSwarmCoordinator() *SwarmCoordinator {
	return &SwarmCoordinator{
		Agents:      make(map[string]*Agent),
		MessageLog:  make([]A2AMessage, 0),
		TrustMatrix: make(map[string]map[string]float64),
	}
}

// RegisterAgent adds an agent to the swarm
func (sc *SwarmCoordinator) RegisterAgent(agent *Agent) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	sc.Agents[agent.ID] = agent
	sc.TrustMatrix[agent.ID] = make(map[string]float64)
	
	// Initialize trust with all existing agents
	for id := range sc.Agents {
		if id != agent.ID {
			sc.TrustMatrix[agent.ID][id] = 1.0
			sc.TrustMatrix[id][agent.ID] = 1.0
		}
	}
}

// RemoveAgent removes an agent from the swarm
func (sc *SwarmCoordinator) RemoveAgent(agentID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	delete(sc.Agents, agentID)
	delete(sc.TrustMatrix, agentID)
	
	// Remove from other agents' trust matrices
	for id := range sc.TrustMatrix {
		delete(sc.TrustMatrix[id], agentID)
	}
}

// BroadcastMessage sends a message to all agents
func (sc *SwarmCoordinator) BroadcastMessage(senderID string, msgType MessageType, payload string) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	for id := range sc.Agents {
		if id != senderID {
			msg := A2AMessage{
				SenderID:   senderID,
				ReceiverID: id,
				Type:       msgType,
				Payload:    payload,
				Trust:      sc.TrustMatrix[senderID][id],
				Timestamp:  time.Now(),
			}
			sc.Agents[id].ReceiveMessage(msg)
			sc.MessageLog = append(sc.MessageLog, msg)
		}
	}
}

// RouteMessage sends a point-to-point message
func (sc *SwarmCoordinator) RouteMessage(senderID, receiverID string, msgType MessageType, payload string) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	receiver, exists := sc.Agents[receiverID]
	if !exists {
		return false
	}
	
	msg := A2AMessage{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Type:       msgType,
		Payload:    payload,
		Trust:      sc.TrustMatrix[senderID][receiverID],
		Timestamp:  time.Now(),
	}
	
	receiver.ReceiveMessage(msg)
	sc.MessageLog = append(sc.MessageLog, msg)
	
	// Update trust based on message type
	sc.updateTrust(senderID, receiverID, msgType)
	
	return true
}

// updateTrust adjusts trust scores between agents
func (sc *SwarmCoordinator) updateTrust(senderID, receiverID string, msgType MessageType) {
	// Simple trust update logic
	switch msgType {
	case TELL:
		sc.TrustMatrix[senderID][receiverID] *= 1.01
		sc.TrustMatrix[receiverID][senderID] *= 1.01
	case ASK:
		sc.TrustMatrix[senderID][receiverID] *= 1.005
		sc.TrustMatrix[receiverID][senderID] *= 1.005
	case DELEGATE:
		sc.TrustMatrix[senderID][receiverID] *= 1.02
		sc.TrustMatrix[receiverID][senderID] *= 1.02
	}
	
	// Cap trust at 10.0
	if sc.TrustMatrix[senderID][receiverID] > 10.0 {
		sc.TrustMatrix[senderID][receiverID] = 10.0
	}
	if sc.TrustMatrix[receiverID][senderID] > 10.0 {
		sc.TrustMatrix[receiverID][senderID] = 10.0
	}
}

// RunTimestep executes all agents for one step
func (sc *SwarmCoordinator) RunTimestep() {
	sc.mu.RLock()
	agents := make([]*Agent, 0, len(sc.Agents))
	for _, agent := range sc.Agents {
		agents = append(agents, agent)
	}
	sc.mu.RUnlock()
	
	for _, agent := range agents {
		agent.Process()
	}
}

// GetFormation returns current agent topology
func (sc *SwarmCoordinator) GetFormation() map[string][]string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	formation := make(map[string][]string)
	for id := range sc.Agents {
		formation[id] = make([]string, 0)
		// For simplicity, connect each agent to up to 3 others
		count := 0
		for otherID := range sc.Agents {
			if otherID != id && count < 3 {
				formation[id] = append(formation[id], otherID)
				count++
			}
		}
	}
	return formation
}

// CollectResults aggregates all agent results
func (sc *SwarmCoordinator) CollectResults() map[string][16]int64 {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	results := make(map[string][16]int64)
	for id, agent := range sc.Agents {
		results[id] = agent.GetResult()
	}
	return results
}

// GetTotalMessages returns the number of messages sent
func (sc *SwarmCoordinator) GetTotalMessages() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return len(sc.MessageLog)
}

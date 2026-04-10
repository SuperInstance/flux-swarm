package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// SwarmCoordinator manages the fleet of agents
type SwarmCoordinator struct {
	Agents      map[string]*Agent
	mu          sync.RWMutex
	MessageLog  []A2AMessage
	TrustMatrix map[string]map[string]float64 // agentID -> agentID -> trust
	Consensus   *SwarmConsensus
	Lifecycle   *AgentLifecycle
	Vocabulary  *Vocabulary
}

// AgentLifecycle manages agent lifecycle (spawn, evolve, terminate)
type AgentLifecycle struct {
	BirthThreshold  float64 // Trust score needed to spawn new agents
	DeathThreshold  float64 // Trust score below which agents are terminated
	EvolutionRate   float64 // How fast agents evolve
	MutationRate    float64 // Chance of mutation during evolution
	MaxAgents       int     // Maximum number of agents in swarm
	TimestepCount   int     // Number of timesteps since creation
}

// SwarmConsensus manages consensus among agents
type SwarmConsensus struct {
	Proposals      map[string]*ConsensusProposal // proposalID -> proposal
	ActiveVotes    map[string]map[string]bool    // proposalID -> agentID -> voted
	RequiredQuorum float64                        // Percentage of agents needed for quorum
	Timeout        time.Duration                  // Time to wait for consensus
	mu             sync.RWMutex
}

// ConsensusProposal represents a proposal for consensus
type ConsensusProposal struct {
	ID          string
	ProposerID  string
	Type        ProposalType
	Content     string
	VotesFor    int
	VotesAgainst int
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Status      ProposalStatus
}

// ProposalType represents the type of consensus proposal
type ProposalType string

const (
	TaskAssignment  ProposalType = "task_assignment"
	ResourceShare   ProposalType = "resource_share"
	TopologyChange ProposalType = "topology_change"
	Evolution      ProposalType = "evolution"
)

// ProposalStatus represents the status of a proposal
type ProposalStatus string

const (
	ProposalOpen   ProposalStatus = "open"
	ProposalAccepted ProposalStatus = "accepted"
	ProposalRejected ProposalStatus = "rejected"
	ProposalExpired ProposalStatus = "expired"
)

// NewSwarmCoordinator creates a new swarm coordinator
func NewSwarmCoordinator() *SwarmCoordinator {
	return &SwarmCoordinator{
		Agents:      make(map[string]*Agent),
		MessageLog:  make([]A2AMessage, 0),
		TrustMatrix: make(map[string]map[string]float64),
		Vocabulary:  NewVocabulary(),
		Lifecycle: &AgentLifecycle{
			BirthThreshold: 2.0,
			DeathThreshold: 0.5,
			EvolutionRate:  0.1,
			MutationRate:   0.05,
			MaxAgents:      100,
			TimestepCount:  0,
		},
		Consensus: &SwarmConsensus{
			Proposals:      make(map[string]*ConsensusProposal),
			ActiveVotes:    make(map[string]map[string]bool),
			RequiredQuorum: 0.6,
			Timeout:        5 * time.Second,
		},
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

// ========== AGENT LIFECYCLE ==========

// SpawnAgent creates and registers a new agent
func (sc *SwarmCoordinator) SpawnAgent(role Role, bytecode []byte) *Agent {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if len(sc.Agents) >= sc.Lifecycle.MaxAgents {
		return nil
	}

	agent := NewAgent(role, bytecode)
	sc.Agents[agent.ID] = agent
	sc.TrustMatrix[agent.ID] = make(map[string]float64)

	// Initialize trust with all existing agents
	for id := range sc.Agents {
		if id != agent.ID {
			sc.TrustMatrix[agent.ID][id] = 1.0
			sc.TrustMatrix[id][agent.ID] = 1.0
		}
	}

	return agent
}

// EvolveAgent updates an agent based on its performance
func (sc *SwarmCoordinator) EvolveAgent(agentID string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	agent, exists := sc.Agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	// Check for evolution
	if agent.Trust > sc.Lifecycle.BirthThreshold {
		// Trust is high, spawn a child agent
		agent.Trust -= 0.5 // Cost of reproduction
		child := NewAgent(agent.Role, agent.VM.Bytecode)
		sc.Agents[child.ID] = child
		sc.TrustMatrix[child.ID] = make(map[string]float64)

		// Inherit trust relationships
		for id, trust := range sc.TrustMatrix[agentID] {
			sc.TrustMatrix[child.ID][id] = trust * sc.Lifecycle.EvolutionRate
			sc.TrustMatrix[id][child.ID] = trust * sc.Lifecycle.EvolutionRate
		}
	}

	// Apply mutations
	if sc.shouldMutate() {
		sc.mutateAgent(agent)
	}

	return nil
}

// TerminateAgent removes an agent from the swarm
func (sc *SwarmCoordinator) TerminateAgent(agentID string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	agent, exists := sc.Agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	// Check termination criteria
	if agent.Trust < sc.Lifecycle.DeathThreshold {
		return sc.removeAgent(agentID)
	}

	return nil
}

// removeAgent internal helper
func (sc *SwarmCoordinator) removeAgent(agentID string) error {
	delete(sc.Agents, agentID)
	delete(sc.TrustMatrix, agentID)

	for id := range sc.TrustMatrix {
		delete(sc.TrustMatrix[id], agentID)
	}

	return nil
}

// shouldMutate determines if an agent should mutate
func (sc *SwarmCoordinator) shouldMutate() bool {
	return randFloat64() < sc.Lifecycle.MutationRate
}

// mutateAgent applies random mutations to an agent
func (sc *SwarmCoordinator) mutateAgent(agent *Agent) {
	// Randomly adjust trust
	adjustment := randFloat64()*0.2 - 0.1
	agent.Trust += adjustment

	// Clamp trust to reasonable range
	if agent.Trust < 0 {
		agent.Trust = 0
	} else if agent.Trust > 10 {
		agent.Trust = 10
	}
}

// ProcessLifecycle runs lifecycle management for all agents
func (sc *SwarmCoordinator) ProcessLifecycle() {
	sc.mu.RLock()
	agentIDs := make([]string, 0, len(sc.Agents))
	for id := range sc.Agents {
		agentIDs = append(agentIDs, id)
	}
	sc.mu.RUnlock()

	for _, id := range agentIDs {
		sc.EvolveAgent(id)
		sc.TerminateAgent(id)
	}

	sc.Lifecycle.TimestepCount++
}

// GetLifecycleStats returns lifecycle statistics
func (sc *SwarmCoordinator) GetLifecycleStats() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	avgTrust := 0.0
	trustSum := 0.0
	for _, agent := range sc.Agents {
		trustSum += agent.Trust
	}
	if len(sc.Agents) > 0 {
		avgTrust = trustSum / float64(len(sc.Agents))
	}

	return map[string]interface{}{
		"total_agents":      len(sc.Agents),
		"avg_trust":         avgTrust,
		"timestep_count":    sc.Lifecycle.TimestepCount,
		"birth_threshold":   sc.Lifecycle.BirthThreshold,
		"death_threshold":   sc.Lifecycle.DeathThreshold,
		"max_agents":        sc.Lifecycle.MaxAgents,
	}
}

// ========== SWARM CONSENSUS ==========

// CreateProposal creates a new consensus proposal
func (sc *SwarmCoordinator) CreateProposal(proposerID string, ptype ProposalType, content string) string {
	sc.Consensus.mu.Lock()
	defer sc.Consensus.mu.Unlock()

	proposalID := fmt.Sprintf("prop-%d", time.Now().UnixNano())

	proposal := &ConsensusProposal{
		ID:          proposalID,
		ProposerID:  proposerID,
		Type:        ptype,
		Content:     content,
		VotesFor:    0,
		VotesAgainst: 0,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(sc.Consensus.Timeout),
		Status:      ProposalOpen,
	}

	sc.Consensus.Proposals[proposalID] = proposal
	sc.Consensus.ActiveVotes[proposalID] = make(map[string]bool)

	return proposalID
}

// VoteOnProposal casts a vote on a proposal
func (sc *SwarmCoordinator) VoteOnProposal(agentID, proposalID string, vote bool) error {
	sc.Consensus.mu.Lock()
	defer sc.Consensus.mu.Unlock()

	proposal, exists := sc.Consensus.Proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalOpen {
		return fmt.Errorf("proposal is not open for voting")
	}

	if time.Now().After(proposal.ExpiresAt) {
		proposal.Status = ProposalExpired
		return fmt.Errorf("proposal has expired")
	}

	// Check if already voted
	if _, voted := sc.Consensus.ActiveVotes[proposalID][agentID]; voted {
		return fmt.Errorf("agent has already voted on this proposal")
	}

	// Record vote
	sc.Consensus.ActiveVotes[proposalID][agentID] = true

	if vote {
		proposal.VotesFor++
	} else {
		proposal.VotesAgainst++
	}

	// Check for consensus
	sc.checkConsensus(proposal)

	return nil
}

// checkConsensus checks if consensus has been reached
func (sc *SwarmCoordinator) checkConsensus(proposal *ConsensusProposal) {
	sc.mu.RLock()
	totalAgents := len(sc.Agents)
	sc.mu.RUnlock()

	totalVotes := proposal.VotesFor + proposal.VotesAgainst
	requiredVotes := int(math.Ceil(float64(totalAgents) * sc.Consensus.RequiredQuorum))

	// Check if quorum reached
	if totalVotes >= requiredVotes {
		if proposal.VotesFor > proposal.VotesAgainst {
			proposal.Status = ProposalAccepted
			sc.executeProposal(proposal)
		} else {
			proposal.Status = ProposalRejected
		}
	}
}

// executeProposal executes an accepted proposal
func (sc *SwarmCoordinator) executeProposal(proposal *ConsensusProposal) {
	switch proposal.Type {
	case TaskAssignment:
		// Task assignment logic would be implemented here
	case ResourceShare:
		// Resource sharing logic would be implemented here
	case TopologyChange:
		// Topology change logic would be implemented here
	case Evolution:
		// Trigger evolution based on proposal
		sc.ProcessLifecycle()
	}
}

// GetProposal returns a proposal by ID
func (sc *SwarmCoordinator) GetProposal(proposalID string) (*ConsensusProposal, error) {
	sc.Consensus.mu.RLock()
	defer sc.Consensus.mu.RUnlock()

	proposal, exists := sc.Consensus.Proposals[proposalID]
	if !exists {
		return nil, fmt.Errorf("proposal not found: %s", proposalID)
	}

	return proposal, nil
}

// GetAllProposals returns all proposals
func (sc *SwarmCoordinator) GetAllProposals() []*ConsensusProposal {
	sc.Consensus.mu.RLock()
	defer sc.Consensus.mu.RUnlock()

	proposals := make([]*ConsensusProposal, 0, len(sc.Consensus.Proposals))
	for _, p := range sc.Consensus.Proposals {
		proposals = append(proposals, p)
	}

	return proposals
}

// CleanupExpiredProposals removes expired proposals
func (sc *SwarmCoordinator) CleanupExpiredProposals() {
	sc.Consensus.mu.Lock()
	defer sc.Consensus.mu.Unlock()

	now := time.Now()
	for _, proposal := range sc.Consensus.Proposals {
		if proposal.Status == ProposalOpen && now.After(proposal.ExpiresAt) {
			proposal.Status = ProposalExpired
		}
	}
}

// GetConsensusStats returns consensus statistics
func (sc *SwarmCoordinator) GetConsensusStats() map[string]interface{} {
	sc.Consensus.mu.RLock()
	defer sc.Consensus.mu.RUnlock()

	sc.mu.RLock()
	totalAgents := len(sc.Agents)
	sc.mu.RUnlock()

	totalProposals := len(sc.Consensus.Proposals)
	openProposals := 0
	acceptedProposals := 0
	rejectedProposals := 0

	for _, p := range sc.Consensus.Proposals {
		switch p.Status {
		case ProposalOpen:
			openProposals++
		case ProposalAccepted:
			acceptedProposals++
		case ProposalRejected:
			rejectedProposals++
		}
	}

	return map[string]interface{}{
		"total_proposals":    totalProposals,
		"open_proposals":     openProposals,
		"accepted_proposals": acceptedProposals,
		"rejected_proposals": rejectedProposals,
		"required_quorum":    sc.Consensus.RequiredQuorum,
		"total_agents":       totalAgents,
	}
}

// Utility function
func randFloat64() float64 {
	return float64(time.Now().UnixNano()%1000) / 1000.0
}

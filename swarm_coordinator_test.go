package main

import "testing"

// ========== SWARM COORDINATOR TESTS ==========

func TestRemoveAgent(t *testing.T) {
        swarm := NewSwarmCoordinator()
        agent1 := NewAgent(Worker, []byte{HALT})
        agent2 := NewAgent(Scout, []byte{HALT})

        swarm.RegisterAgent(agent1)
        swarm.RegisterAgent(agent2)

        if len(swarm.Agents) != 2 {
                t.Fatalf("Expected 2 agents, got %d", len(swarm.Agents))
        }

        swarm.RemoveAgent(agent1.ID)

        if len(swarm.Agents) != 1 {
                t.Errorf("Expected 1 agent after removal, got %d", len(swarm.Agents))
        }
        if _, exists := swarm.Agents[agent1.ID]; exists {
                t.Error("Agent1 still exists after removal")
        }

        // Trust matrix should be cleaned
        if _, exists := swarm.TrustMatrix[agent1.ID]; exists {
                t.Error("Agent1 still in trust matrix after removal")
        }
        if _, exists := swarm.TrustMatrix[agent2.ID][agent1.ID]; exists {
                t.Error("Agent1 still in agent2's trust row after removal")
        }
}

func TestRemoveNonexistentAgent(t *testing.T) {
        swarm := NewSwarmCoordinator()
        // Removing nonexistent agent should not panic
        swarm.RemoveAgent("nonexistent-id")
}

func TestBroadcastMessage(t *testing.T) {
        swarm := NewSwarmCoordinator()
        agent1 := NewAgent(Worker, []byte{HALT})
        agent2 := NewAgent(Scout, []byte{HALT})
        agent3 := NewAgent(Coordinator, []byte{HALT})

        swarm.RegisterAgent(agent1)
        swarm.RegisterAgent(agent2)
        swarm.RegisterAgent(agent3)

        swarm.BroadcastMessage(agent1.ID, TELL, "broadcast payload")

        // agent2 and agent3 should each receive the message
        if len(agent2.MessageQueue) != 1 {
                t.Errorf("Agent2 should have 1 message, got %d", len(agent2.MessageQueue))
        }
        if len(agent3.MessageQueue) != 1 {
                t.Errorf("Agent3 should have 1 message, got %d", len(agent3.MessageQueue))
        }

        // Sender should not receive their own broadcast
        if len(agent1.MessageQueue) != 0 {
                t.Errorf("Sender should have 0 messages, got %d", len(agent1.MessageQueue))
        }

        // Message log should have 2 entries
        if swarm.GetTotalMessages() != 2 {
                t.Errorf("Expected 2 messages in log, got %d", swarm.GetTotalMessages())
        }
}

func TestRouteMessageToNonexistent(t *testing.T) {
        swarm := NewSwarmCoordinator()
        agent := NewAgent(Worker, []byte{HALT})
        swarm.RegisterAgent(agent)

        success := swarm.RouteMessage(agent.ID, "nonexistent", TELL, "test")
        if success {
                t.Error("Routing to nonexistent agent should return false")
        }
}

func TestCollectResults(t *testing.T) {
        swarm := NewSwarmCoordinator()
        agent1 := NewAgent(Worker, []byte{MOVI, 0, 42, HALT})
        agent2 := NewAgent(Scout, []byte{MOVI, 0, 99, HALT})

        swarm.RegisterAgent(agent1)
        swarm.RegisterAgent(agent2)

        // Execute both agents
        for agent1.VM.Running {
                agent1.Process()
        }
        for agent2.VM.Running {
                agent2.Process()
        }

        results := swarm.CollectResults()
        if len(results) != 2 {
                t.Errorf("Expected results for 2 agents, got %d", len(results))
        }

        r1 := results[agent1.ID]
        if r1[0] != 42 {
                t.Errorf("Agent1 R0 expected 42, got %d", r1[0])
        }

        r2 := results[agent2.ID]
        if r2[0] != 99 {
                t.Errorf("Agent2 R0 expected 99, got %d", r2[0])
        }
}

func TestSpawnAgentMaxLimit(t *testing.T) {
        swarm := NewSwarmCoordinator()
        swarm.Lifecycle.MaxAgents = 3

        // Spawn up to limit
        for i := 0; i < 3; i++ {
                agent := swarm.SpawnAgent(Worker, []byte{HALT})
                if agent == nil {
                        t.Fatalf("Failed to spawn agent %d", i)
                }
        }

        // Next spawn should fail
        agent := swarm.SpawnAgent(Worker, []byte{HALT})
        if agent != nil {
                t.Error("Expected nil when exceeding MaxAgents limit")
        }
}

func TestEvolveAgentNotFound(t *testing.T) {
        swarm := NewSwarmCoordinator()
        err := swarm.EvolveAgent("nonexistent")
        if err == nil {
                t.Error("Expected error for nonexistent agent")
        }
}

func TestTerminateAgentNotFound(t *testing.T) {
        swarm := NewSwarmCoordinator()
        err := swarm.TerminateAgent("nonexistent")
        if err == nil {
                t.Error("Expected error for nonexistent agent")
        }
}

func TestTerminateAgentTrustTooHigh(t *testing.T) {
        swarm := NewSwarmCoordinator()
        agent := swarm.SpawnAgent(Worker, []byte{HALT})
        agent.Trust = 5.0 // Above DeathThreshold of 0.5

        err := swarm.TerminateAgent(agent.ID)
        if err != nil {
                t.Errorf("Should not terminate agent with high trust: %v", err)
        }
        if _, exists := swarm.Agents[agent.ID]; !exists {
                t.Error("Agent should still exist when trust is high")
        }
}

func TestGetProposalNotFound(t *testing.T) {
        swarm := NewSwarmCoordinator()
        _, err := swarm.GetProposal("nonexistent")
        if err == nil {
                t.Error("Expected error for nonexistent proposal")
        }
}

func TestGetAllProposals(t *testing.T) {
        swarm := NewSwarmCoordinator()
        agent := swarm.SpawnAgent(Worker, []byte{HALT})

        swarm.CreateProposal(agent.ID, TaskAssignment, "Task A")
        swarm.CreateProposal(agent.ID, ResourceShare, "Share B")

        proposals := swarm.GetAllProposals()
        if len(proposals) != 2 {
                t.Errorf("Expected 2 proposals, got %d", len(proposals))
        }
}

func TestVoteOnProposalErrors(t *testing.T) {
        swarm := NewSwarmCoordinator()
        agent := swarm.SpawnAgent(Worker, []byte{HALT})
        proposalID := swarm.CreateProposal(agent.ID, Evolution, "Evolve")

        // Vote on nonexistent proposal
        err := swarm.VoteOnProposal(agent.ID, "nonexistent", true)
        if err == nil {
                t.Error("Expected error for voting on nonexistent proposal")
        }

        // Double vote
        err = swarm.VoteOnProposal(agent.ID, proposalID, true)
        if err != nil {
                t.Fatalf("First vote should succeed: %v", err)
        }
        err = swarm.VoteOnProposal(agent.ID, proposalID, false)
        if err == nil {
                t.Error("Expected error for double voting")
        }
}

func TestProcessLifecycle(t *testing.T) {
        swarm := NewSwarmCoordinator()

        // Add agents with low trust so they get terminated
        for i := 0; i < 3; i++ {
                agent := swarm.SpawnAgent(Worker, []byte{HALT})
                agent.Trust = 0.1 // Below DeathThreshold of 0.5
        }

        initialTimestep := swarm.Lifecycle.TimestepCount
        swarm.ProcessLifecycle()

        // Timestep should have incremented
        if swarm.Lifecycle.TimestepCount != initialTimestep+1 {
                t.Errorf("Expected timestep %d, got %d", initialTimestep+1, swarm.Lifecycle.TimestepCount)
        }
}

func TestCleanupExpiredProposals(t *testing.T) {
        swarm := NewSwarmCoordinator()
        swarm.Consensus.Timeout = 0 // Expire immediately BEFORE creating proposal
        agent := swarm.SpawnAgent(Worker, []byte{HALT})

        // Create a proposal (will expire immediately since Timeout is 0)
        proposalID := swarm.CreateProposal(agent.ID, TopologyChange, "Change")

        swarm.CleanupExpiredProposals()

        proposal, err := swarm.GetProposal(proposalID)
        if err != nil {
                t.Fatalf("Failed to get proposal: %v", err)
        }
        if proposal.Status != ProposalExpired {
                t.Errorf("Expected expired status, got %s", proposal.Status)
        }
}

func TestConsensusRejected(t *testing.T) {
        swarm := NewSwarmCoordinator()

        agents := make([]*Agent, 3)
        for i := 0; i < 3; i++ {
                agents[i] = swarm.SpawnAgent(Worker, []byte{HALT})
        }

        proposalID := swarm.CreateProposal(agents[0].ID, TaskAssignment, "Task")

        // All vote against
        for _, agent := range agents {
                swarm.VoteOnProposal(agent.ID, proposalID, false)
        }

        proposal, _ := swarm.GetProposal(proposalID)
        if proposal.Status != ProposalRejected {
                t.Errorf("Expected rejected status, got %s", proposal.Status)
        }
}

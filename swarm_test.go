package main

import (
	"testing"
)

func TestVMExecution(t *testing.T) {
	// Test simple addition
	bytecode := []byte{
		MOVI, 0, 5,
		MOVI, 1, 3,
		IADD, 0, 1,
		HALT,
	}
	
	vm := NewFluxVM(bytecode)
	
	// Execute all steps
	for vm.Running {
		err := vm.ExecuteStep()
		if err != nil {
			t.Fatalf("Execution error: %v", err)
		}
	}
	
	regs := vm.GetRegisters()
	if regs[0] != 8 {
		t.Errorf("Expected R0=8, got %d", regs[0])
	}
}

func TestAgentMessaging(t *testing.T) {
	agent1 := NewAgent(Worker, []byte{HALT})
	agent2 := NewAgent(Scout, []byte{HALT})
	
	msg := agent1.SendMessage(agent2.ID, TELL, "Hello")
	agent2.ReceiveMessage(msg)
	
	nextMsg := agent2.GetNextMessage()
	if nextMsg == nil {
		t.Fatal("Expected a message in queue")
	}
	
	if nextMsg.Payload != "Hello" {
		t.Errorf("Expected payload 'Hello', got %s", nextMsg.Payload)
	}
}

func TestSwarmCoordination(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	agent1 := NewAgent(Worker, []byte{HALT})
	agent2 := NewAgent(Scout, []byte{HALT})
	
	swarm.RegisterAgent(agent1)
	swarm.RegisterAgent(agent2)
	
	if len(swarm.Agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(swarm.Agents))
	}
	
	// Test message routing
	success := swarm.RouteMessage(agent1.ID, agent2.ID, TELL, "Test")
	if !success {
		t.Error("Failed to route message")
	}
	
	if len(swarm.MessageLog) != 1 {
		t.Errorf("Expected 1 message in log, got %d", len(swarm.MessageLog))
	}
}

func TestTrustUpdate(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	agent1 := NewAgent(Worker, []byte{HALT})
	agent2 := NewAgent(Scout, []byte{HALT})
	
	swarm.RegisterAgent(agent1)
	swarm.RegisterAgent(agent2)
	
	initialTrust := swarm.TrustMatrix[agent1.ID][agent2.ID]
	
	// Send a message to update trust
	swarm.RouteMessage(agent1.ID, agent2.ID, DELEGATE, "Task")
	
	newTrust := swarm.TrustMatrix[agent1.ID][agent2.ID]
	
	if newTrust <= initialTrust {
		t.Errorf("Expected trust to increase, got initial=%.2f, new=%.2f", 
			initialTrust, newTrust)
	}
}

func TestFormation(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	// Add 5 agents
	for i := 0; i < 5; i++ {
		agent := NewAgent(Worker, []byte{HALT})
		swarm.RegisterAgent(agent)
	}
	
	formation := swarm.GetFormation()
	
	if len(formation) != 5 {
		t.Errorf("Expected formation for 5 agents, got %d", len(formation))
	}
	
	for agentID, connections := range formation {
		if len(connections) > 3 {
			t.Errorf("Agent %s has too many connections: %d", 
				agentID[:4], len(connections))
		}
	}
}

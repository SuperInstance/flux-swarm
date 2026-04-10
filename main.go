package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starting FLUX Swarm Coordinator Demo")
	
	// Create swarm coordinator
	swarm := NewSwarmCoordinator()
	
	// Create bytecode for different agent types
	factorialBytecode := []byte{
		MOVI, 0, 5,   // R0 = 5 (n)
		MOVI, 1, 1,   // R1 = 1 (result)
		MOVI, 2, 1,   // R2 = 1 (counter)
		// loop:
		CMP, 0, 2,    // Compare R0 and R2
		JZ, 18,       // If equal, jump to end
		IMUL, 1, 2,   // R1 *= R2
		INC, 2,       // R2++
		JMP, 6,       // Jump back to loop
		HALT,
	}
	
	fibonacciBytecode := []byte{
		MOVI, 0, 10,  // R0 = n
		MOVI, 1, 0,   // R1 = a
		MOVI, 2, 1,   // R2 = b
		MOVI, 3, 0,   // R3 = i
		// loop:
		CMP, 3, 0,
		JZ, 24,
		MOVI, 4, 0,
		IADD, 4, 2,
		MOVI, 5, 0,
		IADD, 5, 1,
		IADD, 5, 2,
		MOVI, 1, 0,
		IADD, 1, 2,
		MOVI, 2, 0,
		IADD, 2, 5,
		INC, 3,
		JMP, 8,
		HALT,
	}
	
	// Create agents with different roles and bytecode
	roles := []Role{Worker, Scout, Coordinator, Specialist}
	bytecodes := [][]byte{factorialBytecode, fibonacciBytecode}
	
	for i := 0; i < 20; i++ {
		role := roles[i%len(roles)]
		bytecode := bytecodes[i%len(bytecodes)]
		agent := NewAgent(role, bytecode)
		swarm.RegisterAgent(agent)
	}
	
	// Run simulation for 50 timesteps
	fmt.Printf("Running simulation for 50 timesteps...\n")
	for step := 0; step < 50; step++ {
		// Run all agents
		swarm.RunTimestep()
		
		// Simulate some communication
		if step%5 == 0 {
			// Broadcast a message from first agent
			firstAgentID := ""
			for id := range swarm.Agents {
				firstAgentID = id
				break
			}
			swarm.BroadcastMessage(firstAgentID, TELL, fmt.Sprintf("Step %d", step))
		}
		
		if step%7 == 0 {
			// Route some point-to-point messages
			agentIDs := make([]string, 0, len(swarm.Agents))
			for id := range swarm.Agents {
				agentIDs = append(agentIDs, id)
			}
			if len(agentIDs) >= 2 {
				swarm.RouteMessage(agentIDs[0], agentIDs[1], ASK, "Need help")
			}
		}
		
		// Print visualizations every 10 steps
		if step%10 == 0 {
			fmt.Printf("\n=== TIMESTEP %d ===\n", step)
			PrintSwarmStatus(swarm)
			PrintMessageFlow(swarm, 5)
			PrintFormation(swarm)
		}
		
		time.Sleep(10 * time.Millisecond)
	}
	
	// Final report
	fmt.Println("\n=== FINAL REPORT ===")
	fmt.Printf("Total agents: %d\n", len(swarm.Agents))
	fmt.Printf("Total messages sent: %d\n", swarm.GetTotalMessages())
	
	// Print trust matrix
	PrintTrustMatrix(swarm)
	
	// Collect and show some results
	results := swarm.CollectResults()
	fmt.Println("\nSample agent results (first 3 agents):")
	count := 0
	for agentID, regs := range results {
		if count >= 3 {
			break
		}
		fmt.Printf("Agent %s: R0=%d, R1=%d, R2=%d\n", 
			agentID[:8], regs[0], regs[1], regs[2])
		count++
	}
	
	fmt.Println("\nDemo completed successfully!")
}

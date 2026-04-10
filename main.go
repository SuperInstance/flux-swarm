package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starting FLUX Swarm Coordinator Demo")
	fmt.Println("========================================")

	// Create swarm coordinator
	swarm := NewSwarmCoordinator()

	// Demo 1: Vocabulary System
	fmt.Println("--- VOCABULARY SYSTEM ---")
	stats := swarm.Vocabulary.GetVocabularyStats()
	fmt.Printf("Vocabulary: %d opcodes, %d aliases, %d macros, %d categories\n",
		stats["opcodes"], stats["aliases"], stats["macros"], stats["categories"])
	
	// Show some vocabulary info
	fmt.Println("\nArithmetic instructions:")
	for _, instr := range swarm.Vocabulary.GetCategory("arithmetic") {
		doc := swarm.Vocabulary.GetDocstring(instr)
		fmt.Printf("  %s: %s\n", instr, doc)
	}
	
	// Demo 2: Assembler
	fmt.Println("\n--- ASSEMBLER DEMO ---")
	assemblySource := `; Calculate 5! (factorial of 5)
loop_start:
	MOVI R0, 5    ; n = 5
	MOVI R1, 1    ; result = 1
	MOVI R2, 1    ; counter = 1
compare:
	CMP R0, R2    ; compare n and counter
	JZ end        ; if equal, we're done
	IMUL R1, R2   ; result *= counter
	INC R2        ; counter++
	JMP compare   ; loop back
end:
	HALT
`

	factorialBytecode, err := AssembleString(assemblySource)
	if err != nil {
		fmt.Printf("Assembly error: %v\n", err)
	} else {
		fmt.Printf("Assembled %d bytes\n", len(factorialBytecode))
	}

	// Demo 3: Disassembler
	fmt.Println("\n--- DISASSEMBLER DEMO ---")
	diasmOutput := DisassembleString(factorialBytecode)
	fmt.Println(diasmOutput)

	// Demo 4: Format Bytecode
	fmt.Println("--- BYTECODE FORMAT ---")
	fmt.Println(FormatBytecode(factorialBytecode))

	// Demo 5: Verify Bytecode
	fmt.Println("--- BYTECODE VERIFICATION ---")
	errors := VerifyBytecode(factorialBytecode)
	if len(errors) == 0 {
		fmt.Println("Bytecode is valid FLUX ISA!")
	} else {
		fmt.Println("Errors:")
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	// Demo 6: Create agents with different bytecode
	fmt.Println("\n--- CREATING AGENTS ---")
	fibonacciBytecode, _ := AssembleString(`; Fibonacci sequence
	MOVI R0, 10   ; n = 10
	MOVI R1, 0    ; a = 0
	MOVI R2, 1    ; b = 1
	MOVI R3, 10   ; counter = n
loop_fib:
	CMP R3, 0
	JZ end_fib
	MOVI R4, 0    ; temp = a
	IADD R4, R2   ; temp += b
	MOVI R1, 0    ; a = b
	IADD R1, R2
	MOVI R2, 0    ; b = temp
	IADD R2, R4
	DEC R3        ; counter--
	JMP loop_fib
end_fib:
	HALT
`)

	// Create initial agents
	roles := []Role{Worker, Scout, Coordinator, Specialist}
	bytecodes := [][]byte{factorialBytecode, fibonacciBytecode}

	for i := 0; i < 10; i++ {
		role := roles[i%len(roles)]
		bytecode := bytecodes[i%len(bytecodes)]
		swarm.SpawnAgent(role, bytecode)
	}
	fmt.Printf("Created %d initial agents\n", len(swarm.Agents))

	// Demo 7: Swarm Consensus
	fmt.Println("\n--- SWARM CONSENSUS ---")
	agentIDs := make([]string, 0, len(swarm.Agents))
	for id := range swarm.Agents {
		agentIDs = append(agentIDs, id)
	}
	if len(agentIDs) > 0 {
		proposalID := swarm.CreateProposal(agentIDs[0], Evolution, "Trigger evolution phase")
		fmt.Printf("Created proposal: %s\n", proposalID)
		
		// Cast votes
		for i, id := range agentIDs {
			if i < len(agentIDs)/2+1 { // Ensure majority
				swarm.VoteOnProposal(id, proposalID, true)
			}
		}
		
		proposal, _ := swarm.GetProposal(proposalID)
		fmt.Printf("Proposal status: %s (For: %d, Against: %d)\n", 
			proposal.Status, proposal.VotesFor, proposal.VotesAgainst)
	}

	consensusStats := swarm.GetConsensusStats()
	fmt.Printf("Consensus stats: %d proposals, %d open, %d accepted, %d rejected\n",
		consensusStats["total_proposals"], consensusStats["open_proposals"],
		consensusStats["accepted_proposals"], consensusStats["rejected_proposals"])

	// Demo 8: Run simulation
	fmt.Println("\n--- RUNNING SIMULATION ---")
	fmt.Printf("Running for 30 timesteps...\n")
	
	for step := 0; step < 30; step++ {
		// Run all agents
		swarm.RunTimestep()
		
		// Process lifecycle every 5 steps
		if step%5 == 0 {
			swarm.ProcessLifecycle()
			swarm.CleanupExpiredProposals()
		}
		
		// Simulate communication
		if step%3 == 0 {
			firstAgentID := agentIDs[0]
			if len(agentIDs) > 1 {
				swarm.BroadcastMessage(firstAgentID, TELL, fmt.Sprintf("Timestep %d", step))
			}
		}
		
		if step%4 == 0 && len(agentIDs) >= 2 {
			swarm.RouteMessage(agentIDs[0], agentIDs[1], ASK, "Status report")
		}
		
		// Print status every 10 steps
		if step%10 == 0 {
			fmt.Printf("\n=== TIMESTEP %d ===\n", step)
			PrintSwarmStatus(swarm)
			
			// Show lifecycle stats
			lifecycleStats := swarm.GetLifecycleStats()
			fmt.Printf("Lifecycle: Agents=%d, AvgTrust=%.2f, Timesteps=%d\n",
				lifecycleStats["total_agents"],
				lifecycleStats["avg_trust"],
				lifecycleStats["timestep_count"])
		}
		
		time.Sleep(5 * time.Millisecond)
	}

	// Demo 9: Final Report
	fmt.Println("\n=== FINAL REPORT ===")
	
	// Agent statistics
	fmt.Printf("Total agents: %d\n", len(swarm.Agents))
	fmt.Printf("Total messages: %d\n", swarm.GetTotalMessages())
	
	// Lifecycle final stats
	lifecycleStats := swarm.GetLifecycleStats()
	fmt.Printf("Final lifecycle stats:\n")
	fmt.Printf("  Total agents: %d\n", lifecycleStats["total_agents"])
	fmt.Printf("  Average trust: %.3f\n", lifecycleStats["avg_trust"])
	fmt.Printf("  Total timesteps: %d\n", lifecycleStats["timestep_count"])
	
	// Consensus final stats
	consensusStats = swarm.GetConsensusStats()
	fmt.Printf("\nFinal consensus stats:\n")
	fmt.Printf("  Total proposals: %d\n", consensusStats["total_proposals"])
	fmt.Printf("  Accepted: %d, Rejected: %d, Expired: %d\n",
		consensusStats["accepted_proposals"],
		consensusStats["rejected_proposals"],
		consensusStats["open_proposals"])
	
	// Print formation
	fmt.Println("\n--- AGENT FORMATION ---")
	PrintFormation(swarm)
	
	// Print sample results
	fmt.Println("\n--- SAMPLE AGENT RESULTS (first 3) ---")
	results := swarm.CollectResults()
	count := 0
	for agentID, regs := range results {
		if count >= 3 {
			break
		}
		
		// Get agent info
		agent := swarm.Agents[agentID]
		fmt.Printf("Agent %s [%s]: R0=%d, R1=%d, R2=%d, Trust=%.2f\n",
			agentID[:8], agent.Role, regs[0], regs[1], regs[2], agent.Trust)
		count++
	}
	
	// Print trust matrix (subset)
	fmt.Println("\n--- TRUST MATRIX (first 5 agents) ---")
	agentSubset := make([]string, 0)
	for id := range swarm.Agents {
		if len(agentSubset) >= 5 {
			break
		}
		agentSubset = append(agentSubset, id)
	}
	
	fmt.Print("     ")
	for _, id := range agentSubset {
		fmt.Printf("%4s", id[:2])
	}
	fmt.Println()
	
	for _, id1 := range agentSubset {
		fmt.Printf("%2s: ", id1[:2])
		for _, id2 := range agentSubset {
			if id1 == id2 {
				fmt.Print("  - ")
			} else {
				trust := swarm.TrustMatrix[id1][id2]
				fmt.Printf("%4.1f", trust)
			}
		}
		fmt.Println()
	}

	fmt.Println("\nDemo completed successfully!")
	fmt.Println("========================================")
}

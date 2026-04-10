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

// ========== ASSEMBLER TESTS ==========

func TestAssemblerMOV(t *testing.T) {
	source := `MOVI R0, 5
MOV R1, R0
HALT`

	bytecode, err := AssembleString(source)
	if err != nil {
		t.Fatalf("Assembly failed: %v", err)
	}

	// Should be: MOVI(0x2B), 0, 5, MOV(0x01), 1, 0, HALT(0x80)
	expected := []byte{0x2B, 0, 5, 0x01, 1, 0, 0x80}

	if len(bytecode) != len(expected) {
		t.Fatalf("Expected %d bytes, got %d", len(expected), len(bytecode))
	}

	for i := range expected {
		if bytecode[i] != expected[i] {
			t.Errorf("Byte %d: expected 0x%02X, got 0x%02X", i, expected[i], bytecode[i])
		}
	}
}

func TestAssemblerMOVExecution(t *testing.T) {
	// Test that MOV instruction works correctly
	bytecode := []byte{
		MOVI, 0, 42,  // R0 = 42
		MOV, 1, 0,    // R1 = R0
		HALT,
	}

	vm := NewFluxVM(bytecode)

	for vm.Running {
		err := vm.ExecuteStep()
		if err != nil {
			t.Fatalf("Execution error: %v", err)
		}
	}

	regs := vm.GetRegisters()
	if regs[0] != 42 {
		t.Errorf("Expected R0=42, got %d", regs[0])
	}
	if regs[1] != 42 {
		t.Errorf("Expected R1=42 (copied from R0), got %d", regs[1])
	}
}

func TestVocabularyPatterns(t *testing.T) {
	vocab := NewVocabulary()

	// Test pattern recognition
	pattern, ok := vocab.RecognizePattern("compute 5 + 3")
	if !ok {
		t.Error("Failed to recognize compute pattern")
	}
	if pattern != "compute_add" {
		t.Errorf("Expected compute_add pattern, got %s", pattern)
	}

	pattern, ok = vocab.RecognizePattern("factorial of 5")
	if !ok {
		t.Error("Failed to recognize factorial pattern")
	}
	if pattern != "factorial" {
		t.Errorf("Expected factorial pattern, got %s", pattern)
	}

	pattern, ok = vocab.RecognizePattern("hello world")
	if !ok {
		t.Error("Failed to recognize hello pattern")
	}
	if pattern != "hello" {
		t.Errorf("Expected hello pattern, got %s", pattern)
	}

	// Test pattern expansion
	args := map[string]string{"X": "5", "Y": "3"}
	expanded, err := vocab.ExpandPattern("compute_add", args)
	if err != nil {
		t.Errorf("Failed to expand compute_add pattern: %v", err)
	}
	if !containsSubstring(expanded, "MOVI R0, 5") {
		t.Error("Expected MOVI R0, 5 in expanded pattern")
	}
	if !containsSubstring(expanded, "MOVI R1, 3") {
		t.Error("Expected MOVI R1, 3 in expanded pattern")
	}
}



func TestAssemblerSimple(t *testing.T) {
	source := `MOVI R0, 5
MOVI R1, 3
IADD R0, R1
HALT`

	bytecode, err := AssembleString(source)
	if err != nil {
		t.Fatalf("Assembly failed: %v", err)
	}

	// Should be: MOVI(0x2B), 0, 5, MOVI(0x2B), 1, 3, IADD(0x08), 0, 1, HALT(0x80)
	expected := []byte{0x2B, 0, 5, 0x2B, 1, 3, 0x08, 0, 1, 0x80}

	if len(bytecode) != len(expected) {
		t.Fatalf("Expected %d bytes, got %d", len(expected), len(bytecode))
	}

	for i := range expected {
		if bytecode[i] != expected[i] {
			t.Errorf("Byte %d: expected 0x%02X, got 0x%02X", i, expected[i], bytecode[i])
		}
	}
}

func TestAssemblerWithLabels(t *testing.T) {
	source := `loop:
	INC R0
	CMP R0, 10
	JNZ loop
	HALT`

	bytecode, err := AssembleString(source)
	if err != nil {
		t.Fatalf("Assembly failed: %v", err)
	}

	if len(bytecode) < 7 {
		t.Fatalf("Bytecode too short: %d bytes", len(bytecode))
	}

	// Verify first instruction is INC (0x0E)
	if bytecode[0] != 0x0E {
		t.Errorf("Expected INC (0x0E), got 0x%02X", bytecode[0])
	}
}

func TestAssemblerInvalidOpcode(t *testing.T) {
	source := "INVALID R0, 5"
	
	_, err := AssembleString(source)
	if err == nil {
		t.Error("Expected error for invalid opcode")
	}
}

// ========== DISASSEMBLER TESTS ==========

func TestDisassemblerSimple(t *testing.T) {
	bytecode := []byte{0x2B, 0, 5, 0x2B, 1, 3, 0x08, 0, 1, 0x80}

	assembly := DisassembleString(bytecode)

	if len(assembly) == 0 {
		t.Fatal("Disassembly produced no output")
	}

	// Check for expected mnemonics
	if !containsSubstring(assembly, "MOVI") {
		t.Error("Expected MOVI in disassembly")
	}
	if !containsSubstring(assembly, "IADD") {
		t.Error("Expected IADD in disassembly")
	}
	if !containsSubstring(assembly, "HALT") {
		t.Error("Expected HALT in disassembly")
	}
}

func TestDisassemblerUnknownOpcode(t *testing.T) {
	bytecode := []byte{0x99, 0, 5} // Unknown opcode
	
	assembly := DisassembleString(bytecode)
	
	if !containsSubstring(assembly, "UNKNOWN") {
		t.Error("Expected UNKNOWN marker for invalid opcode")
	}
}

func TestVerifyBytecode(t *testing.T) {
	// Valid bytecode
	validBytecode := []byte{0x2B, 0, 5, 0x08, 0, 1, 0x80}
	errors := VerifyBytecode(validBytecode)

	if len(errors) != 0 {
		t.Errorf("Valid bytecode produced errors: %v", errors)
	}

	// Invalid bytecode (out of range register)
	invalidBytecode := []byte{0x2B, 20, 5}
	errors = VerifyBytecode(invalidBytecode)

	if len(errors) == 0 {
		t.Error("Expected error for out-of-range register")
	}
}

// ========== VOCABULARY TESTS ==========

func TestVocabularyLookup(t *testing.T) {
	vocab := NewVocabulary()
	
	// Test direct opcode lookup
	opcode, ok := vocab.LookupOpcode("MOVI")
	if !ok {
		t.Error("MOVI not found in vocabulary")
	}
	if opcode != MOVI {
		t.Errorf("Expected MOVI opcode 0x%02X, got 0x%02X", MOVI, opcode)
	}
	
	// Test alias lookup
	opcode, ok = vocab.LookupOpcode("ADD")
	if !ok {
		t.Error("ADD alias not found")
	}
	if opcode != IADD {
		t.Errorf("Expected IADD opcode for ADD alias, got 0x%02X", opcode)
	}
}

func TestVocabularyCategories(t *testing.T) {
	vocab := NewVocabulary()
	
	categories := vocab.GetCategories()
	if len(categories) == 0 {
		t.Fatal("No categories found")
	}
	
	arithmetic := vocab.GetCategory("arithmetic")
	if len(arithmetic) == 0 {
		t.Error("Arithmetic category is empty")
	}
	
	// Check that IADD is in arithmetic
	found := false
	for _, instr := range arithmetic {
		if instr == "IADD" {
			found = true
			break
		}
	}
	if !found {
		t.Error("IADD not found in arithmetic category")
	}
}

func TestVocabularyMacros(t *testing.T) {
	vocab := NewVocabulary()
	
	// Check CLEAR macro exists
	macro, ok := vocab.GetMacro("CLEAR")
	if !ok {
		t.Error("CLEAR macro not found")
	}
	
	if len(macro.Params) != 1 {
		t.Errorf("Expected 1 parameter for CLEAR macro, got %d", len(macro.Params))
	}
	
	// Expand macro
	expanded, err := vocab.ExpandMacro("CLEAR", []string{"R0"})
	if err != nil {
		t.Fatalf("Failed to expand CLEAR macro: %v", err)
	}
	
	if !containsSubstring(expanded, "MOVI") {
		t.Error("Expected MOVI in expanded CLEAR macro")
	}
	if !containsSubstring(expanded, "R0") {
		t.Error("Expected R0 in expanded CLEAR macro")
	}
}

func TestVocabularyValidation(t *testing.T) {
	vocab := NewVocabulary()
	
	errors := vocab.Validate()
	
	if len(errors) > 0 {
		t.Errorf("Vocabulary validation failed: %v", errors)
	}
}

func TestVocabularyStats(t *testing.T) {
	vocab := NewVocabulary()
	
	stats := vocab.GetVocabularyStats()
	
	if stats["opcodes"] == 0 {
		t.Error("No opcodes in vocabulary")
	}
	if stats["macros"] == 0 {
		t.Error("No macros in vocabulary")
	}
	if stats["categories"] == 0 {
		t.Error("No categories in vocabulary")
	}
}

// ========== LIFECYCLE TESTS ==========

func TestSpawnAgent(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	initialCount := len(swarm.Agents)
	
	agent := swarm.SpawnAgent(Worker, []byte{HALT})
	if agent == nil {
		t.Fatal("Failed to spawn agent")
	}
	
	if len(swarm.Agents) != initialCount+1 {
		t.Errorf("Expected %d agents, got %d", initialCount+1, len(swarm.Agents))
	}
}

func TestTerminateAgent(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	agent := swarm.SpawnAgent(Worker, []byte{HALT})
	if agent == nil {
		t.Fatal("Failed to spawn agent")
	}
	
	// Set trust low enough to trigger termination
	agent.Trust = 0.4
	swarm.Lifecycle.DeathThreshold = 0.5
	
	err := swarm.TerminateAgent(agent.ID)
	if err != nil {
		t.Errorf("Failed to terminate agent: %v", err)
	}
	
	if _, exists := swarm.Agents[agent.ID]; exists {
		t.Error("Agent still exists after termination")
	}
}

func TestEvolveAgent(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	agent := swarm.SpawnAgent(Worker, []byte{HALT})
	if agent == nil {
		t.Fatal("Failed to spawn agent")
	}
	
	initialCount := len(swarm.Agents)
	
	// Set trust high enough to trigger evolution
	agent.Trust = 3.0
	swarm.Lifecycle.BirthThreshold = 2.0
	
	err := swarm.EvolveAgent(agent.ID)
	if err != nil {
		t.Errorf("Failed to evolve agent: %v", err)
	}
	
	// Should have spawned a child
	if len(swarm.Agents) <= initialCount {
		t.Errorf("Expected agent count to increase, got %d", len(swarm.Agents))
	}
}

func TestLifecycleStats(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	// Add some agents
	for i := 0; i < 5; i++ {
		swarm.SpawnAgent(Worker, []byte{HALT})
	}
	
	stats := swarm.GetLifecycleStats()
	
	if stats["total_agents"] != 5 {
		t.Errorf("Expected 5 agents, got %v", stats["total_agents"])
	}
	if stats["timestep_count"] == nil {
		t.Error("Missing timestep_count in stats")
	}
}

// ========== CONSENSUS TESTS ==========

func TestCreateProposal(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	agent := swarm.SpawnAgent(Worker, []byte{HALT})
	if agent == nil {
		t.Fatal("Failed to spawn agent")
	}
	
	proposalID := swarm.CreateProposal(agent.ID, TaskAssignment, "Do task")
	if proposalID == "" {
		t.Fatal("Failed to create proposal")
	}
	
	proposal, err := swarm.GetProposal(proposalID)
	if err != nil {
		t.Fatalf("Failed to get proposal: %v", err)
	}
	
	if proposal.ProposerID != agent.ID {
		t.Errorf("Expected proposer %s, got %s", agent.ID, proposal.ProposerID)
	}
	
	if proposal.Status != ProposalOpen {
		t.Errorf("Expected status %s, got %s", ProposalOpen, proposal.Status)
	}
}

func TestVoteOnProposal(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	// Create agents
	agents := make([]*Agent, 3)
	for i := 0; i < 3; i++ {
		agents[i] = swarm.SpawnAgent(Worker, []byte{HALT})
		if agents[i] == nil {
			t.Fatal("Failed to spawn agent")
		}
	}
	
	// Create proposal
	proposalID := swarm.CreateProposal(agents[0].ID, Evolution, "Evolve")
	
	// Cast votes
	swarm.VoteOnProposal(agents[0].ID, proposalID, true)
	swarm.VoteOnProposal(agents[1].ID, proposalID, true)
	
	proposal, _ := swarm.GetProposal(proposalID)
	
	if proposal.VotesFor != 2 {
		t.Errorf("Expected 2 votes for, got %d", proposal.VotesFor)
	}
}

func TestConsensusQuorum(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	// Create agents (quorum is 60%, so need 2 of 3)
	agents := make([]*Agent, 3)
	for i := 0; i < 3; i++ {
		agents[i] = swarm.SpawnAgent(Worker, []byte{HALT})
	}
	
	// Create proposal
	proposalID := swarm.CreateProposal(agents[0].ID, TaskAssignment, "Task")
	
	// Cast all votes - majority wins
	swarm.VoteOnProposal(agents[0].ID, proposalID, true)
	swarm.VoteOnProposal(agents[1].ID, proposalID, true)
	swarm.VoteOnProposal(agents[2].ID, proposalID, false)
	
	proposal, _ := swarm.GetProposal(proposalID)
	
	if proposal.Status != ProposalAccepted {
		t.Errorf("Expected proposal to be accepted, got status %s", proposal.Status)
	}
}

func TestConsensusStats(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	agent := swarm.SpawnAgent(Worker, []byte{HALT})
	
	// Create a proposal
	swarm.CreateProposal(agent.ID, Evolution, "Evolve")
	
	stats := swarm.GetConsensusStats()
	
	if stats["total_proposals"] != 1 {
		t.Errorf("Expected 1 proposal, got %v", stats["total_proposals"])
	}
	if stats["open_proposals"] != 1 {
		t.Errorf("Expected 1 open proposal, got %v", stats["open_proposals"])
	}
}

// ========== INTEGRATION TESTS ==========

func TestFullAssemblyAndExecution(t *testing.T) {
	// Assemble factorial program
	source := `MOVI R0, 5
MOVI R1, 1
MOVI R2, 0
loop:
	INC R2
	IMUL R1, R2
	CMP R0, R2
	JNZ loop
end:
	HALT`
	
	bytecode, err := AssembleString(source)
	if err != nil {
		t.Fatalf("Assembly failed: %v", err)
	}
	
	// Verify bytecode
	errors := VerifyBytecode(bytecode)
	if len(errors) != 0 {
		t.Fatalf("Bytecode verification failed: %v", errors)
	}
	
	// Create VM and execute
	vm := NewFluxVM(bytecode)
	
	for vm.Running {
		err := vm.ExecuteStep()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
	}
	
	// R1 should contain 5! = 120
	regs := vm.GetRegisters()
	if regs[1] != 120 {
		t.Errorf("Expected R1=120 (5!), got %d", regs[1])
	}
}

func TestSwarmWithBytecode(t *testing.T) {
	swarm := NewSwarmCoordinator()
	
	// Assemble bytecode
	source := `MOVI R0, 10
MOVI R1, 0
MOVI R2, 0
loop:
	INC R1
	IADD R2, R1
	CMP R1, R0
	JNZ loop
	HALT`
	
	bytecode, err := AssembleString(source)
	if err != nil {
		t.Fatalf("Assembly failed: %v", err)
	}
	
	// Spawn agents with bytecode
	agent := swarm.SpawnAgent(Worker, bytecode)
	if agent == nil {
		t.Fatal("Failed to spawn agent")
	}
	
	// Run agent
	swarm.RunTimestep()
	
	// Run until complete
	for agent.VM.Running {
		agent.Process()
	}
	
	// R2 should contain sum 1+2+...+10 = 55
	regs := agent.GetResult()
	if regs[2] != 55 {
		t.Errorf("Expected R2=55, got %d", regs[2])
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

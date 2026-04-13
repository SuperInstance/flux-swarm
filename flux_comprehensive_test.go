package main

import (
        "encoding/json"
        "strings"
        "testing"
)

// ========== A2A MESSAGE TESTS ==========

func TestMessageSerialization(t *testing.T) {
        msg := A2AMessage{
                SenderID:   "agent-001",
                ReceiverID: "agent-002",
                Type:       TELL,
                Payload:    "hello world",
                Trust:      1.5,
        }

        data, err := msg.Serialize()
        if err != nil {
                t.Fatalf("Failed to serialize message: %v", err)
        }

        if len(data) == 0 {
                t.Fatal("Serialized data is empty")
        }

        // Verify it's valid JSON
        var parsed map[string]interface{}
        if err := json.Unmarshal(data, &parsed); err != nil {
                t.Fatalf("Serialized data is not valid JSON: %v", err)
        }

        if parsed["sender_id"] != "agent-001" {
                t.Errorf("Expected sender_id 'agent-001', got %v", parsed["sender_id"])
        }
}

func TestMessageDeserialization(t *testing.T) {
        original := A2AMessage{
                SenderID:   "agent-001",
                ReceiverID: "agent-002",
                Type:       ASK,
                Payload:    "what is 2+2?",
                Trust:      0.9,
        }

        data, _ := original.Serialize()
        parsed, err := DeserializeMessage(data)
        if err != nil {
                t.Fatalf("Failed to deserialize message: %v", err)
        }

        if parsed.SenderID != original.SenderID {
                t.Errorf("Expected SenderID '%s', got '%s'", original.SenderID, parsed.SenderID)
        }
        if parsed.ReceiverID != original.ReceiverID {
                t.Errorf("Expected ReceiverID '%s', got '%s'", original.ReceiverID, parsed.ReceiverID)
        }
        if parsed.Type != original.Type {
                t.Errorf("Expected Type '%s', got '%s'", original.Type, parsed.Type)
        }
        if parsed.Payload != original.Payload {
                t.Errorf("Expected Payload '%s', got '%s'", original.Payload, parsed.Payload)
        }
}

func TestDeserializeMessageInvalid(t *testing.T) {
        _, err := DeserializeMessage([]byte("not json"))
        if err == nil {
                t.Error("Expected error for invalid JSON")
        }
}

func TestAllMessageTypes(t *testing.T) {
        types := []MessageType{TELL, ASK, DELEGATE, BROADCAST}
        for _, mt := range types {
                msg := A2AMessage{Type: mt, Payload: "test"}
                data, err := msg.Serialize()
                if err != nil {
                        t.Errorf("Failed to serialize message type %s: %v", mt, err)
                }
                parsed, err := DeserializeMessage(data)
                if err != nil {
                        t.Errorf("Failed to deserialize message type %s: %v", mt, err)
                }
                if parsed.Type != mt {
                        t.Errorf("Expected type %s, got %s", mt, parsed.Type)
                }
        }
}

// ========== VM EDGE CASE TESTS ==========

func TestVMHalt(t *testing.T) {
        vm := NewFluxVM([]byte{HALT})
        err := vm.ExecuteStep()
        if err != nil {
                t.Fatalf("HALT should not produce error: %v", err)
        }
        if vm.Running {
                t.Error("VM should not be running after HALT")
        }
}

func TestVMNOP(t *testing.T) {
        vm := NewFluxVM([]byte{NOP, NOP, HALT})
        steps := 0
        for vm.Running {
                vm.ExecuteStep()
                steps++
        }
        if steps != 3 {
                t.Errorf("Expected 3 steps (NOP, NOP, HALT), got %d", steps)
        }
}

func TestVMSubtraction(t *testing.T) {
        vm := NewFluxVM([]byte{MOVI, 0, 10, MOVI, 1, 3, ISUB, 0, 1, HALT})
        for vm.Running {
                vm.ExecuteStep()
        }
        regs := vm.GetRegisters()
        if regs[0] != 7 {
                t.Errorf("Expected R0=7 (10-3), got %d", regs[0])
        }
}

func TestVMMultiplication(t *testing.T) {
        vm := NewFluxVM([]byte{MOVI, 0, 6, MOVI, 1, 7, IMUL, 0, 1, HALT})
        for vm.Running {
                vm.ExecuteStep()
        }
        regs := vm.GetRegisters()
        if regs[0] != 42 {
                t.Errorf("Expected R0=42 (6*7), got %d", regs[0])
        }
}

func TestVMDivision(t *testing.T) {
        vm := NewFluxVM([]byte{MOVI, 0, 100, MOVI, 1, 5, IDIV, 0, 1, HALT})
        for vm.Running {
                vm.ExecuteStep()
        }
        regs := vm.GetRegisters()
        if regs[0] != 20 {
                t.Errorf("Expected R0=20 (100/5), got %d", regs[0])
        }
}

func TestVMDivisionByZero(t *testing.T) {
        vm := NewFluxVM([]byte{MOVI, 0, 10, MOVI, 1, 0, IDIV, 0, 1, HALT})
        for vm.Running {
                err := vm.ExecuteStep()
                if err != nil && strings.Contains(err.Error(), "division by zero") {
                        return // Success
                }
        }
        t.Error("Expected division by zero error")
}

func TestVMIncrement(t *testing.T) {
        vm := NewFluxVM([]byte{MOVI, 0, 0, INC, 0, INC, 0, INC, 0, HALT})
        for vm.Running {
                vm.ExecuteStep()
        }
        regs := vm.GetRegisters()
        if regs[0] != 3 {
                t.Errorf("Expected R0=3 after 3 INC, got %d", regs[0])
        }
}

func TestVMDecrement(t *testing.T) {
        vm := NewFluxVM([]byte{MOVI, 0, 5, DEC, 0, DEC, 0, HALT})
        for vm.Running {
                vm.ExecuteStep()
        }
        regs := vm.GetRegisters()
        if regs[0] != 3 {
                t.Errorf("Expected R0=3 after 2 DEC, got %d", regs[0])
        }
}

func TestVMCompareEqual(t *testing.T) {
        vm := NewFluxVM([]byte{MOVI, 0, 5, MOVI, 1, 5, CMP, 0, 1, HALT})
        for vm.Running {
                vm.ExecuteStep()
        }
        regs := vm.GetRegisters()
        if regs[15] != 0 {
                t.Errorf("Expected R15=0 (equal), got %d", regs[15])
        }
}

func TestVMCompareGreater(t *testing.T) {
        vm := NewFluxVM([]byte{MOVI, 0, 10, MOVI, 1, 5, CMP, 0, 1, HALT})
        for vm.Running {
                vm.ExecuteStep()
        }
        regs := vm.GetRegisters()
        if regs[15] != 1 {
                t.Errorf("Expected R15=1 (greater), got %d", regs[15])
        }
}

func TestVMCompareLess(t *testing.T) {
        vm := NewFluxVM([]byte{MOVI, 0, 3, MOVI, 1, 7, CMP, 0, 1, HALT})
        for vm.Running {
                vm.ExecuteStep()
        }
        regs := vm.GetRegisters()
        if regs[15] != -1 {
                t.Errorf("Expected R15=-1 (less), got %d", regs[15])
        }
}

func TestVMJump(t *testing.T) {
        // JMP to address 3, skip HALT at 2
        vm := NewFluxVM([]byte{JMP, 3, HALT, MOVI, 0, 42, HALT})
        for vm.Running {
                vm.ExecuteStep()
        }
        regs := vm.GetRegisters()
        // Should have jumped over HALT at position 2 to MOVI at position 3
        if regs[0] != 42 {
                t.Errorf("Expected R0=42, got %d", regs[0])
        }
}

func TestVMOutOfBytes(t *testing.T) {
        vm := NewFluxVM([]byte{MOVI, 0}) // Missing operand
        err := vm.ExecuteStep()
        if err == nil {
                t.Error("Expected error for truncated bytecode")
        }
}

func TestVMUnknownOpcode(t *testing.T) {
        vm := NewFluxVM([]byte{0xFF}) // Unknown opcode
        err := vm.ExecuteStep()
        if err == nil {
                t.Error("Expected error for unknown opcode")
        }
}

func TestVMEmptyBytecode(t *testing.T) {
        vm := NewFluxVM([]byte{})
        err := vm.ExecuteStep()
        if err == nil {
                t.Error("Expected error for empty bytecode")
        }
}

// ========== ASSEMBLER EDGE CASE TESTS ==========

func TestAssemblerNOP(t *testing.T) {
        source := "NOP"
        bytecode, err := AssembleString(source)
        if err != nil {
                t.Fatalf("Assembly failed: %v", err)
        }
        if len(bytecode) != 1 || bytecode[0] != NOP {
                t.Errorf("Expected single NOP byte, got %v", bytecode)
        }
}

func TestAssemblerHALTOnly(t *testing.T) {
        source := "HALT"
        bytecode, err := AssembleString(source)
        if err != nil {
                t.Fatalf("Assembly failed: %v", err)
        }
        if len(bytecode) != 1 || bytecode[0] != HALT {
                t.Errorf("Expected single HALT byte, got %v", bytecode)
        }
}

func TestAssemblerComments(t *testing.T) {
        source := `; This is a comment
MOVI R0, 5 ; inline comment
HALT ; end`
        bytecode, err := AssembleString(source)
        if err != nil {
                t.Fatalf("Assembly with comments failed: %v", err)
        }
        if len(bytecode) < 3 {
                t.Errorf("Expected at least 3 bytes, got %d", len(bytecode))
        }
}

func TestAssemblerEmptySource(t *testing.T) {
        source := ""
        bytecode, err := AssembleString(source)
        if err != nil {
                t.Fatalf("Assembly of empty source should succeed: %v", err)
        }
        if len(bytecode) != 0 {
                t.Errorf("Expected empty bytecode, got %d bytes", len(bytecode))
        }
}

func TestAssemblerWhitespaceOnly(t *testing.T) {
        source := "   \n  \n  "
        bytecode, err := AssembleString(source)
        if err != nil {
                t.Fatalf("Assembly of whitespace should succeed: %v", err)
        }
        if len(bytecode) != 0 {
                t.Errorf("Expected empty bytecode, got %d bytes", len(bytecode))
        }
}

func TestAssemblerAllInstructions(t *testing.T) {
        source := `NOP
MOVI R0, 0
MOV R1, R0
IADD R0, R1
ISUB R0, R1
IMUL R0, R1
IDIV R0, R1
INC R0
DEC R0
CMP R0, R1
JZ 0
JNZ 0
JMP 0
HALT`
        bytecode, err := AssembleString(source)
        if err != nil {
                t.Fatalf("Assembly of all instructions failed: %v", err)
        }
        if len(bytecode) == 0 {
                t.Error("Expected non-empty bytecode")
        }
}

func TestParseRegister(t *testing.T) {
        tests := []struct {
                input   string
                expect  int
                wantErr bool
        }{
                {"R0", 0, false},
                {"R15", 15, false},
                {"r5", 5, false},
                {"R16", -1, true}, // out of range
                {"X0", -1, true},  // invalid prefix
                {"", -1, true},    // empty
        }

        for _, tc := range tests {
                reg, err := ParseRegister(tc.input)
                if tc.wantErr {
                        if err == nil {
                                t.Errorf("ParseRegister(%q): expected error, got reg=%d", tc.input, reg)
                        }
                } else {
                        if err != nil {
                                t.Errorf("ParseRegister(%q): unexpected error: %v", tc.input, err)
                        }
                        if reg != tc.expect {
                                t.Errorf("ParseRegister(%q): expected %d, got %d", tc.input, tc.expect, reg)
                        }
                }
        }
}

// ========== DISASSEMBLER EDGE CASES ==========

func TestDisassembleEmpty(t *testing.T) {
        output := DisassembleString([]byte{})
        if output != "" {
                t.Errorf("Expected empty output for empty bytecode, got %q", output)
        }
}

func TestDisassembleSingleHALT(t *testing.T) {
        output := DisassembleString([]byte{HALT})
        if !strings.Contains(output, "HALT") {
                t.Error("Expected HALT in disassembly")
        }
}

func TestFormatBytecode(t *testing.T) {
        bytecode := []byte{MOVI, 0, 5, IADD, 0, 1, HALT}
        output := FormatBytecode(bytecode)

        if !strings.Contains(output, "ADDRESS") {
                t.Error("Expected header in formatted output")
        }
        if !strings.Contains(output, "MOVI") {
                t.Error("Expected MOVI in formatted output")
        }
        if !strings.Contains(output, "HALT") {
                t.Error("Expected HALT in formatted output")
        }
}

func TestGetInstructionAt(t *testing.T) {
        bytecode := []byte{MOVI, 0, 5, HALT}

        d := NewDisassembler(bytecode)

        // Valid address
        instr, err := d.GetInstructionAt(0)
        if err != nil {
                t.Fatalf("GetInstructionAt(0) failed: %v", err)
        }
        if !strings.Contains(instr, "MOVI") {
                t.Errorf("Expected MOVI at address 0, got %s", instr)
        }

        // Out of bounds
        _, err = d.GetInstructionAt(100)
        if err == nil {
                t.Error("Expected error for out-of-bounds address")
        }

        // Negative address
        _, err = d.GetInstructionAt(-1)
        if err == nil {
                t.Error("Expected error for negative address")
        }
}

// ========== SWARM MESSAGING TESTS ==========

func TestBroadcastMessage(t *testing.T) {
        swarm := NewSwarmCoordinator()

        sender := NewAgent(Worker, []byte{HALT})
        receiver1 := NewAgent(Scout, []byte{HALT})
        receiver2 := NewAgent(Coordinator, []byte{HALT})

        swarm.RegisterAgent(sender)
        swarm.RegisterAgent(receiver1)
        swarm.RegisterAgent(receiver2)

        swarm.BroadcastMessage(sender.ID, TELL, "Hello all")

        if swarm.GetTotalMessages() != 2 {
                t.Errorf("Expected 2 messages after broadcast, got %d", swarm.GetTotalMessages())
        }
}

func TestRouteMessageNonExistent(t *testing.T) {
        swarm := NewSwarmCoordinator()

        agent := NewAgent(Worker, []byte{HALT})
        swarm.RegisterAgent(agent)

        success := swarm.RouteMessage(agent.ID, "non-existent", TELL, "test")
        if success {
                t.Error("Routing to non-existent agent should fail")
        }
}

func TestRemoveAgent(t *testing.T) {
        swarm := NewSwarmCoordinator()

        agent := NewAgent(Worker, []byte{HALT})
        swarm.RegisterAgent(agent)

        swarm.RemoveAgent(agent.ID)

        if _, exists := swarm.Agents[agent.ID]; exists {
                t.Error("Agent should be removed")
        }

        if _, exists := swarm.TrustMatrix[agent.ID]; exists {
                t.Error("Agent trust should be removed")
        }
}

func TestCollectResults(t *testing.T) {
        swarm := NewSwarmCoordinator()

        bytecode := []byte{MOVI, 0, 42, HALT}
        agent := swarm.SpawnAgent(Worker, bytecode)

        for agent.VM.Running {
                agent.Process()
        }

        results := swarm.CollectResults()
        if len(results) != 1 {
                t.Fatalf("Expected 1 result, got %d", len(results))
        }

        for _, regs := range results {
                if regs[0] != 42 {
                        t.Errorf("Expected R0=42, got %d", regs[0])
                }
        }
}

// ========== MAX AGENTS TEST ==========

func TestMaxAgentsLimit(t *testing.T) {
        swarm := NewSwarmCoordinator()
        swarm.Lifecycle.MaxAgents = 3

        swarm.SpawnAgent(Worker, []byte{HALT})
        swarm.SpawnAgent(Worker, []byte{HALT})
        swarm.SpawnAgent(Worker, []byte{HALT})

        // This should fail - max reached
        agent := swarm.SpawnAgent(Worker, []byte{HALT})
        if agent != nil {
                t.Error("Expected nil when max agents reached")
        }
}

// ========== PROPOSAL EDGE CASES ==========

func TestVoteOnNonExistentProposal(t *testing.T) {
        swarm := NewSwarmCoordinator()

        agent := swarm.SpawnAgent(Worker, []byte{HALT})

        err := swarm.VoteOnProposal(agent.ID, "non-existent", true)
        if err == nil {
                t.Error("Expected error for voting on non-existent proposal")
        }
}

func TestGetAllProposals(t *testing.T) {
        swarm := NewSwarmCoordinator()

        agent := swarm.SpawnAgent(Worker, []byte{HALT})

        swarm.CreateProposal(agent.ID, TaskAssignment, "Task 1")
        swarm.CreateProposal(agent.ID, ResourceShare, "Share resource")

        proposals := swarm.GetAllProposals()
        if len(proposals) != 2 {
                t.Errorf("Expected 2 proposals, got %d", len(proposals))
        }
}

func TestGetNonExistentProposal(t *testing.T) {
        swarm := NewSwarmCoordinator()

        _, err := swarm.GetProposal("non-existent")
        if err == nil {
                t.Error("Expected error for non-existent proposal")
        }
}

func TestProposalTypes(t *testing.T) {
        types := []ProposalType{TaskAssignment, ResourceShare, TopologyChange, Evolution}
        swarm := NewSwarmCoordinator()

        agent := swarm.SpawnAgent(Worker, []byte{HALT})

        for _, ptype := range types {
                pid := swarm.CreateProposal(agent.ID, ptype, "test")
                proposal, err := swarm.GetProposal(pid)
                if err != nil {
                        t.Errorf("Failed to get proposal of type %s: %v", ptype, err)
                }
                if proposal.Type != ptype {
                        t.Errorf("Expected type %s, got %s", ptype, proposal.Type)
                }
        }
}

func TestDuplicateVote(t *testing.T) {
        swarm := NewSwarmCoordinator()

        agent := swarm.SpawnAgent(Worker, []byte{HALT})

        pid := swarm.CreateProposal(agent.ID, TaskAssignment, "test")

        // First vote should succeed
        err := swarm.VoteOnProposal(agent.ID, pid, true)
        if err != nil {
                t.Errorf("First vote should succeed: %v", err)
        }

        // Second vote should fail
        err = swarm.VoteOnProposal(agent.ID, pid, false)
        if err == nil {
                t.Error("Expected error for duplicate vote")
        }
}

// ========== VOCABULARY ADDITIONAL TESTS ==========

func TestVocabularyDocstrings(t *testing.T) {
        vocab := NewVocabulary()

        doc := vocab.GetDocstring("IADD")
        if doc == "" {
                t.Error("Expected docstring for IADD")
        }
        if !strings.Contains(doc, "add") {
                t.Errorf("Expected 'add' in IADD docstring, got: %s", doc)
        }
}

func TestVocabularyAllDocstrings(t *testing.T) {
        vocab := NewVocabulary()

        opcodes := []string{"NOP", "MOV", "MOVI", "IADD", "ISUB", "IMUL", "IDIV",
                "INC", "DEC", "CMP", "JZ", "JNZ", "JMP", "HALT"}

        for _, op := range opcodes {
                doc := vocab.GetDocstring(op)
                if doc == "" {
                        t.Errorf("Expected docstring for %s", op)
                }
        }
}

func TestVocabularyResolveAlias(t *testing.T) {
        vocab := NewVocabulary()

        tests := []struct {
                alias    string
                expected string
                found    bool
        }{
                {"ADD", "IADD", true},
                {"SUB", "ISUB", true},
                {"MUL", "IMUL", true},
                {"DIV", "IDIV", true},
                {"LOAD", "MOVI", true},
                {"JE", "JZ", true},
                {"JNE", "JNZ", true},
                {"STOP", "HALT", true},
                {"NONEXISTENT", "", false},
        }

        for _, tc := range tests {
                resolved, ok := vocab.ResolveAlias(tc.alias)
                if ok != tc.found {
                        t.Errorf("ResolveAlias(%q): expected found=%v, got %v", tc.alias, tc.found, ok)
                        continue
                }
                if tc.found && resolved != tc.expected {
                        t.Errorf("ResolveAlias(%q): expected %s, got %s", tc.alias, tc.expected, resolved)
                }
        }
}

func TestVocabularyAddCustomMacro(t *testing.T) {
        vocab := NewVocabulary()

        vocab.AddMacro("CUSTOM_MUL3", "MOVI {dest}, 0\nIADD {dest}, {a}\nIADD {dest}, {b}\nIADD {dest}, {c}",
                []string{"dest", "a", "b", "c"}, "Add three values")

        macro, ok := vocab.GetMacro("CUSTOM_MUL3")
        if !ok {
                t.Fatal("Custom macro not found")
        }
        if macro.Name != "CUSTOM_MUL3" {
                t.Errorf("Expected name CUSTOM_MUL3, got %s", macro.Name)
        }

        expanded, err := vocab.ExpandMacro("CUSTOM_MUL3", []string{"R0", "1", "2", "3"})
        if err != nil {
                t.Fatalf("Failed to expand custom macro: %v", err)
        }
        if !strings.Contains(expanded, "R0") {
                t.Error("Expected R0 in expanded macro")
        }
}

func TestVocabularyMacroWrongArgs(t *testing.T) {
        vocab := NewVocabulary()

        _, err := vocab.ExpandMacro("CLEAR", []string{})
        if err == nil {
                t.Error("Expected error for wrong number of args")
        }
}

func TestVocabularyNonExistentMacro(t *testing.T) {
        vocab := NewVocabulary()

        _, err := vocab.ExpandMacro("NONEXISTENT", []string{"arg1"})
        if err == nil {
                t.Error("Expected error for non-existent macro")
        }
}

func TestVocabularyListInstructions(t *testing.T) {
        vocab := NewVocabulary()

        instructions := vocab.ListInstructions()
        if len(instructions) == 0 {
                t.Fatal("No instructions listed")
        }

        // Check for some expected instructions
        foundMOVI := false
        foundMacro := false
        for _, instr := range instructions {
                if instr == "MOVI" {
                        foundMOVI = true
                }
                if strings.Contains(instr, "(macro)") {
                        foundMacro = true
                }
        }
        if !foundMOVI {
                t.Error("MOVI not found in instruction list")
        }
        if !foundMacro {
                t.Error("No macros found in instruction list")
        }
}

func TestVocabularyPatternRecognition(t *testing.T) {
        vocab := NewVocabulary()

        // Test subtraction pattern
        _, ok := vocab.RecognizePattern("I want to subtract numbers")
        if !ok {
                t.Error("Failed to recognize subtract pattern")
        }

        // Test multiply pattern
        _, ok = vocab.RecognizePattern("multiply two values")
        if !ok {
                t.Error("Failed to recognize multiply pattern")
        }

        // Test divide pattern
        _, ok = vocab.RecognizePattern("divide by zero")
        if !ok {
                t.Error("Failed to recognize divide pattern")
        }

        // Test add pattern
        _, ok = vocab.RecognizePattern("add these together")
        if !ok {
                t.Error("Failed to recognize add pattern")
        }
}

func TestVocabularyPatternExpansionHello(t *testing.T) {
        vocab := NewVocabulary()

        expanded, err := vocab.ExpandPattern("hello", nil)
        if err != nil {
                t.Fatalf("Failed to expand hello pattern: %v", err)
        }
        if !strings.Contains(expanded, "HALT") {
                t.Error("Expected HALT in hello pattern expansion")
        }
}

func TestVocabularyPatternExpansionUnknown(t *testing.T) {
        vocab := NewVocabulary()

        _, err := vocab.ExpandPattern("unknown_pattern", nil)
        if err == nil {
                t.Error("Expected error for unknown pattern")
        }
}

// ========== AGENT ROLE TESTS ==========

func TestAgentRoles(t *testing.T) {
        roles := []Role{Worker, Scout, Coordinator, Specialist}
        for _, role := range roles {
                agent := NewAgent(role, []byte{HALT})
                if agent.Role != role {
                        t.Errorf("Expected role %s, got %s", role, agent.Role)
                }
        }
}

func TestAgentDefaultTrust(t *testing.T) {
        agent := NewAgent(Worker, []byte{HALT})
        if agent.Trust != 1.0 {
                t.Errorf("Expected default trust 1.0, got %f", agent.Trust)
        }
}

func TestAgentMessageQueue(t *testing.T) {
        agent := NewAgent(Worker, []byte{HALT})

        // Queue should be empty
        msg := agent.GetNextMessage()
        if msg != nil {
                t.Error("Expected nil from empty queue")
        }

        // Add messages
        agent.ReceiveMessage(A2AMessage{Payload: "msg1"})
        agent.ReceiveMessage(A2AMessage{Payload: "msg2"})

        msg = agent.GetNextMessage()
        if msg == nil || msg.Payload != "msg1" {
                t.Error("Expected first message 'msg1'")
        }

        msg = agent.GetNextMessage()
        if msg == nil || msg.Payload != "msg2" {
                t.Error("Expected second message 'msg2'")
        }

        // Queue should be empty again
        msg = agent.GetNextMessage()
        if msg != nil {
                t.Error("Expected nil after consuming all messages")
        }
}

func TestAgentSendMessageFormat(t *testing.T) {
        sender := NewAgent(Worker, []byte{HALT})
        msg := sender.SendMessage("receiver-id", DELEGATE, "do task")

        if msg.SenderID != sender.ID {
                t.Error("Message sender ID mismatch")
        }
        if msg.ReceiverID != "receiver-id" {
                t.Error("Message receiver ID mismatch")
        }
        if msg.Type != DELEGATE {
                t.Error("Message type mismatch")
        }
        if msg.Payload != "do task" {
                t.Error("Message payload mismatch")
        }
        if msg.Trust != sender.Trust {
                t.Error("Message trust mismatch")
        }
}

// ========== INTEGRATION: FULL PROGRAM TESTS ==========

func TestFullSubtractionProgram(t *testing.T) {
        source := `MOVI R0, 100
MOVI R1, 37
ISUB R0, R1
HALT`

        bytecode, err := AssembleString(source)
        if err != nil {
                t.Fatalf("Assembly failed: %v", err)
        }

        vm := NewFluxVM(bytecode)
        for vm.Running {
                vm.ExecuteStep()
        }

        regs := vm.GetRegisters()
        if regs[0] != 63 {
                t.Errorf("Expected R0=63 (100-37), got %d", regs[0])
        }
}

func TestFullMultiplicationProgram(t *testing.T) {
        source := `MOVI R0, 7
MOVI R1, 8
IMUL R0, R1
HALT`

        bytecode, err := AssembleString(source)
        if err != nil {
                t.Fatalf("Assembly failed: %v", err)
        }

        vm := NewFluxVM(bytecode)
        for vm.Running {
                vm.ExecuteStep()
        }

        regs := vm.GetRegisters()
        if regs[0] != 56 {
                t.Errorf("Expected R0=56 (7*8), got %d", regs[0])
        }
}

func TestVerifyBytecodeEmpty(t *testing.T) {
        errors := VerifyBytecode([]byte{})
        if len(errors) != 0 {
                t.Errorf("Empty bytecode should produce no errors, got %v", errors)
        }
}

func TestVerifyBytecodeInvalidOpcode(t *testing.T) {
        errors := VerifyBytecode([]byte{0xFE})
        if len(errors) == 0 {
                t.Error("Expected error for invalid opcode")
        }
}

func TestVerifyBytecodeIncomplete(t *testing.T) {
        // MOVI needs 2 operands but only has 1
        errors := VerifyBytecode([]byte{MOVI, 0})
        if len(errors) == 0 {
                t.Error("Expected error for incomplete instruction")
        }
}

func TestVerifyBytecodeAllValid(t *testing.T) {
        source := `MOVI R0, 10
MOVI R1, 20
IADD R0, R1
MOV R2, R0
HALT`

        bytecode, _ := AssembleString(source)
        errors := VerifyBytecode(bytecode)
        if len(errors) != 0 {
                t.Errorf("Expected no errors for valid bytecode, got %v", errors)
        }
}

// ========== ROUND-TRIP ASSEMBLE/DISASSEMBLE ==========

func TestRoundTripDisassembly(t *testing.T) {
        source := `MOVI R0, 42
MOVI R1, 7
IADD R0, R1
HALT`

        bytecode, err := AssembleString(source)
        if err != nil {
                t.Fatalf("Assembly failed: %v", err)
        }

        output := DisassembleString(bytecode)

        // Check that all instructions appear in disassembly
        expectedMnemonics := []string{"MOVI", "IADD", "HALT"}
        for _, mnemonic := range expectedMnemonics {
                if !strings.Contains(output, mnemonic) {
                        t.Errorf("Expected %s in disassembly output", mnemonic)
                }
        }
}

// ========== SWARM LIFECYCLE PROCESS ==========

func TestProcessLifecycle(t *testing.T) {
        swarm := NewSwarmCoordinator()

        swarm.SpawnAgent(Worker, []byte{HALT})
        swarm.SpawnAgent(Scout, []byte{HALT})
        swarm.SpawnAgent(Coordinator, []byte{HALT})

        swarm.ProcessLifecycle()

        stats := swarm.GetLifecycleStats()
        if stats["timestep_count"] != 1 {
                t.Errorf("Expected timestep_count=1, got %v", stats["timestep_count"])
        }
}

func TestEvolveNonExistentAgent(t *testing.T) {
        swarm := NewSwarmCoordinator()

        err := swarm.EvolveAgent("non-existent")
        if err == nil {
                t.Error("Expected error for evolving non-existent agent")
        }
}

func TestTerminateNonExistentAgent(t *testing.T) {
        swarm := NewSwarmCoordinator()

        err := swarm.TerminateAgent("non-existent")
        if err == nil {
                t.Error("Expected error for terminating non-existent agent")
        }
}

func TestCleanupExpiredProposals(t *testing.T) {
        swarm := NewSwarmCoordinator()

        agent := swarm.SpawnAgent(Worker, []byte{HALT})

        // Create proposal and immediately expire it
        pid := swarm.CreateProposal(agent.ID, TaskAssignment, "test")

        // Get the proposal and manually expire it
        proposal, _ := swarm.GetProposal(pid)
        proposal.ExpiresAt = proposal.ExpiresAt.Add(-10 * 1000 * 1000 * 1000) // 10 seconds ago

        swarm.CleanupExpiredProposals()

        proposal, _ = swarm.GetProposal(pid)
        if proposal.Status != ProposalExpired {
                t.Errorf("Expected proposal status expired, got %s", proposal.Status)
        }
}

package main

import "testing"

// ========== DISASSEMBLER COMPREHENSIVE TESTS ==========

func TestDisassembleInstruction(t *testing.T) {
	tests := []struct {
		name    string
		bc      []byte
		offset  int
		want    string
		wantLen int
	}{
		{"HALT at start", []byte{HALT}, 0, "HALT", 1},
		{"NOP at start", []byte{NOP}, 0, "NOP", 1},
		{"MOVI shows register prefix", []byte{MOVI, 0, 42}, 0, "MOVI R0", 3},
		{"IADD", []byte{IADD, 0, 1}, 0, "IADD R0, R1", 3},
		{"INC", []byte{INC, 5}, 0, "INC R5", 2},
		{"DEC", []byte{DEC, 3}, 0, "DEC R3", 2},
		{"JMP", []byte{JMP, 10}, 0, "JMP 10", 2},
		{"JZ", []byte{JZ, 5}, 0, "JZ 5", 2},
		{"JNZ", []byte{JNZ, 7}, 0, "JNZ 7", 2},
		{"CMP", []byte{CMP, 0, 1}, 0, "CMP R0, R1", 3},
		{"unknown opcode", []byte{0xFF}, 0, "UNKNOWN", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDisassembler(tt.bc)
			got, gotLen := d.DisassembleInstruction(tt.offset)

			if gotLen != tt.wantLen {
				t.Errorf("DisassembleInstruction() len = %d, want %d", gotLen, tt.wantLen)
			}
			if !containsSubstring(got, tt.want) {
				t.Errorf("DisassembleInstruction() = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

func TestDisassembleInstructionOffset(t *testing.T) {
	bc := []byte{HALT, MOVI, 0, 42, MOVI, 1, 99}

	d := NewDisassembler(bc)

	// Offset 0: HALT
	got, size := d.DisassembleInstruction(0)
	if !containsSubstring(got, "HALT") || size != 1 {
		t.Errorf("Offset 0: got %q, size %d", got, size)
	}

	// Offset 1: MOVI R0, R42 (disassembler shows R prefix for immediate values too)
	got, size = d.DisassembleInstruction(1)
	if !containsSubstring(got, "MOVI R0") || size != 3 {
		t.Errorf("Offset 1: got %q, size %d", got, size)
	}

	// Offset 4: MOVI R1, R99
	got, size = d.DisassembleInstruction(4)
	if !containsSubstring(got, "MOVI R1") || size != 3 {
		t.Errorf("Offset 4: got %q, size %d", got, size)
	}
}

func TestGetInstructionAt(t *testing.T) {
	bc := []byte{MOVI, 0, 42, IADD, 0, 1, HALT}

	d := NewDisassembler(bc)

	// Valid addresses
	instr, err := d.GetInstructionAt(0)
	if err != nil {
		t.Fatalf("GetInstructionAt(0) error: %v", err)
	}
	if !containsSubstring(instr, "MOVI R0") {
		t.Errorf("GetInstructionAt(0) = %q", instr)
	}

	instr, err = d.GetInstructionAt(3)
	if err != nil {
		t.Fatalf("GetInstructionAt(3) error: %v", err)
	}
	if !containsSubstring(instr, "IADD R0, R1") {
		t.Errorf("GetInstructionAt(3) = %q", instr)
	}

	// Out of bounds
	_, err = d.GetInstructionAt(100)
	if err == nil {
		t.Error("Expected error for out of bounds address")
	}

	// Negative address
	_, err = d.GetInstructionAt(-1)
	if err == nil {
		t.Error("Expected error for negative address")
	}
}

func TestGetInstructionAtIncomplete(t *testing.T) {
	bc := []byte{MOVI, 0} // MOVI needs 3 bytes, only has 2

	d := NewDisassembler(bc)
	_, err := d.GetInstructionAt(0)
	if err == nil {
		t.Error("Expected error for incomplete instruction")
	}
}

func TestFormatBytecode(t *testing.T) {
	bc := []byte{MOVI, 0, 42, HALT}

	result := FormatBytecode(bc)

	if len(result) == 0 {
		t.Fatal("FormatBytecode returned empty string")
	}

	// Should contain the header
	if !containsSubstring(result, "ADDRESS") {
		t.Error("Expected header in FormatBytecode output")
	}
	if !containsSubstring(result, "MOVI R0") {
		t.Error("Expected 'MOVI R0' in formatted output")
	}
	if !containsSubstring(result, "HALT") {
		t.Error("Expected 'HALT' in formatted output")
	}
}

func TestFormatBytecodeEmpty(t *testing.T) {
	result := FormatBytecode([]byte{})
	// Should at least have the header
	if !containsSubstring(result, "ADDRESS") {
		t.Error("Expected header even for empty bytecode")
	}
}

func TestVerifyBytecodeInvalidOpcode(t *testing.T) {
	bc := []byte{0xFF, 0, 0}
	errors := VerifyBytecode(bc)
	if len(errors) == 0 {
		t.Error("Expected error for invalid opcode 0xFF")
	}
}

func TestVerifyBytecodeIncomplete(t *testing.T) {
	bc := []byte{MOVI, 0} // needs 3 bytes
	errors := VerifyBytecode(bc)
	if len(errors) == 0 {
		t.Error("Expected error for incomplete instruction")
	}
}

func TestDisassembleStringEmpty(t *testing.T) {
	result := DisassembleString([]byte{})
	if result != "" {
		t.Errorf("Expected empty string for empty bytecode, got %q", result)
	}
}

func TestDisassembleStringJumpOperands(t *testing.T) {
	// Jump instructions should show numeric addresses, not R prefixes
	bc := []byte{JMP, 10, JZ, 5, JNZ, 3}
	result := DisassembleString(bc)

	if !containsSubstring(result, "JMP 10") {
		t.Errorf("Expected 'JMP 10' in %q", result)
	}
	if !containsSubstring(result, "JZ 5") {
		t.Errorf("Expected 'JZ 5' in %q", result)
	}
	if !containsSubstring(result, "JNZ 3") {
		t.Errorf("Expected 'JNZ 3' in %q", result)
	}
}

func TestNewDisassembler(t *testing.T) {
	bc := []byte{HALT, MOVI, 0, 42}
	d := NewDisassembler(bc)

	if d == nil {
		t.Fatal("NewDisassembler returned nil")
	}
	if d.pc != 0 {
		t.Errorf("Expected pc=0, got %d", d.pc)
	}
	if len(d.bytecode) != 4 {
		t.Errorf("Expected 4 bytes, got %d", len(d.bytecode))
	}
}

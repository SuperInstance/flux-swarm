package main

import "testing"

// ========== ASSEMBLER COMPREHENSIVE TESTS ==========

func TestParseRegister(t *testing.T) {
	tests := []struct {
		input  string
		want   int
		wantOK bool
	}{
		{"R0", 0, true},
		{"R1", 1, true},
		{"R15", 15, true},
		{"r0", 0, true},
		{"R7", 7, true},
		{"R16", -1, false}, // out of range
		{"R-1", -1, false}, // invalid
		{"", -1, false},    // empty
		{"RX", -1, false},  // non-numeric
		{"123", 123, true}, // plain number
		{"0", 0, true},     // plain number zero
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseRegister(tt.input)
			if tt.wantOK {
				if err != nil {
					t.Errorf("ParseRegister(%q) unexpected error: %v", tt.input, err)
				}
				if got != tt.want {
					t.Errorf("ParseRegister(%q) = %d, want %d", tt.input, got, tt.want)
				}
			} else {
				if err == nil {
					t.Errorf("ParseRegister(%q) expected error, got nil", tt.input)
				}
			}
		})
	}
}

func TestAssemblerEmptySource(t *testing.T) {
	bytecode, err := AssembleString("")
	if err != nil {
		t.Fatalf("Empty source should assemble without error: %v", err)
	}
	if len(bytecode) != 0 {
		t.Errorf("Expected 0 bytes, got %d", len(bytecode))
	}
}

func TestAssemblerCommentsOnly(t *testing.T) {
	source := "; This is a comment\n; Another comment\n"
	bytecode, err := AssembleString(source)
	if err != nil {
		t.Fatalf("Comment-only source should assemble without error: %v", err)
	}
	if len(bytecode) != 0 {
		t.Errorf("Expected 0 bytes, got %d", len(bytecode))
	}
}

func TestAssemblerUndefinedLabel(t *testing.T) {
	source := "JMP undefined_label\n"
	_, err := AssembleString(source)
	if err == nil {
		t.Error("Expected error for undefined label")
	}
}

func TestAssemblerWrongOperandCount(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"too many ops for HALT", "HALT 1"},
		{"too few ops for MOVI", "MOVI R0"},
		{"too many ops for INC", "INC R0 R1"},
		{"too few ops for IADD", "IADD R0"},
		{"too many ops for JMP", "JMP 5 10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AssembleString(tt.source)
			if err == nil {
				t.Errorf("Expected error for %q", tt.source)
			}
		})
	}
}

func TestAssemblerInlineComment(t *testing.T) {
	source := `MOVI R0, 42 ; load constant
HALT ; stop execution`

	bytecode, err := AssembleString(source)
	if err != nil {
		t.Fatalf("Assembly with inline comments failed: %v", err)
	}

	if len(bytecode) < 4 {
		t.Errorf("Expected at least 4 bytes, got %d", len(bytecode))
	}

	vm := NewFluxVM(bytecode)
	for vm.Running {
		vm.ExecuteStep()
	}

	regs := vm.GetRegisters()
	if regs[0] != 42 {
		t.Errorf("Expected R0=42, got %d", regs[0])
	}
}

func TestAssemblerAllOpcodes(t *testing.T) {
	source := `NOP
MOVI R0, 10
MOVI R1, 20
MOV R2, R0
IADD R0, R1
ISUB R1, R0
IMUL R2, R0
IDIV R0, R1
INC R0
DEC R1
CMP R0, R1
MOVI R15, 1
JZ 100
JNZ 100
JMP 100
HALT`

	bytecode, err := AssembleString(source)
	if err != nil {
		t.Fatalf("Assembly of all opcodes failed: %v", err)
	}

	if len(bytecode) == 0 {
		t.Fatal("Expected non-empty bytecode")
	}
}

func TestAssemblerOperandOutOfRange(t *testing.T) {
	source := `MOVI R0, 300`
	_, err := AssembleString(source)
	if err == nil {
		t.Error("Expected error for operand out of range (>255)")
	}
}

func TestAssemblerNegativeOperand(t *testing.T) {
	source := `MOVI R0, -1`
	_, err := AssembleString(source)
	if err == nil {
		t.Error("Expected error for negative operand")
	}
}

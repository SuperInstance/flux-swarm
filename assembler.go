package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Assembler converts FLUX assembly code to bytecode
type Assembler struct {
	labels      map[string]int
	instructions []Instruction
	pc          int
}

// Instruction represents a parsed assembly instruction
type Instruction struct {
	Opcode         byte
	Operands       []int
	LabelRefs      []string // Label references to be resolved in second pass
	Label          string
	IsLabel        bool
	Comment        string
	Original       string
}

// NewAssembler creates a new assembler
func NewAssembler() *Assembler {
	return &Assembler{
		labels:     make(map[string]int),
		instructions: make([]Instruction, 0),
		pc:         0,
	}
}

// Assemble parses and compiles assembly code to bytecode
func (a *Assembler) Assemble(source string) ([]byte, error) {
	a.labels = make(map[string]int)
	a.instructions = make([]Instruction, 0)
	a.pc = 0

	// First pass: collect labels and parse instructions
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		// Remove comments
		if idx := strings.Index(line, ";"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		// Check for label definition
		if strings.HasSuffix(line, ":") {
			labelName := strings.TrimSuffix(line, ":")
			a.labels[labelName] = a.pc
			continue
		}

		// Parse instruction
		instr, err := a.parseInstruction(line)
		if err != nil {
			return nil, fmt.Errorf("parse error at line '%s': %w", line, err)
		}

		a.instructions = append(a.instructions, instr)
		a.pc += 1 + len(instr.Operands)
	}

	// Second pass: resolve labels and generate bytecode
	bytecode := make([]byte, 0, a.pc)
	for _, instr := range a.instructions {
		bytecode = append(bytecode, instr.Opcode)
		
		// Resolve label references
		labelIdx := 0
		for _, op := range instr.Operands {
			// Check if this operand is a placeholder for a label reference
			if labelIdx < len(instr.LabelRefs) {
				labelName := instr.LabelRefs[labelIdx]
				if labelAddr, ok := a.labels[labelName]; ok {
					op = labelAddr
					labelIdx++
				} else {
					return nil, fmt.Errorf("undefined label: %s", labelName)
				}
			}
			
			if op >= 256 || op < 0 {
				return nil, fmt.Errorf("operand out of range: %d", op)
			}
			bytecode = append(bytecode, byte(op))
		}
		
		// Verify all label references were resolved
		if labelIdx < len(instr.LabelRefs) {
			return nil, fmt.Errorf("not all label references resolved")
		}
	}

	return bytecode, nil
}

// parseInstruction parses a single instruction line
func (a *Assembler) parseInstruction(line string) (Instruction, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return Instruction{}, fmt.Errorf("empty instruction")
	}

	opcodeStr := strings.ToUpper(parts[0])
	opcode, ok := opcodeMap[opcodeStr]
	if !ok {
		return Instruction{}, fmt.Errorf("unknown opcode: %s", opcodeStr)
	}

	instr := Instruction{
		Opcode:    opcode,
		Operands:  make([]int, 0),
		LabelRefs: make([]string, 0),
		Original:  line,
	}

	// Parse operands based on opcode
	expectedOps := operandCount[opcode]
	if len(parts)-1 != expectedOps {
		return Instruction{}, fmt.Errorf("%s expects %d operands, got %d", opcodeStr, expectedOps, len(parts)-1)
	}

	for i := 1; i < len(parts); i++ {
		opStr := parts[i]
		
		// Trim trailing comma if present
		opStr = strings.TrimSuffix(opStr, ",")
		
		// Check if it's a label reference (will be resolved in second pass)
		if labelAddr, ok := a.labels[opStr]; ok {
			instr.Operands = append(instr.Operands, labelAddr)
			continue
		}

		// Store label reference for later resolution
		// If it looks like a label (not a number or register), save it
		_, regErr := ParseRegister(opStr)
		_, numErr := strconv.Atoi(opStr)
		if regErr != nil && numErr != nil {
			// It's likely a label reference - store it
			instr.LabelRefs = append(instr.LabelRefs, opStr)
			// Add placeholder operand
			instr.Operands = append(instr.Operands, 0)
			continue
		}

		// Try to parse as register reference (e.g., R0, R1)
		if regErr == nil {
			reg, _ := ParseRegister(opStr)
			instr.Operands = append(instr.Operands, reg)
			continue
		}

		// Parse as immediate value
		val, err := strconv.Atoi(opStr)
		if err != nil {
			return Instruction{}, fmt.Errorf("invalid operand: %s", opStr)
		}
		instr.Operands = append(instr.Operands, val)
	}

	return instr, nil
}

// opcodeMap maps mnemonic to opcode byte
var opcodeMap = map[string]byte{
	"NOP":  NOP,
	"MOV":  MOV,
	"MOVI": MOVI,
	"IADD": IADD,
	"ISUB": ISUB,
	"IMUL": IMUL,
	"IDIV": IDIV,
	"INC":  INC,
	"DEC":  DEC,
	"CMP":  CMP,
	"JZ":   JZ,
	"JNZ":  JNZ,
	"JMP":  JMP,
	"HALT": HALT,
}

// operandCount specifies the number of operands per opcode
var operandCount = map[byte]int{
	NOP:  0,
	MOV:  2, // rd, rs
	MOVI: 2, // rd, immediate
	IADD: 2, // rd, rs
	ISUB: 2, // rd, rs
	IMUL: 2, // rd, rs
	IDIV: 2, // rd, rs
	INC:  1, // rd
	DEC:  1, // rd
	CMP:  2, // ra, rb
	JZ:   1, // target
	JNZ:  1, // target
	JMP:  1, // target
	HALT: 0,
}

// AssembleString is a convenience function to assemble a string
func AssembleString(source string) ([]byte, error) {
	assembler := NewAssembler()
	return assembler.Assemble(source)
}

// ParseRegister parses a register reference (e.g., "R0", "R1")
func ParseRegister(ref string) (int, error) {
	re := regexp.MustCompile(`^[Rr](\d+)$|^\d+$`)
	if !re.MatchString(ref) {
		return -1, fmt.Errorf("invalid register reference: %s", ref)
	}

	// Extract the number part
	if strings.HasPrefix(strings.ToUpper(ref), "R") {
		numStr := ref[1:]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return -1, fmt.Errorf("invalid register number: %s", numStr)
		}
		if num < 0 || num > 15 {
			return -1, fmt.Errorf("register out of range: R%d (0-15)", num)
		}
		return num, nil
	}

	// Plain number, convert to int
	return strconv.Atoi(ref)
}

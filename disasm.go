package main

import (
	"fmt"
	"strings"
)

// Disassembler converts bytecode to assembly text
type Disassembler struct {
	bytecode []byte
	pc       int
}

// NewDisassembler creates a new disassembler
func NewDisassembler(bytecode []byte) *Disassembler {
	return &Disassembler{
		bytecode: bytecode,
		pc:       0,
	}
}

// Disassemble converts bytecode to assembly text
func (d *Disassembler) Disassemble() string {
	var sb strings.Builder

	for d.pc < len(d.bytecode) {
		instr := d.disassembleInstruction()
		sb.WriteString(instr)
		sb.WriteString("\n")
	}

	return sb.String()
}

// disassembleInstruction disassembles a single instruction
func (d *Disassembler) disassembleInstruction() string {
	if d.pc >= len(d.bytecode) {
		return "; END"
	}

	opcode := d.bytecode[d.pc]
	d.pc++

	mnemonic, ok := reverseOpcodeMap[opcode]
	if !ok {
		return fmt.Sprintf("; UNKNOWN opcode 0x%02X at PC %d", opcode, d.pc-1)
	}

	// Check for operand availability
	opCount := operandCount[opcode]
	if d.pc+opCount > len(d.bytecode) {
		return fmt.Sprintf("; INCOMPLETE instruction: %s at PC %d", mnemonic, d.pc-1)
	}

	var operands []string
	for i := 0; i < opCount; i++ {
		op := d.bytecode[d.pc]
		d.pc++
		
		// For jump targets, show as address
		if opcode == JMP || opcode == JZ || opcode == JNZ {
			operands = append(operands, fmt.Sprintf("%d", op))
		} else {
			operands = append(operands, fmt.Sprintf("R%d", op))
		}
	}

	if len(operands) == 0 {
		return mnemonic
	}

	return fmt.Sprintf("%s %s", mnemonic, strings.Join(operands, ", "))
}

// DisassembleInstruction disassembles a single instruction at given offset
func (d *Disassembler) DisassembleInstruction(offset int) (string, int) {
	if offset >= len(d.bytecode) {
		return "; END", 0
	}

	opcode := d.bytecode[offset]
	mnemonic, ok := reverseOpcodeMap[opcode]
	if !ok {
		return fmt.Sprintf("; UNKNOWN 0x%02X", opcode), 1
	}

	opCount := operandCount[opcode]
	if offset+1+opCount > len(d.bytecode) {
		return fmt.Sprintf("; INCOMPLETE: %s", mnemonic), len(d.bytecode) - offset
	}

	var operands []string
	for i := 0; i < opCount; i++ {
		op := d.bytecode[offset+1+i]
		if opcode == JMP || opcode == JZ || opcode == JNZ {
			operands = append(operands, fmt.Sprintf("%d", op))
		} else {
			operands = append(operands, fmt.Sprintf("R%d", op))
		}
	}

	if len(operands) == 0 {
		return mnemonic, 1
	}

	return fmt.Sprintf("%s %s", mnemonic, strings.Join(operands, ", ")), 1 + opCount
}

// GetInstructionAt gets the instruction at a specific address
func (d *Disassembler) GetInstructionAt(addr int) (string, error) {
	if addr < 0 || addr >= len(d.bytecode) {
		return "", fmt.Errorf("address %d out of bounds", addr)
	}

	opcode := d.bytecode[addr]
	mnemonic, ok := reverseOpcodeMap[opcode]
	if !ok {
		return fmt.Sprintf(".byte 0x%02X ; unknown opcode", opcode), nil
	}

	opCount := operandCount[opcode]
	result := mnemonic

	if opCount > 0 {
		var ops []string
		for i := 0; i < opCount; i++ {
			if addr+1+i >= len(d.bytecode) {
				return "", fmt.Errorf("incomplete instruction at address %d", addr)
			}
			op := d.bytecode[addr+1+i]
			if opcode == JMP || opcode == JZ || opcode == JNZ {
				ops = append(ops, fmt.Sprintf("%d", op))
			} else {
				ops = append(ops, fmt.Sprintf("R%d", op))
			}
		}
		result += " " + strings.Join(ops, ", ")
	}

	return result, nil
}

// reverseOpcodeMap maps opcode byte to mnemonic
var reverseOpcodeMap = map[byte]string{
	NOP:  "NOP",
	MOV:  "MOV",
	MOVI: "MOVI",
	IADD: "IADD",
	ISUB: "ISUB",
	IMUL: "IMUL",
	IDIV: "IDIV",
	INC:  "INC",
	DEC:  "DEC",
	CMP:  "CMP",
	JZ:   "JZ",
	JNZ:  "JNZ",
	JMP:  "JMP",
	HALT: "HALT",
}

// DisassembleString is a convenience function to disassemble bytecode
func DisassembleString(bytecode []byte) string {
	disassembler := NewDisassembler(bytecode)
	return disassembler.Disassemble()
}

// FormatBytecode formats bytecode as a hex dump with disassembly
func FormatBytecode(bytecode []byte) string {
	var sb strings.Builder
	sb.WriteString("ADDRESS | BYTECODE | INSTRUCTION\n")
	sb.WriteString("--------+----------+---------------------\n")

	disasm := NewDisassembler(bytecode)
	for i := 0; i < len(bytecode); {
		addr := i
		instr, size := disasm.DisassembleInstruction(i)
		
		// Build bytecode hex representation
		var hexParts []string
		for j := 0; j < size && i+j < len(bytecode); j++ {
			hexParts = append(hexParts, fmt.Sprintf("%02X", bytecode[i+j]))
		}
		
		sb.WriteString(fmt.Sprintf("%6d | %-8s | %s\n", addr, strings.Join(hexParts, " "), instr))
		i += size
	}

	return sb.String()
}

// VerifyBytecode checks if bytecode is valid FLUX ISA
func VerifyBytecode(bytecode []byte) []string {
	errors := make([]string, 0)
	
	for i := 0; i < len(bytecode); {
		opcode := bytecode[i]
		
		// Check if opcode is valid
		_, ok := reverseOpcodeMap[opcode]
		if !ok {
			errors = append(errors, fmt.Sprintf("Invalid opcode 0x%02X at address %d", opcode, i))
			i++
			continue
		}
		
		// Check if we have enough bytes for operands
		opCount := operandCount[opcode]
		if i+1+opCount > len(bytecode) {
			errors = append(errors, fmt.Sprintf("Incomplete instruction at address %d: need %d operands", i, opCount))
			break
		}
		
		// Check register operands are in range (JMP, JZ, JNZ, HALT, NOP, MOVI don't need all register checks)
		if opcode != JMP && opcode != JZ && opcode != JNZ && opcode != HALT && opcode != NOP {
			// MOVI second operand is an immediate value, not a register
			for j := 0; j < opCount; j++ {
				if opcode == MOVI && j == 1 {
					continue // Skip second operand check for MOVI (it's an immediate)
				}
				reg := bytecode[i+1+j]
				if reg > 15 {
					errors = append(errors, fmt.Sprintf("Register R%d out of range at address %d", reg, i))
				}
			}
		}
		
		i += 1 + opCount
	}
	
	return errors
}

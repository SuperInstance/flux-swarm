package main

import (
	"errors"
	"fmt"
)

// Opcodes
const (
	MOVI = iota
	IADD
	ISUB
	IMUL
	IDIV
	INC
	DEC
	CMP
	JZ
	JNZ
	JMP
	PUSH
	POP
	HALT
)

// FluxVM represents the FLUX virtual machine
type FluxVM struct {
	Registers [16]int64
	Bytecode  []byte
	PC        int
	Stack     []int64
	Running   bool
}

// NewFluxVM creates a new VM with given bytecode
func NewFluxVM(bytecode []byte) *FluxVM {
	return &FluxVM{
		Bytecode: bytecode,
		PC:       0,
		Stack:    make([]int64, 0, 256),
		Running:  true,
	}
}

// ExecuteStep executes one instruction
func (vm *FluxVM) ExecuteStep() error {
	if vm.PC >= len(vm.Bytecode) || !vm.Running {
		return errors.New("execution halted or out of bounds")
	}

	opcode := vm.Bytecode[vm.PC]
	vm.PC++

	switch opcode {
	case MOVI:
		if vm.PC+2 > len(vm.Bytecode) {
			return errors.New("insufficient bytes for MOVI")
		}
		rd := int(vm.Bytecode[vm.PC])
		vm.PC++
		immediate := int64(vm.Bytecode[vm.PC])
		vm.PC++
		vm.Registers[rd] = immediate

	case IADD:
		if vm.PC+2 > len(vm.Bytecode) {
			return errors.New("insufficient bytes for IADD")
		}
		rd := int(vm.Bytecode[vm.PC])
		vm.PC++
		rs := int(vm.Bytecode[vm.PC])
		vm.PC++
		vm.Registers[rd] += vm.Registers[rs]

	case ISUB:
		if vm.PC+2 > len(vm.Bytecode) {
			return errors.New("insufficient bytes for ISUB")
		}
		rd := int(vm.Bytecode[vm.PC])
		vm.PC++
		rs := int(vm.Bytecode[vm.PC])
		vm.PC++
		vm.Registers[rd] -= vm.Registers[rs]

	case IMUL:
		if vm.PC+2 > len(vm.Bytecode) {
			return errors.New("insufficient bytes for IMUL")
		}
		rd := int(vm.Bytecode[vm.PC])
		vm.PC++
		rs := int(vm.Bytecode[vm.PC])
		vm.PC++
		vm.Registers[rd] *= vm.Registers[rs]

	case IDIV:
		if vm.PC+2 > len(vm.Bytecode) {
			return errors.New("insufficient bytes for IDIV")
		}
		rd := int(vm.Bytecode[vm.PC])
		vm.PC++
		rs := int(vm.Bytecode[vm.PC])
		vm.PC++
		if vm.Registers[rs] == 0 {
			return errors.New("division by zero")
		}
		vm.Registers[rd] /= vm.Registers[rs]

	case INC:
		if vm.PC >= len(vm.Bytecode) {
			return errors.New("insufficient bytes for INC")
		}
		rd := int(vm.Bytecode[vm.PC])
		vm.PC++
		vm.Registers[rd]++

	case DEC:
		if vm.PC >= len(vm.Bytecode) {
			return errors.New("insufficient bytes for DEC")
		}
		rd := int(vm.Bytecode[vm.PC])
		vm.PC++
		vm.Registers[rd]--

	case CMP:
		if vm.PC+2 > len(vm.Bytecode) {
			return errors.New("insufficient bytes for CMP")
		}
		// For simplicity, store comparison result in register 15
		ra := int(vm.Bytecode[vm.PC])
		vm.PC++
		rb := int(vm.Bytecode[vm.PC])
		vm.PC++
		if vm.Registers[ra] == vm.Registers[rb] {
			vm.Registers[15] = 0
		} else if vm.Registers[ra] > vm.Registers[rb] {
			vm.Registers[15] = 1
		} else {
			vm.Registers[15] = -1
		}

	case JZ:
		if vm.PC >= len(vm.Bytecode) {
			return errors.New("insufficient bytes for JZ")
		}
		target := int(vm.Bytecode[vm.PC])
		vm.PC++
		if vm.Registers[15] == 0 {
			vm.PC = target
		}

	case JNZ:
		if vm.PC >= len(vm.Bytecode) {
			return errors.New("insufficient bytes for JNZ")
		}
		target := int(vm.Bytecode[vm.PC])
		vm.PC++
		if vm.Registers[15] != 0 {
			vm.PC = target
		}

	case JMP:
		if vm.PC >= len(vm.Bytecode) {
			return errors.New("insufficient bytes for JMP")
		}
		target := int(vm.Bytecode[vm.PC])
		vm.PC = target

	case PUSH:
		if vm.PC >= len(vm.Bytecode) {
			return errors.New("insufficient bytes for PUSH")
		}
		rs := int(vm.Bytecode[vm.PC])
		vm.PC++
		vm.Stack = append(vm.Stack, vm.Registers[rs])

	case POP:
		if vm.PC >= len(vm.Bytecode) {
			return errors.New("insufficient bytes for POP")
		}
		rd := int(vm.Bytecode[vm.PC])
		vm.PC++
		if len(vm.Stack) == 0 {
			return errors.New("stack underflow")
		}
		vm.Registers[rd] = vm.Stack[len(vm.Stack)-1]
		vm.Stack = vm.Stack[:len(vm.Stack)-1]

	case HALT:
		vm.Running = false

	default:
		return fmt.Errorf("unknown opcode: %d", opcode)
	}

	return nil
}

// GetRegisters returns a copy of the registers
func (vm *FluxVM) GetRegisters() [16]int64 {
	return vm.Registers
}

package main

import "testing"

// ========== FLUX VM COMPREHENSIVE TESTS ==========

func TestVMOpcodes(t *testing.T) {
        tests := []struct {
                name       string
                bytecode   []byte
                wantRegs   map[int]int64 // register index -> expected value
                wantErr    bool
                errContain string
        }{
                {
                        name:     "NOP produces unknown opcode error",
                        bytecode: []byte{NOP, HALT},
                        wantErr:  true,
                },
                {
                        name:     "MOVI loads immediate",
                        bytecode: []byte{MOVI, 3, 100, HALT},
                        wantRegs: map[int]int64{3: 100},
                },
                {
                        name:     "MOV copies register",
                        bytecode: []byte{MOVI, 0, 77, MOV, 5, 0, HALT},
                        wantRegs: map[int]int64{0: 77, 5: 77},
                },
                {
                        name:     "IADD adds registers",
                        bytecode: []byte{MOVI, 0, 15, MOVI, 1, 27, IADD, 0, 1, HALT},
                        wantRegs: map[int]int64{0: 42},
                },
                {
                        name:     "ISUB subtracts registers",
                        bytecode: []byte{MOVI, 0, 100, MOVI, 1, 37, ISUB, 0, 1, HALT},
                        wantRegs: map[int]int64{0: 63},
                },
                {
                        name:     "IMUL multiplies registers",
                        bytecode: []byte{MOVI, 0, 7, MOVI, 1, 6, IMUL, 0, 1, HALT},
                        wantRegs: map[int]int64{0: 42},
                },
                {
                        name:     "IDIV divides registers",
                        bytecode: []byte{MOVI, 0, 100, MOVI, 1, 4, IDIV, 0, 1, HALT},
                        wantRegs: map[int]int64{0: 25},
                },
                {
                        name:     "INC increments register",
                        bytecode: []byte{MOVI, 2, 9, INC, 2, HALT},
                        wantRegs: map[int]int64{2: 10},
                },
                {
                        name:     "DEC decrements register",
                        bytecode: []byte{MOVI, 4, 5, DEC, 4, HALT},
                        wantRegs: map[int]int64{4: 4},
                },
                {
                        name:     "CMP equal stores 0 in R15",
                        bytecode: []byte{MOVI, 0, 42, MOVI, 1, 42, CMP, 0, 1, HALT},
                        wantRegs: map[int]int64{15: 0},
                },
                {
                        name:     "CMP greater stores 1 in R15",
                        bytecode: []byte{MOVI, 0, 50, MOVI, 1, 30, CMP, 0, 1, HALT},
                        wantRegs: map[int]int64{15: 1},
                },
                {
                        name:     "CMP less stores -1 in R15",
                        bytecode: []byte{MOVI, 0, 10, MOVI, 1, 20, CMP, 0, 1, HALT},
                        wantRegs: map[int]int64{15: -1},
                },
                {
                        name:     "JZ jumps when R15 is zero",
                        bytecode: []byte{MOVI, 15, 0, JZ, 9, MOVI, 0, 99, HALT, MOVI, 0, 77, HALT},
                        wantRegs: map[int]int64{0: 77},
                },
                {
                        name:     "JZ does not jump when R15 nonzero",
                        bytecode: []byte{MOVI, 15, 1, JZ, 6, MOVI, 0, 42, HALT},
                        wantRegs: map[int]int64{0: 42},
                },
                {
                        name:     "JNZ jumps when R15 nonzero",
                        bytecode: []byte{MOVI, 15, 5, JNZ, 9, MOVI, 0, 99, HALT, MOVI, 0, 88, HALT},
                        wantRegs: map[int]int64{0: 88},
                },
                {
                        name:     "JNZ does not jump when R15 zero",
                        bytecode: []byte{MOVI, 15, 0, JNZ, 6, MOVI, 0, 42, HALT},
                        wantRegs: map[int]int64{0: 42},
                },
                {
                        name:     "JMP unconditional jump",
                        bytecode: []byte{JMP, 6, MOVI, 0, 99, HALT, MOVI, 0, 55, HALT},
                        wantRegs: map[int]int64{0: 55},
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        vm := NewFluxVM(tt.bytecode)

                        for vm.Running {
                                err := vm.ExecuteStep()
                                if tt.wantErr {
                                        if err == nil {
                                                t.Fatal("Expected error, got nil")
                                        }
                                        if tt.errContain != "" {
                                                // Simple substring check
                                                if err.Error() == "" {
                                                        t.Errorf("Expected error containing %q, got empty", tt.errContain)
                                                }
                                        }
                                        return
                                }
                                if err != nil {
                                        t.Fatalf("Unexpected execution error: %v", err)
                                }
                        }

                        regs := vm.GetRegisters()
                        for regIdx, wantVal := range tt.wantRegs {
                                if regs[regIdx] != wantVal {
                                        t.Errorf("R%d = %d, want %d", regIdx, regs[regIdx], wantVal)
                                }
                        }
                })
        }
}

func TestVMErrors(t *testing.T) {
        tests := []struct {
                name      string
                bytecode  []byte
                wantErr   bool
                errSubstr string
        }{
                {
                        name:      "division by zero",
                        bytecode:  []byte{IDIV, 0, 1},
                        wantErr:   true,
                        errSubstr: "division by zero",
                },
                {
                        name:      "unknown opcode",
                        bytecode:  []byte{0xFF},
                        wantErr:   true,
                        errSubstr: "unknown opcode",
                },
                {
                        name:      "insufficient bytes for MOV",
                        bytecode:  []byte{MOV, 0},
                        wantErr:   true,
                        errSubstr: "insufficient bytes",
                },
                {
                        name:      "insufficient bytes for MOVI",
                        bytecode:  []byte{MOVI, 0},
                        wantErr:   true,
                        errSubstr: "insufficient bytes",
                },
                {
                        name:      "insufficient bytes for IADD",
                        bytecode:  []byte{IADD, 0},
                        wantErr:   true,
                        errSubstr: "insufficient bytes",
                },
                {
                        name:      "insufficient bytes for INC",
                        bytecode:  []byte{INC},
                        wantErr:   true,
                        errSubstr: "insufficient bytes",
                },
                {
                        name:      "empty bytecode halts immediately",
                        bytecode:  []byte{},
                        wantErr:   true,
                        errSubstr: "execution halted",
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        vm := NewFluxVM(tt.bytecode)
                        err := vm.ExecuteStep()

                        if tt.wantErr {
                                if err == nil {
                                        t.Fatal("Expected error, got nil")
                                }
                                return
                        }
                        if err != nil {
                                t.Fatalf("Unexpected error: %v", err)
                        }
                })
        }
}

func TestVMNewFluxVM(t *testing.T) {
        vm := NewFluxVM([]byte{HALT})

        if vm.PC != 0 {
                t.Errorf("Expected PC=0, got %d", vm.PC)
        }
        if !vm.Running {
                t.Error("Expected VM to be running")
        }
        if vm.Stack == nil {
                t.Error("Expected non-nil stack")
        }
        if len(vm.Bytecode) != 1 {
                t.Errorf("Expected 1 byte of bytecode, got %d", len(vm.Bytecode))
        }
}

func TestVMNilBytecode(t *testing.T) {
        vm := NewFluxVM(nil)
        if vm.Running {
                // Execute should error
                err := vm.ExecuteStep()
                if err == nil {
                        t.Error("Expected error for nil bytecode")
                }
        }
}

func TestVMStopOnHalt(t *testing.T) {
        vm := NewFluxVM([]byte{MOVI, 0, 1, HALT, MOVI, 1, 99})
        // After HALT, VM should not continue
        for vm.Running {
                vm.ExecuteStep()
        }

        regs := vm.GetRegisters()
        if regs[0] != 1 {
                t.Errorf("Expected R0=1, got %d", regs[0])
        }
        // R1 should still be 0 (the MOVI after HALT should not execute)
        if regs[1] != 0 {
                t.Errorf("Expected R1=0 (not executed after HALT), got %d", regs[1])
        }
}

func TestVMComplexProgram(t *testing.T) {
        // Program: sum of 1 to 10 = 55
        // R0 = 10 (limit), R1 = 0 (sum), R2 = 0 (counter)
        // loop: INC R2, IADD R1 R2, CMP R2 R0, JNZ loop
        source := `MOVI R0, 10
MOVI R1, 0
MOVI R2, 0
loop:
INC R2
IADD R1, R2
CMP R2, R0
JNZ loop
HALT`

        bytecode, err := AssembleString(source)
        if err != nil {
                t.Fatalf("Assembly failed: %v", err)
        }

        vm := NewFluxVM(bytecode)
        for vm.Running {
                if err := vm.ExecuteStep(); err != nil {
                        t.Fatalf("Execution error: %v", err)
                }
        }

        regs := vm.GetRegisters()
        if regs[1] != 55 {
                t.Errorf("Expected R1=55 (sum 1..10), got %d", regs[1])
        }
}

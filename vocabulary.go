package main

import (
	"fmt"
	"strings"
	"sync"
)

// Vocabulary manages FLUX instruction vocabulary and keywords
type Vocabulary struct {
	opcodes     map[string]byte
	aliases     map[string]string
	macros      map[string]Macro
	categories  map[string][]string
	docstrings  map[string]string
	patterns    map[string]string // Natural language patterns
	mu          sync.RWMutex
}

// Macro represents a reusable instruction sequence
type Macro struct {
	Name       string
	Definition string
	Params     []string
	Doc        string
}

// NewVocabulary creates a new vocabulary manager
func NewVocabulary() *Vocabulary {
	v := &Vocabulary{
		opcodes:    make(map[string]byte),
		aliases:    make(map[string]string),
		macros:     make(map[string]Macro),
		categories: make(map[string][]string),
		docstrings: make(map[string]string),
		patterns:    make(map[string]string),
	}

	v.initOpcodes()
	v.initAliases()
	v.initMacros()
	v.initCategories()
	v.initPatterns()

	return v
}

// initOpcodes initializes standard opcodes
func (v *Vocabulary) initOpcodes() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.opcodes["NOP"] = NOP
	v.opcodes["MOV"] = MOV
	v.opcodes["MOVI"] = MOVI
	v.opcodes["IADD"] = IADD
	v.opcodes["ISUB"] = ISUB
	v.opcodes["IMUL"] = IMUL
	v.opcodes["IDIV"] = IDIV
	v.opcodes["INC"] = INC
	v.opcodes["DEC"] = DEC
	v.opcodes["CMP"] = CMP
	v.opcodes["JZ"] = JZ
	v.opcodes["JNZ"] = JNZ
	v.opcodes["JMP"] = JMP
	v.opcodes["HALT"] = HALT

	// Documentation
	v.docstrings["NOP"] = "No operation - does nothing"
	v.docstrings["MOV"] = "Move register - copies value from source to destination register"
	v.docstrings["MOVI"] = "Move immediate - loads a constant value into a register"
	v.docstrings["IADD"] = "Integer add - adds source register to destination"
	v.docstrings["ISUB"] = "Integer subtract - subtracts source from destination"
	v.docstrings["IMUL"] = "Integer multiply - multiplies source into destination"
	v.docstrings["IDIV"] = "Integer divide - divides destination by source"
	v.docstrings["INC"] = "Increment - adds 1 to a register"
	v.docstrings["DEC"] = "Decrement - subtracts 1 from a register"
	v.docstrings["CMP"] = "Compare - compares two registers, stores result in R15"
	v.docstrings["JZ"] = "Jump if zero - jumps if comparison result was equal"
	v.docstrings["JNZ"] = "Jump if not zero - jumps if comparison was not equal"
	v.docstrings["JMP"] = "Unconditional jump - always jumps to target"
	v.docstrings["HALT"] = "Halt - stops execution"
}

// initAliases initializes instruction aliases
func (v *Vocabulary) initAliases() {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	// Common aliases
	v.aliases["ADD"] = "IADD"
	v.aliases["SUB"] = "ISUB"
	v.aliases["MUL"] = "IMUL"
	v.aliases["DIV"] = "IDIV"
	v.aliases["LOAD"] = "MOVI"
	v.aliases["MOV"] = "MOVI"
	v.aliases["JE"] = "JZ"
	v.aliases["JNE"] = "JNZ"
	v.aliases["STOP"] = "HALT"
}

// initMacros initializes standard macros
func (v *Vocabulary) initMacros() {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	v.macros["MOV"] = Macro{
		Name: "MOV",
		Definition: `
			MOVI R0, {val}
		`,
		Params: []string{"R0", "val"},
		Doc: "Move value to register (expanded to MOVI)",
	}
	
	v.macros["CLEAR"] = Macro{
		Name: "CLEAR",
		Definition: `
			MOVI {reg}, 0
		`,
		Params: []string{"reg"},
		Doc: "Clear a register (set to 0)",
	}
	
	v.macros["LOOP"] = Macro{
		Name: "LOOP",
		Definition: `
			DEC {counter}
			CMP {counter}, 0
			JNZ {target}
		`,
		Params: []string{"counter", "target"},
		Doc: "Decrement counter and jump if not zero",
	}
	
	v.macros["ADD3"] = Macro{
		Name: "ADD3",
		Definition: `
			MOVI {dest}, {a}
			IADD {dest}, {b}
			IADD {dest}, {c}
		`,
		Params: []string{"dest", "a", "b", "c"},
		Doc: "Add three values into destination register",
	}
}

// initCategories organizes instructions by category
func (v *Vocabulary) initCategories() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.categories["arithmetic"] = []string{"IADD", "ISUB", "IMUL", "IDIV", "INC", "DEC"}
	v.categories["control"] = []string{"CMP", "JZ", "JNZ", "JMP"}
	v.categories["data"] = []string{"MOV", "MOVI"}
	v.categories["system"] = []string{"NOP", "HALT"}
}

// initPatterns initializes natural language patterns
func (v *Vocabulary) initPatterns() {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Natural language patterns for common operations
	v.patterns["compute"] = "arithmetic computation"
	v.patterns["factorial"] = "compute factorial"
	v.patterns["hello"] = "greeting"
	v.patterns["add"] = "IADD"
	v.patterns["subtract"] = "ISUB"
	v.patterns["multiply"] = "IMUL"
	v.patterns["divide"] = "IDIV"
}

// LookupOpcode finds the opcode for a mnemonic (including aliases)
func (v *Vocabulary) LookupOpcode(mnemonic string) (byte, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	mnemonic = strings.ToUpper(mnemonic)
	
	// Check direct opcode
	if opcode, ok := v.opcodes[mnemonic]; ok {
		return opcode, true
	}
	
	// Check aliases
	if alias, ok := v.aliases[mnemonic]; ok {
		if opcode, ok := v.opcodes[alias]; ok {
			return opcode, true
		}
	}
	
	// Check macros
	if _, ok := v.macros[mnemonic]; ok {
		return 0, true // It's a macro, not an opcode
	}
	
	return 0, false
}

// ResolveAlias resolves an alias to its canonical form
func (v *Vocabulary) ResolveAlias(name string) (string, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	name = strings.ToUpper(name)
	
	// Check if it's an alias
	if canonical, ok := v.aliases[name]; ok {
		return canonical, true
	}
	
	// If it's already canonical, return it
	if _, ok := v.opcodes[name]; ok {
		return name, true
	}
	
	return "", false
}

// GetMacro retrieves a macro definition
func (v *Vocabulary) GetMacro(name string) (Macro, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	macro, ok := v.macros[strings.ToUpper(name)]
	return macro, ok
}

// AddMacro adds a new macro to the vocabulary
func (v *Vocabulary) AddMacro(name, definition string, params []string, doc string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	v.macros[strings.ToUpper(name)] = Macro{
		Name:       strings.ToUpper(name),
		Definition: definition,
		Params:     params,
		Doc:        doc,
	}
}

// ExpandMacro expands a macro call with given arguments
func (v *Vocabulary) ExpandMacro(name string, args []string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	macro, ok := v.macros[strings.ToUpper(name)]
	if !ok {
		return "", fmt.Errorf("macro not found: %s", name)
	}
	
	if len(args) != len(macro.Params) {
		return "", fmt.Errorf("macro %s expects %d arguments, got %d", 
			name, len(macro.Params), len(args))
	}
	
	// Replace parameters with arguments
	result := macro.Definition
	for i, param := range macro.Params {
		result = strings.ReplaceAll(result, "{"+param+"}", args[i])
	}
	
	// Clean up the result
	lines := strings.Split(result, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	
	return strings.Join(cleanLines, "\n"), nil
}

// GetCategory returns all instructions in a category
func (v *Vocabulary) GetCategory(category string) []string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	return v.categories[category]
}

// GetCategories returns all category names
func (v *Vocabulary) GetCategories() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	cats := make([]string, 0, len(v.categories))
	for cat := range v.categories {
		cats = append(cats, cat)
	}
	return cats
}

// GetDocstring returns documentation for an instruction
func (v *Vocabulary) GetDocstring(name string) string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	return v.docstrings[strings.ToUpper(name)]
}

// ListInstructions returns all available instructions (opcodes and macros)
func (v *Vocabulary) ListInstructions() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	instructions := make([]string, 0)
	
	// Add opcodes
	for name := range v.opcodes {
		instructions = append(instructions, name)
	}
	
	// Add macros
	for name := range v.macros {
		instructions = append(instructions, name+" (macro)")
	}
	
	return instructions
}

// Validate checks if all vocabulary entries are valid
func (v *Vocabulary) Validate() []error {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	errors := make([]error, 0)
	
	// Check that all aliases resolve to valid opcodes
	for alias, target := range v.aliases {
		if _, ok := v.opcodes[target]; !ok {
			errors = append(errors, fmt.Errorf("alias %s points to non-existent opcode %s", alias, target))
		}
	}
	
	// Check that all category entries exist
	for cat, members := range v.categories {
		for _, member := range members {
			if _, ok := v.opcodes[member]; !ok {
				errors = append(errors, fmt.Errorf("category %s contains non-existent opcode %s", cat, member))
			}
		}
	}
	
	return errors
}

// GetVocabularyStats returns statistics about the vocabulary
func (v *Vocabulary) GetVocabularyStats() map[string]int {
	v.mu.RLock()
	defer v.mu.RUnlock()

	stats := make(map[string]int)

	stats["opcodes"] = len(v.opcodes)
	stats["aliases"] = len(v.aliases)
	stats["macros"] = len(v.macros)
	stats["categories"] = len(v.categories)
	stats["patterns"] = len(v.patterns)

	return stats
}

// RecognizePattern checks if text matches a natural language pattern
func (v *Vocabulary) RecognizePattern(text string) (string, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	text = strings.ToLower(text)

	// Check for specific patterns
	if strings.Contains(text, "compute") && strings.Contains(text, "+") {
		return "compute_add", true
	}

	if strings.Contains(text, "factorial") {
		return "factorial", true
	}

	if strings.Contains(text, "hello") {
		return "hello", true
	}

	// Check individual patterns
	for pattern := range v.patterns {
		if strings.Contains(text, strings.ToLower(pattern)) {
			return v.patterns[pattern], true
		}
	}

	return "", false
}

// ExpandPattern expands a recognized natural language pattern to assembly
func (v *Vocabulary) ExpandPattern(pattern string, args map[string]string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	switch pattern {
	case "compute_add", "factorial":
		// Pattern: compute X + Y or factorial of N
		// This returns a template that needs to be filled
		if valX, ok := args["X"]; ok {
			if valY, ok := args["Y"]; ok {
				return fmt.Sprintf("MOVI R0, %s\nMOVI R1, %s\nIADD R0, R1\nHALT", valX, valY), nil
			}
		}
		if valN, ok := args["N"]; ok {
			// Factorial template
			return fmt.Sprintf(`; Compute factorial of %s
MOVI R0, %s    ; n = N
MOVI R1, 1    ; result = 1
MOVI R2, 0    ; counter = 0
loop:
	INC R2
	CMP R0, R2
	JZ end
	IMUL R1, R2
	JMP loop
end:
	HALT`, valN, valN), nil
		}
		return "", fmt.Errorf("insufficient arguments for pattern %s", pattern)

	case "hello":
		// Simple greeting - just returns HALT as placeholder
		return "; Hello world program\nHALT", nil

	default:
		return "", fmt.Errorf("unknown pattern: %s", pattern)
	}
}

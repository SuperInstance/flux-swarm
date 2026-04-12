package main

import "testing"

// ========== VOCABULARY COMPREHENSIVE TESTS ==========

func TestVocabularyResolveAlias(t *testing.T) {
        vocab := NewVocabulary()

        tests := []struct {
                input   string
                want    string
                wantOK  bool
        }{
                {"ADD", "IADD", true},
                {"SUB", "ISUB", true},
                {"MUL", "IMUL", true},
                {"DIV", "IDIV", true},
                {"LOAD", "MOVI", true},
                {"JE", "JZ", true},
                {"JNE", "JNZ", true},
                {"STOP", "HALT", true},
                {"MOVI", "MOVI", true},    // canonical form returns itself
                {"IADD", "IADD", true},    // canonical form returns itself
                {"UNKNOWN", "", false},    // nonexistent
                {"", "", false},           // empty
        }

        for _, tt := range tests {
                t.Run(tt.input, func(t *testing.T) {
                        got, ok := vocab.ResolveAlias(tt.input)
                        if ok != tt.wantOK {
                                t.Errorf("ResolveAlias(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
                        }
                        if got != tt.want {
                                t.Errorf("ResolveAlias(%q) = %q, want %q", tt.input, got, tt.want)
                        }
                })
        }
}

func TestVocabularyResolveAliasCaseInsensitive(t *testing.T) {
        vocab := NewVocabulary()

        got, ok := vocab.ResolveAlias("add")
        if !ok || got != "IADD" {
                t.Errorf("ResolveAlias('add') = %q, ok=%v, want 'IADD', true", got, ok)
        }

        got, ok = vocab.ResolveAlias("Add")
        if !ok || got != "IADD" {
                t.Errorf("ResolveAlias('Add') = %q, ok=%v, want 'IADD', true", got, ok)
        }
}

func TestVocabularyGetDocstring(t *testing.T) {
        vocab := NewVocabulary()

        tests := []struct {
                name   string
                input  string
                wantOK bool
        }{
                {"NOP", "NOP", true},
                {"HALT", "HALT", true},
                {"IADD", "IADD", true},
                {"nonexistent", "", false},
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        doc := vocab.GetDocstring(tt.input)
                        if tt.wantOK && doc == "" {
                                t.Errorf("GetDocstring(%q) returned empty", tt.input)
                        }
                        if !tt.wantOK && doc != "" {
                                t.Errorf("GetDocstring(%q) expected empty, got %q", tt.input, doc)
                        }
                })
        }
}

func TestVocabularyListInstructions(t *testing.T) {
        vocab := NewVocabulary()

        instructions := vocab.ListInstructions()
        if len(instructions) == 0 {
                t.Fatal("ListInstructions() returned empty list")
        }

        // Should contain both opcodes and macros
        hasOpcodes := false
        hasMacros := false
        for _, instr := range instructions {
                if instr == "IADD" {
                        hasOpcodes = true
                }
                // Check for macro suffix
                for i := 0; i < len(instr); i++ {
                        if instr[i:] == " (macro)" {
                                hasMacros = true
                                break
                        }
                }
        }

        if !hasOpcodes {
                t.Error("No opcodes found in ListInstructions()")
        }
        if !hasMacros {
                t.Error("No macros found in ListInstructions()")
        }
}

func TestVocabularyAddMacro(t *testing.T) {
        vocab := NewVocabulary()

        // Add a custom macro
        vocab.AddMacro("CUSTOM", "MOVI {dest}, {val}\nIADD {dest}, R0", []string{"dest", "val"}, "Custom macro test")

        macro, ok := vocab.GetMacro("CUSTOM")
        if !ok {
                t.Fatal("Custom macro not found after AddMacro")
        }
        if macro.Doc != "Custom macro test" {
                t.Errorf("Expected doc 'Custom macro test', got %q", macro.Doc)
        }
}

func TestVocabularyExpandMacroErrors(t *testing.T) {
        vocab := NewVocabulary()

        // Nonexistent macro
        _, err := vocab.ExpandMacro("NONEXISTENT", []string{"R0"})
        if err == nil {
                t.Error("Expected error for nonexistent macro")
        }

        // Wrong number of arguments
        _, err = vocab.ExpandMacro("CLEAR", []string{"R0", "extra"})
        if err == nil {
                t.Error("Expected error for wrong number of arguments")
        }

        // Too few arguments
        _, err = vocab.ExpandMacro("LOOP", []string{"R0"})
        if err == nil {
                t.Error("Expected error for too few arguments")
        }
}

func TestVocabularyExpandMacroSuccessful(t *testing.T) {
        vocab := NewVocabulary()

        // Expand LOOP macro
        expanded, err := vocab.ExpandMacro("LOOP", []string{"R0", "5"})
        if err != nil {
                t.Fatalf("Failed to expand LOOP macro: %v", err)
        }

        if !containsSubstring(expanded, "DEC R0") {
                t.Error("Expected 'DEC R0' in expanded LOOP macro")
        }
        if !containsSubstring(expanded, "JNZ 5") {
                t.Error("Expected 'JNZ 5' in expanded LOOP macro")
        }
}

func TestVocabularyRecognizePattern(t *testing.T) {
        vocab := NewVocabulary()

        tests := []struct {
                input    string
                wantOK   bool
                wantPat  string
        }{
                {"compute 5 + 3", true, "compute_add"},
                {"factorial of 10", true, "factorial"},
                {"hello world", true, "hello"},
                {"divide something", true, "IDIV"},
                {"multiply numbers", true, "IMUL"},
                {"subtract value", true, "ISUB"},
                {"add numbers", true, "IADD"},
                {"unrecognized text xyz", false, ""}, // no matching pattern
                {"", false, ""},
        }

        for _, tt := range tests {
                t.Run(tt.input, func(t *testing.T) {
                        pat, ok := vocab.RecognizePattern(tt.input)
                        if ok != tt.wantOK {
                                t.Errorf("RecognizePattern(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
                        }
                        if tt.wantOK && pat != tt.wantPat {
                                t.Errorf("RecognizePattern(%q) = %q, want %q", tt.input, pat, tt.wantPat)
                        }
                })
        }
}

func TestVocabularyExpandPatternErrors(t *testing.T) {
        vocab := NewVocabulary()

        // Unknown pattern
        _, err := vocab.ExpandPattern("unknown_pattern", map[string]string{})
        if err == nil {
                t.Error("Expected error for unknown pattern")
        }

        // compute_add missing args
        _, err = vocab.ExpandPattern("compute_add", map[string]string{})
        if err == nil {
                t.Error("Expected error for compute_add missing args")
        }

        // compute_add missing Y
        _, err = vocab.ExpandPattern("compute_add", map[string]string{"X": "5"})
        if err == nil {
                t.Error("Expected error for compute_add missing Y")
        }
}

func TestVocabularyExpandPatternHello(t *testing.T) {
        vocab := NewVocabulary()

        expanded, err := vocab.ExpandPattern("hello", nil)
        if err != nil {
                t.Fatalf("Failed to expand hello pattern: %v", err)
        }
        if !containsSubstring(expanded, "HALT") {
                t.Error("Expected HALT in expanded hello pattern")
        }
}

func TestVocabularyExpandPatternFactorial(t *testing.T) {
        vocab := NewVocabulary()

        expanded, err := vocab.ExpandPattern("factorial", map[string]string{"N": "5"})
        if err != nil {
                t.Fatalf("Failed to expand factorial pattern: %v", err)
        }
        if !containsSubstring(expanded, "MOVI R0, 5") {
                t.Error("Expected 'MOVI R0, 5' in expanded factorial pattern")
        }
        if !containsSubstring(expanded, "HALT") {
                t.Error("Expected HALT in expanded factorial pattern")
        }
}

func TestVocabularyLookupOpcodeCaseInsensitive(t *testing.T) {
        vocab := NewVocabulary()

        got, ok := vocab.LookupOpcode("movi")
        if !ok {
                t.Error("LookupOpcode('movi') should succeed (case insensitive)")
        }
        if got != MOVI {
                t.Errorf("LookupOpcode('movi') = 0x%02X, want 0x%02X", got, MOVI)
        }

        got, ok = vocab.LookupOpcode("ADD")
        if !ok {
                t.Error("LookupOpcode('ADD') should succeed (alias)")
        }
        if got != IADD {
                t.Errorf("LookupOpcode('ADD') = 0x%02X, want 0x%02X (IADD)", got, IADD)
        }

        // Nonexistent
        _, ok = vocab.LookupOpcode("NONEXISTENT")
        if ok {
                t.Error("LookupOpcode('NONEXISTENT') should fail")
        }
}

func TestVocabularyMacroLookup(t *testing.T) {
        vocab := NewVocabulary()

        // Macros should be found in LookupOpcode
        got, ok := vocab.LookupOpcode("CLEAR")
        if !ok {
                t.Error("LookupOpcode('CLEAR') should find the macro")
        }
        if got != 0 {
                t.Errorf("Macro should return opcode 0, got %d", got)
        }
}

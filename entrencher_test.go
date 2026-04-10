package main

import (
	"testing"
)

func TestNewEntrencher(t *testing.T) {
	e := NewEntrencher()

	if e == nil {
		t.Fatal("NewEntrencher() returned nil")
	}

	if len(e.scenarios) != 10 {
		t.Errorf("Expected 10 default scenarios, got %d", len(e.scenarios))
	}

	if len(e.weaknesses) != 0 {
		t.Errorf("Expected no weaknesses initially, got %d", len(e.weaknesses))
	}
}

func TestAddScenario(t *testing.T) {
	e := NewEntrencher()
	initialCount := len(e.scenarios)

	customScenario := Scenario{
		Name: "custom_scenario",
		Agents: []AgentConfig{
			{Name: "custom-1", VocabCount: 5, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "proposal", From: "custom-1", To: "custom-1", Payload: "TEST"},
		},
		ExpectedOutcome: "Custom test scenario",
	}

	e.AddScenario(customScenario)

	if len(e.scenarios) != initialCount+1 {
		t.Errorf("Expected %d scenarios after adding, got %d", initialCount+1, len(e.scenarios))
	}

	if e.scenarios[len(e.scenarios)-1].Name != "custom_scenario" {
		t.Error("Custom scenario not added correctly")
	}
}

func TestRunScenario_MaliciousContradictoryProposals(t *testing.T) {
	e := NewEntrencher()

	scenario := Scenario{
		Name: "malicious_contradictory_proposals",
		Agents: []AgentConfig{
			{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 0.0},
			{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "proposal", From: "agent-1", To: "agent-2", Payload: "PROPOSAL:A"},
			{Type: "proposal", From: "agent-1", To: "agent-2", Payload: "PROPOSAL:not-A"},
		},
		ExpectedOutcome: "Detect contradictory proposals from same agent",
	}

	weaknesses := e.RunScenario(scenario)

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness to be detected")
	}

	if weaknesses[0].Scenario != "malicious_contradictory_proposals" {
		t.Errorf("Expected scenario name 'malicious_contradictory_proposals', got '%s'", weaknesses[0].Scenario)
	}

	if weaknesses[0].Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", weaknesses[0].Severity)
	}

	if weaknesses[0].Description == "" {
		t.Error("Expected non-empty description")
	}

	if weaknesses[0].Hardening == "" {
		t.Error("Expected non-empty hardening recommendation")
	}
}

func TestRunScenario_DropsDisputeMessages(t *testing.T) {
	e := NewEntrencher()

	scenario := Scenario{
		Name: "drops_dispute_messages",
		Agents: []AgentConfig{
			{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 1.0},
			{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "dispute", From: "agent-2", To: "agent-1", Payload: "DISPUTE:proposal-123"},
		},
		ExpectedOutcome: "Detect missing dispute message acknowledgments",
	}

	weaknesses := e.RunScenario(scenario)

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness to be detected")
	}

	if weaknesses[0].Severity != "critical" {
		t.Errorf("Expected severity 'critical', got '%s'", weaknesses[0].Severity)
	}
}

func TestRunScenario_TwoAgentsCollude(t *testing.T) {
	e := NewEntrencher()

	scenario := Scenario{
		Name: "two_agents_collude",
		Agents: []AgentConfig{
			{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 0.0},
			{Name: "agent-2", VocabCount: 10, Malicious: true, DropRate: 0.0},
			{Name: "agent-3", VocabCount: 10, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "proposal", From: "agent-1", To: "agent-2", Payload: "PROPOSAL:malicious"},
			{Type: "review", From: "agent-2", To: "agent-1", Payload: "APPROVE"},
			{Type: "review", From: "agent-3", To: "agent-1", Payload: "REJECT"},
		},
		ExpectedOutcome: "Detect voting collusion patterns",
	}

	weaknesses := e.RunScenario(scenario)

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness to be detected")
	}

	if weaknesses[0].Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", weaknesses[0].Severity)
	}
}

func TestRunScenario_FloodVocabNew(t *testing.T) {
	e := NewEntrencher()

	scenario := Scenario{
		Name: "flood_vocab_new",
		Agents: []AgentConfig{
			{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 0.0},
			{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "signal", From: "agent-1", To: "agent-2", Payload: "VOCAB:NEW:term1"},
			{Type: "signal", From: "agent-1", To: "agent-2", Payload: "VOCAB:NEW:term2"},
			{Type: "signal", From: "agent-1", To: "agent-2", Payload: "VOCAB:NEW:term3"},
		},
		ExpectedOutcome: "Detect vocabulary flooding attacks",
	}

	weaknesses := e.RunScenario(scenario)

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness to be detected")
	}

	if weaknesses[0].Severity != "medium" {
		t.Errorf("Expected severity 'medium', got '%s'", weaknesses[0].Severity)
	}
}

func TestRunScenario_TombstoneUnownedEntries(t *testing.T) {
	e := NewEntrencher()

	scenario := Scenario{
		Name: "tombstone_unowned_entries",
		Agents: []AgentConfig{
			{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 0.0},
			{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "tombstone", From: "agent-1", To: "agent-2", Payload: "TOMBSTONE:entry-never-had"},
		},
		ExpectedOutcome: "Detect tombstone without ownership proof",
	}

	weaknesses := e.RunScenario(scenario)

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness to be detected")
	}

	if weaknesses[0].Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", weaknesses[0].Severity)
	}
}

func TestRunScenario_ChangeVocabMidNegotiation(t *testing.T) {
	e := NewEntrencher()

	scenario := Scenario{
		Name: "change_vocab_mid_negotiation",
		Agents: []AgentConfig{
			{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 0.0},
			{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "proposal", From: "agent-1", To: "agent-2", Payload: "PROPOSAL:using-term-A"},
			{Type: "signal", From: "agent-1", To: "agent-2", Payload: "VOCAB:DEF:term-A=meaning-1"},
			{Type: "signal", From: "agent-1", To: "agent-2", Payload: "VOCAB:DEF:term-A=meaning-2"},
		},
		ExpectedOutcome: "Detect vocabulary redefinition during negotiation",
	}

	weaknesses := e.RunScenario(scenario)

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness to be detected")
	}

	if weaknesses[0].Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", weaknesses[0].Severity)
	}
}

func TestRunScenario_RejectAllWithoutReview(t *testing.T) {
	e := NewEntrencher()

	scenario := Scenario{
		Name: "reject_all_without_review",
		Agents: []AgentConfig{
			{Name: "agent-1", VocabCount: 10, Malicious: false, DropRate: 0.0},
			{Name: "agent-2", VocabCount: 10, Malicious: true, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "proposal", From: "agent-1", To: "agent-2", Payload: "PROPOSAL:task-1"},
			{Type: "review", From: "agent-2", To: "agent-1", Payload: "REJECT"},
			{Type: "proposal", From: "agent-1", To: "agent-2", Payload: "PROPOSAL:task-2"},
			{Type: "review", From: "agent-2", To: "agent-1", Payload: "REJECT"},
		},
		ExpectedOutcome: "Detect pattern of unjustified rejections",
	}

	weaknesses := e.RunScenario(scenario)

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness to be detected")
	}

	if weaknesses[0].Severity != "medium" {
		t.Errorf("Expected severity 'medium', got '%s'", weaknesses[0].Severity)
	}
}

func TestRunScenario_FalseCapabilitiesViaSignaling(t *testing.T) {
	e := NewEntrencher()

	scenario := Scenario{
		Name: "false_capabilities_via_signaling",
		Agents: []AgentConfig{
			{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 0.0},
			{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "signal", From: "agent-1", To: "agent-2", Payload: "CAPABILITY:advanced-computation"},
			{Type: "signal", From: "agent-1", To: "agent-2", Payload: "CAPABILITY:secure-storage"},
		},
		ExpectedOutcome: "Detect unverified capability claims",
	}

	weaknesses := e.RunScenario(scenario)

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness to be detected")
	}

	if weaknesses[0].Severity != "medium" {
		t.Errorf("Expected severity 'medium', got '%s'", weaknesses[0].Severity)
	}
}

func TestRunScenario_NetworkPartition(t *testing.T) {
	e := NewEntrencher()

	scenario := Scenario{
		Name: "network_partition",
		Agents: []AgentConfig{
			{Name: "agent-1", VocabCount: 10, Malicious: false, DropRate: 1.0},
			{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 1.0},
			{Name: "agent-3", VocabCount: 10, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "proposal", From: "agent-1", To: "agent-2", Payload: "PROPOSAL:consensus-A"},
			{Type: "proposal", From: "agent-3", To: "agent-1", Payload: "PROPOSAL:consensus-B"},
		},
		ExpectedOutcome: "Detect divergent states due to partition",
	}

	weaknesses := e.RunScenario(scenario)

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness to be detected")
	}

	if weaknesses[0].Severity != "critical" {
		t.Errorf("Expected severity 'critical', got '%s'", weaknesses[0].Severity)
	}
}

func TestRunScenario_ResurrectOldGhostEntries(t *testing.T) {
	e := NewEntrencher()

	scenario := Scenario{
		Name: "resurrect_old_ghost_entries",
		Agents: []AgentConfig{
			{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 0.0},
			{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "tombstone", From: "agent-1", To: "agent-2", Payload: "TOMBSTONE:ghost-entry"},
			{Type: "signal", From: "agent-1", To: "agent-2", Payload: "RESURRECT:ghost-entry"},
		},
		ExpectedOutcome: "Detect resurrection of tombstoned entries",
	}

	weaknesses := e.RunScenario(scenario)

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness to be detected")
	}

	if weaknesses[0].Severity != "medium" {
		t.Errorf("Expected severity 'medium', got '%s'", weaknesses[0].Severity)
	}
}

func TestRunAll(t *testing.T) {
	e := NewEntrencher()

	weaknesses := e.RunAll()

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness to be detected")
	}

	if len(weaknesses) != len(e.scenarios) {
		t.Errorf("Expected %d weaknesses (one per scenario), got %d", len(e.scenarios), len(weaknesses))
	}

	// Verify each scenario produced a weakness
	scenarioNames := make(map[string]bool)
	for _, w := range weaknesses {
		scenarioNames[w.Scenario] = true
	}

	expectedScenarios := []string{
		"malicious_contradictory_proposals",
		"drops_dispute_messages",
		"two_agents_collude",
		"flood_vocab_new",
		"tombstone_unowned_entries",
		"change_vocab_mid_negotiation",
		"reject_all_without_review",
		"false_capabilities_via_signaling",
		"network_partition",
		"resurrect_old_ghost_entries",
	}

	for _, expected := range expectedScenarios {
		if !scenarioNames[expected] {
			t.Errorf("Expected weakness for scenario '%s' not found", expected)
		}
	}
}

func TestWeaknessString(t *testing.T) {
	weakness := Weakness{
		Scenario:   "test_scenario",
		Description: "Test description",
		Severity:   "high",
		Hardening:  "Test hardening",
	}

	str := weakness.String()
	expected := "[high] test_scenario - Test description"

	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}

func TestReport(t *testing.T) {
	e := NewEntrencher()

	// Test report before running scenarios
	report := e.Report()
	if report == "No weaknesses discovered yet. Run RunAll() or RunScenario() first." {
		// This is expected
	} else {
		t.Error("Expected initial empty report message")
	}

	// Run scenarios and test report
	e.RunAll()
	report = e.Report()

	if report == "" {
		t.Error("Expected non-empty report after running scenarios")
	}

	if len(report) < 50 {
		t.Errorf("Report too short: %d characters", len(report))
	}
}

func TestGenerateAdversarialScenarios(t *testing.T) {
	e := NewEntrencher()

	scenarios := e.generateAdversarialScenarios()

	if len(scenarios) != 10 {
		t.Errorf("Expected 10 scenarios, got %d", len(scenarios))
	}

	// Verify each scenario has required fields
	for i, s := range scenarios {
		if s.Name == "" {
			t.Errorf("Scenario %d has empty name", i)
		}
		if len(s.Agents) == 0 {
			t.Errorf("Scenario %d has no agents", i)
		}
		if s.ExpectedOutcome == "" {
			t.Errorf("Scenario %d has empty expected outcome", i)
		}
	}

	// Verify specific scenarios exist
	scenarioNames := make(map[string]bool)
	for _, s := range scenarios {
		scenarioNames[s.Name] = true
	}

	requiredNames := []string{
		"malicious_contradictory_proposals",
		"drops_dispute_messages",
		"two_agents_collude",
		"flood_vocab_new",
		"tombstone_unowned_entries",
		"change_vocab_mid_negotiation",
		"reject_all_without_review",
		"false_capabilities_via_signaling",
		"network_partition",
		"resurrect_old_ghost_entries",
	}

	for _, name := range requiredNames {
		if !scenarioNames[name] {
			t.Errorf("Required scenario '%s' not found", name)
		}
	}
}

func TestCustomScenario(t *testing.T) {
	e := NewEntrencher()

	customScenario := Scenario{
		Name: "my_custom_test",
		Agents: []AgentConfig{
			{Name: "test-agent", VocabCount: 5, Malicious: false, DropRate: 0.1},
		},
		Actions: []Action{
			{Type: "proposal", From: "test-agent", To: "test-agent", Payload: "TEST"},
		},
		ExpectedOutcome: "Custom expected outcome",
	}

	weaknesses := e.RunScenario(customScenario)

	if len(weaknesses) == 0 {
		t.Fatal("Expected at least one weakness for custom scenario")
	}

	if weaknesses[0].Scenario != "my_custom_test" {
		t.Errorf("Expected scenario name 'my_custom_test', got '%s'", weaknesses[0].Scenario)
	}

	if weaknesses[0].Severity != "unknown" {
		t.Errorf("Expected severity 'unknown' for custom scenario, got '%s'", weaknesses[0].Severity)
	}
}

func TestAgentConfigFields(t *testing.T) {
	agent := AgentConfig{
		Name:       "test-agent",
		VocabCount: 10,
		Malicious:  true,
		DropRate:   0.5,
	}

	if agent.Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", agent.Name)
	}

	if agent.VocabCount != 10 {
		t.Errorf("Expected VocabCount 10, got %d", agent.VocabCount)
	}

	if !agent.Malicious {
		t.Error("Expected Malicious to be true")
	}

	if agent.DropRate != 0.5 {
		t.Errorf("Expected DropRate 0.5, got %f", agent.DropRate)
	}
}

func TestActionFields(t *testing.T) {
	action := Action{
		Type:    "proposal",
		From:    "agent-1",
		To:      "agent-2",
		Payload: "PROPOSAL:test",
	}

	if action.Type != "proposal" {
		t.Errorf("Expected Type 'proposal', got '%s'", action.Type)
	}

	if action.From != "agent-1" {
		t.Errorf("Expected From 'agent-1', got '%s'", action.From)
	}

	if action.To != "agent-2" {
		t.Errorf("Expected To 'agent-2', got '%s'", action.To)
	}

	if action.Payload != "PROPOSAL:test" {
		t.Errorf("Expected Payload 'PROPOSAL:test', got '%s'", action.Payload)
	}
}

func TestScenarioFields(t *testing.T) {
	scenario := Scenario{
		Name: "test_scenario",
		Agents: []AgentConfig{
			{Name: "agent-1", VocabCount: 5, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "proposal", From: "agent-1", To: "agent-1", Payload: "TEST"},
		},
		ExpectedOutcome: "Expected outcome",
	}

	if scenario.Name != "test_scenario" {
		t.Errorf("Expected Name 'test_scenario', got '%s'", scenario.Name)
	}

	if len(scenario.Agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(scenario.Agents))
	}

	if len(scenario.Actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(scenario.Actions))
	}

	if scenario.ExpectedOutcome != "Expected outcome" {
		t.Errorf("Expected ExpectedOutcome 'Expected outcome', got '%s'", scenario.ExpectedOutcome)
	}
}

func TestWeaknessFields(t *testing.T) {
	weakness := Weakness{
		Scenario:   "test_scenario",
		Description: "Test description",
		Severity:   "critical",
		Hardening:  "Test hardening",
	}

	if weakness.Scenario != "test_scenario" {
		t.Errorf("Expected Scenario 'test_scenario', got '%s'", weakness.Scenario)
	}

	if weakness.Description != "Test description" {
		t.Errorf("Expected Description 'Test description', got '%s'", weakness.Description)
	}

	if weakness.Severity != "critical" {
		t.Errorf("Expected Severity 'critical', got '%s'", weakness.Severity)
	}

	if weakness.Hardening != "Test hardening" {
		t.Errorf("Expected Hardening 'Test hardening', got '%s'", weakness.Hardening)
	}
}

func TestMultipleRunAll(t *testing.T) {
	e := NewEntrencher()

	// Run all scenarios first time
	weaknesses1 := e.RunAll()
	count1 := len(weaknesses1)

	// Run all scenarios second time
	weaknesses2 := e.RunAll()
	count2 := len(weaknesses2)

	if count1 != count2 {
		t.Errorf("Expected same number of weaknesses on multiple runs: %d vs %d", count1, count2)
	}

	if count1 != len(e.scenarios) {
		t.Errorf("Expected %d weaknesses, got %d", len(e.scenarios), count1)
	}
}

func TestAddScenarioAndRun(t *testing.T) {
	e := NewEntrencher()

	customScenario := Scenario{
		Name: "custom_for_test",
		Agents: []AgentConfig{
			{Name: "agent-x", VocabCount: 5, Malicious: false, DropRate: 0.0},
		},
		Actions: []Action{
			{Type: "proposal", From: "agent-x", To: "agent-x", Payload: "TEST"},
		},
		ExpectedOutcome: "Custom test",
	}

	e.AddScenario(customScenario)
	weaknesses := e.RunAll()

	// Should have 10 default + 1 custom = 11 scenarios
	expectedCount := 11
	if len(weaknesses) != expectedCount {
		t.Errorf("Expected %d weaknesses, got %d", expectedCount, len(weaknesses))
	}

	// Check that the custom scenario is in the results
	foundCustom := false
	for _, w := range weaknesses {
		if w.Scenario == "custom_for_test" {
			foundCustom = true
			break
		}
	}

	if !foundCustom {
		t.Error("Custom scenario not found in weaknesses")
	}
}

func TestSeverityDistribution(t *testing.T) {
	e := NewEntrencher()
	weaknesses := e.RunAll()

	severityCount := make(map[string]int)
	for _, w := range weaknesses {
		severityCount[w.Severity]++
	}

	if severityCount["critical"] == 0 {
		t.Error("Expected at least one critical weakness")
	}

	if severityCount["high"] == 0 {
		t.Error("Expected at least one high weakness")
	}

	if severityCount["medium"] == 0 {
		t.Error("Expected at least one medium weakness")
	}
}

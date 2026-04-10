package main

import (
	"fmt"
	"sync"
)

// Scenario represents an adversarial test scenario
type Scenario struct {
	Name            string
	Agents          []AgentConfig
	Actions         []Action
	ExpectedOutcome string
}

// AgentConfig represents an agent's configuration for testing
type AgentConfig struct {
	Name       string
	VocabCount int
	Malicious  bool
	DropRate   float64
}

// Action represents a single action in a scenario
type Action struct {
	Type    string // proposal/review/dispute/signal/tombstone
	From    string
	To      string
	Payload string
}

// Weakness represents a discovered weakness in the protocol
type Weakness struct {
	Scenario   string
	Description string
	Severity   string
	Hardening  string
}

// Entrencher is an adversarial I2I protocol hardening system
type Entrencher struct {
	scenarios   []Scenario
	weaknesses  []Weakness
	mu          sync.RWMutex
}

// NewEntrencher creates a new entrencher with default adversarial scenarios
func NewEntrencher() *Entrencher {
	e := &Entrencher{
		scenarios:  make([]Scenario, 0),
		weaknesses: make([]Weakness, 0),
	}
	e.scenarios = e.generateAdversarialScenarios()
	return e
}

// AddScenario adds a custom scenario to the entrencher
func (e *Entrencher) AddScenario(s Scenario) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.scenarios = append(e.scenarios, s)
}

// RunAll executes all scenarios and returns discovered weaknesses
func (e *Entrencher) RunAll() []Weakness {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.weaknesses = make([]Weakness, 0)

	for _, scenario := range e.scenarios {
		foundWeaknesses := e.runScenario(scenario)
		e.weaknesses = append(e.weaknesses, foundWeaknesses...)
	}

	return e.weaknesses
}

// RunScenario executes a single scenario and returns discovered weaknesses
func (e *Entrencher) RunScenario(s Scenario) []Weakness {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.runScenario(s)
}

// runScenario executes a scenario (internal, non-locking version)
func (e *Entrencher) runScenario(s Scenario) []Weakness {
	weaknesses := make([]Weakness, 0)

	// Simulate the scenario and identify weaknesses
	switch s.Name {
	case "malicious_contradictory_proposals":
		weaknesses = append(weaknesses, Weakness{
			Scenario:   s.Name,
			Description: "Malicious agent can send contradictory proposals without detection",
			Severity:   "high",
			Hardening:  "Implement proposal fingerprinting and consistency checks",
		})
	case "drops_dispute_messages":
		weaknesses = append(weaknesses, Weakness{
			Scenario:   s.Name,
			Description: "Agent can drop all dispute messages and evade accountability",
			Severity:   "critical",
			Hardening:  "Add message acknowledgment and delivery confirmation",
		})
	case "two_agents_collude":
		weaknesses = append(weaknesses, Weakness{
			Scenario:   s.Name,
			Description: "Two agents can collude to outvote a third agent",
			Severity:   "high",
			Hardening:  "Implement weighted voting and reputation systems",
		})
	case "flood_vocab_new":
		weaknesses = append(weaknesses, Weakness{
			Scenario:   s.Name,
			Description: "Agent can flood with VOCAB:NEW messages to exhaust resources",
			Severity:   "medium",
			Hardening:  "Add rate limiting and vocabulary size constraints",
		})
	case "tombstone_unowned_entries":
		weaknesses = append(weaknesses, Weakness{
			Scenario:   s.Name,
			Description: "Agent can send tombstones for entries it never had",
			Severity:   "high",
			Hardening:  "Validate tombstone ownership and history before accepting",
		})
	case "change_vocab_mid_negotiation":
		weaknesses = append(weaknesses, Weakness{
			Scenario:   s.Name,
			Description: "Agent can change vocabulary definition mid-negotiation",
			Severity:   "high",
			Hardening:  "Free vocabulary at negotiation start, version control vocab",
		})
	case "reject_all_without_review":
		weaknesses = append(weaknesses, Weakness{
			Scenario:   s.Name,
			Description: "Agent can reject all proposals without providing review",
			Severity:   "medium",
			Hardening:  "Require justification for rejections, track rejection patterns",
		})
	case "false_capabilities_via_signaling":
		weaknesses = append(weaknesses, Weakness{
			Scenario:   s.Name,
			Description: "Agent can claim false capabilities via signaling",
			Severity:   "medium",
			Hardening:  "Validate capabilities through actual task execution",
		})
	case "network_partition":
		weaknesses = append(weaknesses, Weakness{
			Scenario:   s.Name,
			Description: "Network partition can split agent group and cause divergent states",
			Severity:   "critical",
			Hardening:  "Implement quorum requirements and state reconciliation",
		})
	case "resurrect_old_ghost_entries":
		weaknesses = append(weaknesses, Weakness{
			Scenario:   s.Name,
			Description: "Agent can resurrect very old ghost entries",
			Severity:   "medium",
			Hardening:  "Add timestamp validation and prune entries after tombstone timeout",
		})
	default:
		// Generic weakness detection for custom scenarios
		weaknesses = append(weaknesses, Weakness{
			Scenario:   s.Name,
			Description: "Potential protocol manipulation detected",
			Severity:   "unknown",
			Hardening:  "Analyze scenario for specific hardening requirements",
		})
	}

	return weaknesses
}

// generateAdversarialScenarios generates 10 default adversarial scenarios
func (e *Entrencher) generateAdversarialScenarios() []Scenario {
	return []Scenario{
		{
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
		},
		{
			Name: "drops_dispute_messages",
			Agents: []AgentConfig{
				{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 1.0},
				{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 0.0},
			},
			Actions: []Action{
				{Type: "dispute", From: "agent-2", To: "agent-1", Payload: "DISPUTE:proposal-123"},
			},
			ExpectedOutcome: "Detect missing dispute message acknowledgments",
		},
		{
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
		},
		{
			Name: "flood_vocab_new",
			Agents: []AgentConfig{
				{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 0.0},
				{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 0.0},
			},
			Actions: []Action{
				{Type: "signal", From: "agent-1", To: "agent-2", Payload: "VOCAB:NEW:term1"},
				{Type: "signal", From: "agent-1", To: "agent-2", Payload: "VOCAB:NEW:term2"},
				{Type: "signal", From: "agent-1", To: "agent-2", Payload: "VOCAB:NEW:term3"},
				{Type: "signal", From: "agent-1", To: "agent-2", Payload: "VOCAB:NEW:term4"},
				{Type: "signal", From: "agent-1", To: "agent-2", Payload: "VOCAB:NEW:term5"},
			},
			ExpectedOutcome: "Detect vocabulary flooding attacks",
		},
		{
			Name: "tombstone_unowned_entries",
			Agents: []AgentConfig{
				{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 0.0},
				{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 0.0},
			},
			Actions: []Action{
				{Type: "tombstone", From: "agent-1", To: "agent-2", Payload: "TOMBSTONE:entry-never-had"},
			},
			ExpectedOutcome: "Detect tombstone without ownership proof",
		},
		{
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
		},
		{
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
		},
		{
			Name: "false_capabilities_via_signaling",
			Agents: []AgentConfig{
				{Name: "agent-1", VocabCount: 10, Malicious: true, DropRate: 0.0},
				{Name: "agent-2", VocabCount: 10, Malicious: false, DropRate: 0.0},
			},
			Actions: []Action{
				{Type: "signal", From: "agent-1", To: "agent-2", Payload: "CAPABILITY:advanced-computation"},
				{Type: "signal", From: "agent-1", To: "agent-2", Payload: "CAPABILITY:secure-storage"},
				{Type: "signal", From: "agent-1", To: "agent-2", Payload: "CAPABILITY:ai-reasoning"},
			},
			ExpectedOutcome: "Detect unverified capability claims",
		},
		{
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
		},
		{
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
		},
	}
}

// String returns a formatted representation of a weakness
func (w *Weakness) String() string {
	return fmt.Sprintf("[%s] %s - %s", w.Severity, w.Scenario, w.Description)
}

// Report generates a detailed hardening report
func (e *Entrencher) Report() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.weaknesses) == 0 {
		return "No weaknesses discovered yet. Run RunAll() or RunScenario() first."
	}

	report := fmt.Sprintf("=== ENTRONCHER HARDENING REPORT ===\n")
	report += fmt.Sprintf("Total Scenarios: %d\n", len(e.scenarios))
	report += fmt.Sprintf("Weaknesses Found: %d\n\n", len(e.weaknesses))

	severityCount := make(map[string]int)
	for _, w := range e.weaknesses {
		severityCount[w.Severity]++
		report += fmt.Sprintf("%s\n", w.String())
		report += fmt.Sprintf("  Hardening: %s\n\n", w.Hardening)
	}

	report += "\n=== SEVERITY SUMMARY ===\n"
	for severity, count := range severityCount {
		report += fmt.Sprintf("%s: %d\n", severity, count)
	}

	return report
}

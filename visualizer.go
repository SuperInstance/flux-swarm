package main

import (
	"fmt"
	"strings"
)

// PrintSwarmStatus shows all agents, their roles, and trust scores
func PrintSwarmStatus(sc *SwarmCoordinator) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	fmt.Println("=== SWARM STATUS ===")
	fmt.Printf("Total Agents: %d\n", len(sc.Agents))
	fmt.Println("ID\t\t\t\tRole\t\tTrust")
	fmt.Println(strings.Repeat("-", 80))
	
	for id, agent := range sc.Agents {
		shortID := id[:8]
		fmt.Printf("%s...\t%s\t\t%.2f\n", shortID, agent.Role, agent.Trust)
	}
	fmt.Println()
}

// PrintMessageFlow shows recent A2A messages as arrows
func PrintMessageFlow(sc *SwarmCoordinator, maxMessages int) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	fmt.Println("=== MESSAGE FLOW ===")
	
	if len(sc.MessageLog) == 0 {
		fmt.Println("No messages yet.")
		return
	}
	
	start := 0
	if len(sc.MessageLog) > maxMessages {
		start = len(sc.MessageLog) - maxMessages
	}
	
	for i := start; i < len(sc.MessageLog); i++ {
		msg := sc.MessageLog[i]
		senderShort := msg.SenderID[:4]
		receiverShort := msg.ReceiverID[:4]
		fmt.Printf("%s --[%s:%s]--> %s\n", 
			senderShort, 
			msg.Type, 
			msg.Payload[:min(10, len(msg.Payload))], 
			receiverShort)
	}
	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PrintFormation shows agent topology as ASCII graph
func PrintFormation(sc *SwarmCoordinator) {
	formation := sc.GetFormation()
	
	fmt.Println("=== AGENT FORMATION ===")
	for agentID, connections := range formation {
		shortID := agentID[:4]
		connStrs := make([]string, len(connections))
		for i, conn := range connections {
			connStrs[i] = conn[:4]
		}
		fmt.Printf("%s -> [%s]\n", shortID, strings.Join(connStrs, ", "))
	}
	fmt.Println()
}

// PrintTrustMatrix shows trust scores between agents as grid
func PrintTrustMatrix(sc *SwarmCoordinator) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	fmt.Println("=== TRUST MATRIX ===")
	
	ids := make([]string, 0, len(sc.Agents))
	for id := range sc.Agents {
		ids = append(ids, id)
	}
	
	// Print header
	fmt.Print("     ")
	for _, id := range ids {
		fmt.Printf("%4s", id[:2])
	}
	fmt.Println()
	
	// Print rows
	for _, id1 := range ids {
		fmt.Printf("%2s: ", id1[:2])
		for _, id2 := range ids {
			if id1 == id2 {
				fmt.Print("  - ")
			} else {
				trust := sc.TrustMatrix[id1][id2]
				fmt.Printf("%4.1f", trust)
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

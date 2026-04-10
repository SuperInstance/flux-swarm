# flux-swarm

FLUX swarm coordinator in Go — distributed agent coordination with A2A messaging, trust scoring, and ASCII visualization.

## Features
- FLUX bytecode VM interpreter
- Agent registry with roles (worker, scout, coordinator, specialist)
- A2A protocol: TELL, ASK, DELEGATE, BROADCAST
- Trust matrix between agents
- ASCII swarm visualization

## Tests (5/5 passing)

```bash
go test ./... -v
```
